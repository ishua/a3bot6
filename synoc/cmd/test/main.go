package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/ishua/a3bot6/syno-worker/internal/synology"
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

	fmt.Println("=== Synology Download Station Test ===")
	fmt.Println("URL:", cfg.Synology.URL)
	fmt.Println()

	// Создаем клиент
	client := synology.NewClient(cfg.Synology.URL)

	// 1. Login
	fmt.Println("1. Login...")
	err := client.Login(cfg.Synology.Username, cfg.Synology.Password)
	if err != nil {
		log.Fatal("Login failed:", err)
	}
	fmt.Println("✓ Login successful")
	fmt.Println()

	// 2. Get info
	fmt.Println("2. Get Download Station info...")
	info, err := client.GetInfo()
	if err != nil {
		log.Fatal("Get info failed:", err)
	}
	fmt.Printf("✓ Version: %s, Is Manager: %v\n", info.VersionString, info.IsManager)
	fmt.Println()

	// 3. List tasks
	fmt.Println("3. List existing tasks...")
	tasks, err := client.ListTasks()
	if err != nil {
		log.Fatal("List tasks failed:", err)
	}
	fmt.Printf("✓ Found %d tasks\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("  - %s: %s (status: %s)\n", task.ID, task.Title, task.Status)
	}
	fmt.Println()

	// 4. Create task
	fmt.Println("4. Create test task...")
	testURL := "magnet:?xt=urn:btih:dd8255ecdc7ca55fb0bbf81323d87062db1f6d1c"
	destination := "downloads"
	taskID, err := client.CreateTask(testURL, destination)
	if err != nil {
		log.Fatal("Create task failed:", err)
	}
	fmt.Println("✓ Task created, ID:", taskID)
	fmt.Println()

	// Ждем немного
	fmt.Println("Waiting 3 seconds...")
	time.Sleep(3 * time.Second)

	// 5. List tasks again
	fmt.Println("5. List tasks after creation...")
	tasks, err = client.ListTasks()
	if err != nil {
		log.Fatal("List tasks failed:", err)
	}
	fmt.Printf("✓ Found %d tasks\n", len(tasks))
	for _, task := range tasks {
		fmt.Printf("  - %s: %s (status: %s)\n", task.ID, task.Title, task.Status)
	}
	fmt.Println()

	// 6. Delete created task
	if taskID != "" {
		fmt.Println("6. Delete test task...")
		err = client.DeleteTask(taskID)
		if err != nil {
			log.Fatal("Delete task failed:", err)
		}
		fmt.Println("✓ Task deleted")
		fmt.Println()
	}

	// 7. Logout
	fmt.Println("7. Logout...")
	err = client.Logout()
	if err != nil {
		log.Fatal("Logout failed:", err)
	}
	fmt.Println("✓ Logout successful")
}
