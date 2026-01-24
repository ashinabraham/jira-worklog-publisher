package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"jira-calendar/internal/models"
)

// Load loads the Jira configuration from the user's home directory
func Load() models.JiraConfig {
	homeDir, _ := os.UserHomeDir()
	configPath := fmt.Sprintf("%s/.jira-calendar-config.json", homeDir)

	log.Printf("[CONFIG] Loading config from: %s", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("[CONFIG] No existing config found or error reading: %v", err)
		return models.JiraConfig{}
	}

	var cfg models.JiraConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("[CONFIG] Error parsing config: %v", err)
		return models.JiraConfig{}
	}

	log.Printf("[CONFIG] Config loaded successfully - BaseURL: %s, Username: %s", cfg.BaseURL, cfg.Username)
	return cfg
}

// Save saves the Jira configuration to the user's home directory
func Save(cfg models.JiraConfig) error {
	homeDir, _ := os.UserHomeDir()
	configPath := fmt.Sprintf("%s/.jira-calendar-config.json", homeDir)

	log.Printf("[CONFIG] Saving config to: %s", configPath)
	log.Printf("[CONFIG] BaseURL: %s, Username: %s", cfg.BaseURL, cfg.Username)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("[CONFIG] Error marshaling config: %v", err)
		return err
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		log.Printf("[CONFIG] Error writing config file: %v", err)
		return err
	}

	log.Printf("[CONFIG] Config saved successfully")
	return nil
}
