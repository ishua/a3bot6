package main

import (
	"fmt"
	"log"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/ishua/a3bot6/mcore/pkg/schema"
	"github.com/ishua/a3bot6/synoc/internal/synology"
	"github.com/ishua/a3bot6/synoc/internal/worker"
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

	fmt.Println("=== Synology Worker Test ===")
	fmt.Println("URL:", cfg.Synology.URL)
	fmt.Println()

	client := synology.NewClient(cfg.Synology.URL)

	fmt.Println("1. Login...")
	err := client.Login(cfg.Synology.Username, cfg.Synology.Password)
	if err != nil {
		log.Fatal("Login failed:", err)
	}
	fmt.Println("✓ Login successful")
	fmt.Println()

	w := worker.New(client, cfg.Synology.Paths)

	fmt.Println("2. DoTask (add)...")
	addTask := schema.Task{
		Id:   1,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command:    schema.SynoTaskCmdAdd,
				Category:   schema.SynoCategoryOther,
				TorrentUrl: "https://releases.ubuntu.com/22.04/ubuntu-22.04.3-live-server-amd64.iso",
			},
		},
	}
	addReport := w.DoTask(addTask)
	fmt.Printf("✓ Result: Status=%d, Text=%s\n", addReport.Status, addReport.TextMsg)
	fmt.Println()

	fmt.Println("3. DoTask (list)...")
	listTask := schema.Task{
		Id:   2,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command: schema.SynoTaskCmdList,
			},
		},
	}
	listReport := w.DoTask(listTask)
	fmt.Printf("✓ Result: Status=%d, Text=%s\n", listReport.Status, listReport.TextMsg)
	fmt.Println()

	fmt.Println("4. DoTask (delete)...")
	deleteTask := schema.Task{
		Id:   3,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command: schema.SynoTaskCmdDelete,
				TaskId:  "dbid_17",
			},
		},
	}
	deleteReport := w.DoTask(deleteTask)
	fmt.Printf("✓ Result: Status=%d, Text=%s\n", deleteReport.Status, deleteReport.TextMsg)
	fmt.Println()

	fmt.Println("5. DoTask (list after delete)...")
	listAfterDeleteReport := w.DoTask(listTask)
	fmt.Printf("✓ Result: Status=%d, Text=%s\n", listAfterDeleteReport.Status, listAfterDeleteReport.TextMsg)
	fmt.Println()

	fmt.Println("6. Cleanup - delete created task dbid_18...")
	cleanupTask := schema.Task{
		Id:   4,
		Type: schema.TaskTypeSyno,
		TaskData: schema.TaskData{
			Syno: schema.TaskSyno{
				Command: schema.SynoTaskCmdDelete,
				TaskId:  "dbid_18",
			},
		},
	}
	cleanupReport := w.DoTask(cleanupTask)
	fmt.Printf("✓ Result: Status=%d, Text=%s\n", cleanupReport.Status, cleanupReport.TextMsg)
	fmt.Println()

	fmt.Println("7. Logout...")
	err = client.Logout()
	if err != nil {
		log.Fatal("Logout failed:", err)
	}
	fmt.Println("✓ Logout successful")
}
