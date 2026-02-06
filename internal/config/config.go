// Package config loads and saves Jira configuration to a JSON file
// in the user's home directory (~/.jira-worklog-publisher-config.json).
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"jira-worklog-publisher/internal/models"
)

// Load reads the Jira configuration from ~/.jira-worklog-publisher-config.json.
// Returns an empty JiraConfig if the file is missing or invalid.
func Load() models.JiraConfig {
	homeDir, _ := os.UserHomeDir()
	configPath := fmt.Sprintf("%s/.jira-worklog-publisher-config.json", homeDir)

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

// Save writes the Jira configuration to ~/.jira-worklog-publisher-config.json with mode 0600.
func Save(cfg models.JiraConfig) error {
	homeDir, _ := os.UserHomeDir()
	configPath := fmt.Sprintf("%s/.jira-worklog-publisher-config.json", homeDir)

	log.Printf("[CONFIG] Saving config to: %s", configPath)
	log.Printf("[CONFIG] BaseURL: %s, Username: %s", cfg.BaseURL, cfg.Username)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("[CONFIG] Error marshaling config: %v", err)
		return err
	}

	// Config file contains API token; restrict to owner only (0600).
	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		log.Printf("[CONFIG] Error writing config file: %v", err)
		return err
	}

	log.Printf("[CONFIG] Config saved successfully")
	return nil
}
