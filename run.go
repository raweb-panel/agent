package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

func loadConfig(configPath string) AgentConfig {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", configPath)
	}

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to open config file %s: %v", configPath, err)
	}
	defer file.Close()

	var cfg AgentConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("Failed to parse config file %s: %v", configPath, err)
	}

	// Set defaults if not provided
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.ProjectPath == "" {
		cfg.ProjectPath = "/raweb/web/panel/"
	}

	return cfg
}

func printUsage() {
	fmt.Printf("Usage: %s --config=<full-path-to-config.json>\n", os.Args[0])
	fmt.Printf("\nOptions:\n")
	fmt.Printf("  --config    Full path to the configuration JSON file (required)\n")
	fmt.Printf("  --help      Show this help message\n")
	fmt.Printf("\nExample:\n")
	fmt.Printf("  %s --config=/raweb/apps/agent/config.json\n", os.Args[0])
}

func main() {
	var configPath string
	var showHelp bool

	flag.StringVar(&configPath, "config", "", "Full path to the configuration JSON file")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	if configPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --config flag is required\n\n")
		printUsage()
		os.Exit(1)
	}

	// Validate that the config path is absolute
	if !filepath.IsAbs(configPath) {
		fmt.Fprintf(os.Stderr, "Error: Config path must be absolute (full path): %s\n\n", configPath)
		printUsage()
		os.Exit(1)
	}

	cfg := loadConfig(configPath)
	authorization.InitAuthWithPath(cfg.ProjectPath)
	docker.InitDocker(cfg.Docker)

	mux := http.NewServeMux()
	mux.Handle("/system/user/create", authorization.AuthMiddleware(http.HandlerFunc(user.CreateUserHandler)))

	mux.Handle("/container/list", authorization.AuthMiddleware(http.HandlerFunc(docker.ListContainersHandler)))
	mux.Handle("/container/delete", authorization.AuthMiddleware(http.HandlerFunc(docker.DeleteContainerHandler)))
	mux.Handle("/container/stop", authorization.AuthMiddleware(http.HandlerFunc(docker.StopContainerHandler)))
	mux.Handle("/container/start", authorization.AuthMiddleware(http.HandlerFunc(docker.StartContainerHandler)))
	mux.Handle("/container/kill", authorization.AuthMiddleware(http.HandlerFunc(docker.KillContainerHandler)))
	mux.Handle("/container/create", authorization.AuthMiddleware(http.HandlerFunc(docker.CreateContainerHandler)))
	mux.Handle("/container/get_by_id", authorization.AuthMiddleware(http.HandlerFunc(docker.GetContainerByIDHandler)))
	mux.Handle("/container/get_by_name", authorization.AuthMiddleware(http.HandlerFunc(docker.GetContainerByNameHandler)))
	mux.Handle("/container/stats_by_name", authorization.AuthMiddleware(http.HandlerFunc(docker.GetContainerStatsByNameHandler)))

	mux.Handle("/network/create", authorization.AuthMiddleware(http.HandlerFunc(docker.CreateNetworkHandler)))
	mux.Handle("/network/list", authorization.AuthMiddleware(http.HandlerFunc(docker.ListNetworksHandler)))
	mux.Handle("/network/delete", authorization.AuthMiddleware(http.HandlerFunc(docker.DeleteNetworkHandler)))

	mux.Handle("/image/list", authorization.AuthMiddleware(http.HandlerFunc(docker.ListImagesHandler)))
	mux.Handle("/image/delete", authorization.AuthMiddleware(http.HandlerFunc(docker.DeleteImageHandler)))

	log.Printf("Agent starting with config: %s", configPath)
	log.Printf("Agent running on :%s (project path: %s)\n", cfg.Port, cfg.ProjectPath)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
