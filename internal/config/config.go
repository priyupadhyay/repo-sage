package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile represents an LLM endpoint configuration
type Profile struct {
	APIBase string `yaml:"api_base"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
}

// Config represents the main configuration structure
type Config struct {
	Profiles       map[string]Profile `yaml:"profiles"`
	DefaultProfile string             `yaml:"default_profile"`
}

const (
	configDir  = ".repo-sage"
	configFile = "config.yaml"
)

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

// LoadConfig loads the configuration from disk
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{
				Profiles: make(map[string]Profile),
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Profiles == nil {
		config.Profiles = make(map[string]Profile)
	}

	return &config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddProfile adds or updates a profile in the configuration
func (c *Config) AddProfile(name string, profile Profile) {
	c.Profiles[name] = profile
}

// GetProfile retrieves a profile by name
func (c *Config) GetProfile(name string) (Profile, bool) {
	profile, exists := c.Profiles[name]
	return profile, exists
}

// SetDefaultProfile sets the default profile
func (c *Config) SetDefaultProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return fmt.Errorf("profile %q does not exist", name)
	}
	c.DefaultProfile = name
	return nil
}

// GetDefaultProfile returns the default profile and its name
func (c *Config) GetDefaultProfile() (Profile, string, error) {
	if c.DefaultProfile == "" {
		return Profile{}, "", fmt.Errorf("no default profile configured")
	}

	profile, exists := c.Profiles[c.DefaultProfile]
	if !exists {
		return Profile{}, "", fmt.Errorf("default profile %q not found", c.DefaultProfile)
	}

	return profile, c.DefaultProfile, nil
}
