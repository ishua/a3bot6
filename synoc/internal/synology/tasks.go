package synology

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// Task представляет задачу загрузки в Download Station
type Task struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Size   int64  `json:"size"`
}

// Info представляет информацию о Download Station
type Info struct {
	IsManager     bool   `json:"is_manager"`
	Version       int    `json:"version"`
	VersionString string `json:"version_string"`
}

// GetInfo получает информацию о Download Station
func (c *Client) GetInfo() (*Info, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	apiURL := fmt.Sprintf("%s/webapi/DownloadStation/info.cgi", c.baseURL)
	params := url.Values{
		"api":     {"SYNO.DownloadStation.Info"},
		"version": {"2"},
		"method":  {"getinfo"},
		"_sid":    {c.sid},
	}

	resp, err := c.http.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("getinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	var result synoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("getinfo decode failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("getinfo failed, error code: %d", result.Error.Code)
	}

	var info Info
	if err := json.Unmarshal(result.Data, &info); err != nil {
		return nil, fmt.Errorf("getinfo parse failed: %w", err)
	}

	return &info, nil
}

// ListTasks возвращает список всех задач
func (c *Client) ListTasks() ([]Task, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	apiURL := fmt.Sprintf("%s/webapi/DownloadStation/task.cgi", c.baseURL)
	params := url.Values{
		"api":        {"SYNO.DownloadStation.Task"},
		"version":    {"3"},
		"method":     {"list"},
		"additional": {"detail,transfer"},
		"_sid":       {c.sid},
	}

	resp, err := c.http.Get(apiURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("list request failed: %w", err)
	}
	defer resp.Body.Close()

	var result synoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("list decode failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("list failed, error code: %d", result.Error.Code)
	}

	var data struct {
		Tasks []Task `json:"tasks"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return nil, fmt.Errorf("list parse failed: %w", err)
	}

	return data.Tasks, nil
}

// CreateTask создает новую задачу загрузки
func (c *Client) CreateTask(uri, destination string) (string, error) {
	if !c.IsAuthenticated() {
		return "", fmt.Errorf("not authenticated")
	}

	apiURL := fmt.Sprintf("%s/webapi/DownloadStation/task.cgi", c.baseURL)
	params := url.Values{
		"api":         {"SYNO.DownloadStation.Task"},
		"version":     {"3"},
		"method":      {"create"},
		"uri":         {uri},
		"destination": {destination},
		"_sid":        {c.sid},
	}

	resp, err := c.http.Post(apiURL, "application/x-www-form-urlencoded", bytes.NewBufferString(params.Encode()))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result synoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("create decode failed: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("create failed, error code: %d", result.Error.Code)
	}

	// Получаем ID созданной задачи из списка (последняя задача)
	tasks, err := c.ListTasks()
	if err != nil {
		return "", fmt.Errorf("create get task id failed: %w", err)
	}

	if len(tasks) > 0 {
		return tasks[len(tasks)-1].ID, nil
	}

	return "", nil
}

// DeleteTask удаляет задачу по ID
func (c *Client) DeleteTask(taskID string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	apiURL := fmt.Sprintf("%s/webapi/DownloadStation/task.cgi", c.baseURL)
	params := url.Values{
		"api":     {"SYNO.DownloadStation.Task"},
		"version": {"3"},
		"method":  {"delete"},
		"id":      {taskID},
		"_sid":    {c.sid},
	}

	resp, err := c.http.Post(apiURL, "application/x-www-form-urlencoded", bytes.NewBufferString(params.Encode()))
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	var result synoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("delete decode failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("delete failed, error code: %d", result.Error.Code)
	}

	return nil
}
