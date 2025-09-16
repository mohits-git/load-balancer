package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Protocol            string   `json:"protocol"`
	Port                int      `json:"port"`
	Algorithm           string   `json:"algorithm"`
	HealthCheckInterval int      `json:"healthCheckInterval"`
	RetryLimit          int      `json:"retryLimit"`
	Servers             []Server `json:"servers"`
}

type Server struct {
	Addr                    string `json:"addr"`
	HealthCheckHTTPEndpoint string `json:"healthCheckHTTPEndpoint"`
	Weight                  int    `json:"weight"`
}

func LoadConfig() *Config {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error while getting working directory")
	}
	file, err := os.Open(filepath.Join(wd, "config.json"))
	if err != nil {
		log.Fatalf("Please provide load balancer config")
	}
	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		log.Fatalf("Error while reading and parsing config.json")
	}
	return &config
}
