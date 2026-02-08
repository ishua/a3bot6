package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

type Config struct {
	Synology struct {
		URL      string            `yaml:"url"`
		Username string            `yaml:"username"`
		Password string            `yaml:"password"`
		Paths    map[string]string `yaml:"paths"`
	} `yaml:"synology"`
	Debug bool `yaml:"debug"`
}

type SynoResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *SynoError      `json:"error,omitempty"`
}

type SynoError struct {
	Code int `json:"code"`
}

func main() {
	// Загрузка конфига
	var cfg Config
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files: []string{"conf/config.yaml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})
	if err := loader.Load(); err != nil {
		log.Fatal("Config load error:", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	baseURL := cfg.Synology.URL

	fmt.Println("=== Synology Download Station Test ===")
	fmt.Println("URL:", baseURL)
	fmt.Println()

	// 1. Login
	fmt.Println("1. Login...")
	sid, err := login(client, baseURL, cfg.Synology.Username, cfg.Synology.Password)
	if err != nil {
		log.Fatal("Login failed:", err)
	}
	fmt.Println("✓ Login successful, SID:", sid)
	fmt.Println()

	// 1.5 Get info about Download Station
	fmt.Println("1.5. Get Download Station info...")
	err = getInfo(client, baseURL, sid)
	if err != nil {
		log.Fatal("Get info failed:", err)
	}
	fmt.Println()

	// 2. List tasks
	fmt.Println("2. List existing tasks...")
	tasks, err := listTasks(client, baseURL, sid)
	if err != nil {
		log.Fatal("List tasks failed:", err)
	}
	fmt.Printf("✓ Found %d tasks\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("  - %s: %s\n", task["id"], task["title"])
	}
	fmt.Println()

	// 3. Create task (test URL - небольшой торрент)
	fmt.Println("3. Create test task...")
	testURL := "magnet:?xt=urn:btih:dd8255ecdc7ca55fb0bbf81323d87062db1f6d1c" // Big Buck Bunny
	destination := "sdata/tmp"
	taskID, err := createTask(client, baseURL, sid, testURL, destination)
	if err != nil {
		log.Fatal("Create task failed:", err)
	}
	fmt.Println("✓ Task created, ID:", taskID)
	fmt.Println()

	// Ждем немного
	fmt.Println("Waiting 3 seconds...")
	time.Sleep(3 * time.Second)

	// 4. List tasks again
	fmt.Println("4. List tasks after creation...")
	tasks, err = listTasks(client, baseURL, sid)
	if err != nil {
		log.Fatal("List tasks failed:", err)
	}
	fmt.Printf("✓ Found %d tasks\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("  - %s: %s (status: %s)\n", task["id"], task["title"], task["status"])
	}
	fmt.Println()

	// 5. Delete created task
	if taskID != "" {
		fmt.Println("5. Delete test task...")
		err = deleteTask(client, baseURL, sid, taskID)
		if err != nil {
			log.Fatal("Delete task failed:", err)
		}
		fmt.Println("✓ Task deleted")
		fmt.Println()
	}

	// 6. Logout
	fmt.Println("6. Logout...")
	err = logout(client, baseURL, sid)
	if err != nil {
		log.Fatal("Logout failed:", err)
	}
	fmt.Println("✓ Logout successful")
}

func login(client *http.Client, baseURL, username, password string) (string, error) {
	apiURL := fmt.Sprintf("%s/webapi/auth.cgi", baseURL)
	params := url.Values{
		"api":     {"SYNO.API.Auth"},
		"version": {"7"},
		"method":  {"login"},
		"account": {username},
		"passwd":  {password},
		"session": {"DownloadStation"},
		"format":  {"sid"},
	}

	resp, err := client.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result SynoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("login failed, error code: %d", result.Error.Code)
	}

	var data struct {
		SID string `json:"sid"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return "", err
	}

	return data.SID, nil
}

func logout(client *http.Client, baseURL, sid string) error {
	apiURL := fmt.Sprintf("%s/webapi/auth.cgi", baseURL)
	params := url.Values{
		"api":     {"SYNO.API.Auth"},
		"version": {"7"},
		"method":  {"logout"},
		"session": {"DownloadStation"},
		"_sid":    {sid},
	}

	resp, err := client.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result SynoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("logout failed")
	}

	return nil
}

func getInfo(client *http.Client, baseURL, sid string) error {
	apiURL := fmt.Sprintf("%s/webapi/DownloadStation/info.cgi", baseURL)
	params := url.Values{
		"api":     {"SYNO.DownloadStation.Info"},
		"version": {"2"},
		"method":  {"getinfo"},
		"_sid":    {sid},
	}

	resp, err := client.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Info response:", string(body))

	return nil
}

func listTasks(client *http.Client, baseURL, sid string) ([]map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/webapi/DownloadStation/task.cgi", baseURL)
	params := url.Values{
		"api":        {"SYNO.DownloadStation.Task"},
		"version":    {"3"},
		"method":     {"list"},
		"additional": {"detail,transfer"},
		"_sid":       {sid},
	}

	resp, err := client.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result SynoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("list failed, error code: %d", result.Error.Code)
	}

	var data struct {
		Tasks []map[string]interface{} `json:"tasks"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return nil, err
	}

	return data.Tasks, nil
}

func createTask(client *http.Client, baseURL, sid, taskURL, destination string) (string, error) {
	apiURL := fmt.Sprintf("%s/webapi/DownloadStation/task.cgi", baseURL)
	params := url.Values{
		"api":         {"SYNO.DownloadStation.Task"},
		"version":     {"3"},
		"method":      {"create"},
		"uri":         {taskURL},
		"destination": {destination},
		"_sid":        {sid},
	}

	resp, err := client.Post(apiURL, "application/x-www-form-urlencoded", bytes.NewBufferString(params.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Create response:", string(body))

	var result SynoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("create failed, error code: %d", result.Error.Code)
	}

	tasks, err := listTasks(client, baseURL, sid)
	if err != nil {
		return "", err
	}

	if len(tasks) > 0 {
		if id, ok := tasks[0]["id"].(string); ok {
			return id, nil
		}
	}

	return "", nil
}

func deleteTask(client *http.Client, baseURL, sid, taskID string) error {
	apiURL := fmt.Sprintf("%s/webapi/DownloadStation/task.cgi", baseURL)
	params := url.Values{
		"api":     {"SYNO.DownloadStation.Task"},
		"version": {"3"},
		"method":  {"delete"},
		"id":      {taskID},
		"_sid":    {sid},
	}

	resp, err := client.Post(apiURL, "application/x-www-form-urlencoded", bytes.NewBufferString(params.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result SynoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("delete failed, error code: %d", result.Error.Code)
	}

	return nil
}
