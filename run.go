package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"agent/authorization"
	"agent/docker"
	"agent/user"
)

type AgentConfig struct {
	Port        string `json:"port"`
	ProjectPath string `json:"project_path"`
	Docker      string `json:"docker"`
}

func loadConfig() AgentConfig {
	configPath := filepath.Join(filepath.Dir(os.Args[0]), "config.json")
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to open config.json: %v", err)
	}
	defer file.Close()

	var cfg AgentConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.ProjectPath == "" {
		cfg.ProjectPath = "/raweb/web/panel/"
	}
	return cfg
}

func main() {
	cfg := loadConfig()
	authorization.InitAuthWithPath(cfg.ProjectPath)
	docker.InitDocker(cfg.Docker)

	mux := http.NewServeMux()
	mux.Handle("/system/user/create", authorization.AuthMiddleware(http.HandlerFunc(user.CreateUserHandler)))
	// mux.Handle("/system/user/delete", authorization.AuthMiddleware(http.HandlerFunc(user.DeleteUserHandler)))
	mux.Handle("/app/list_all", authorization.AuthMiddleware(http.HandlerFunc(docker.ListContainersHandler)))
	mux.Handle("/app/delete", authorization.AuthMiddleware(http.HandlerFunc(docker.DeleteContainerHandler)))
	log.Printf("Agent running on :%s (project path: %s)\n", cfg.Port, cfg.ProjectPath)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
