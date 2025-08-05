package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKeys struct {
		OpenWeather string `json:"openweather"`
		// Add other API keys as needed
	} `json:"api_keys"`
	Settings struct {
		DataDirectory string `json:"data_directory"`
		LogLevel      string `json:"log_level"`
		// Add other settings as needed
	} `json:"settings"`
}

var globalConfig *Config

// GetConfig returns the global configuration instance
func GetConfig() *Config {
	if globalConfig == nil {
		config, err := LoadConfig()
		if err != nil {
			// Return default config if load fails
			config = &Config{}
			config.Settings.LogLevel = "info"
		}
		globalConfig = config
	}
	return globalConfig
}

// LoadConfig loads configuration from the config file
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// If config file doesn't exist, create default one
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := &Config{}
		defaultConfig.Settings.LogLevel = "info"
		if err := SaveConfig(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return defaultConfig, nil
	}

	// Read existing config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON with proper formatting
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Update global config
	globalConfig = config
	return nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() (string, error) {
	// Get user's config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Create app-specific config directory
	appConfigDir := filepath.Join(configDir, "FinancialsX")
	return filepath.Join(appConfigDir, "config.json"), nil
}

// UpdateAPIKey updates a specific API key
func UpdateAPIKey(service, key string) error {
	config := GetConfig()
	
	switch service {
	case "openweather":
		config.APIKeys.OpenWeather = key
	default:
		return fmt.Errorf("unknown API service: %s", service)
	}
	
	return SaveConfig(config)
}

// GetAPIKey retrieves a specific API key  
func GetAPIKey(service string) string {
	config := GetConfig()
	
	switch service {
	case "openweather":
		return config.APIKeys.OpenWeather
	default:
		return ""
	}
}