package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig holds the application-wide settings.
type AppConfig struct {
	DefaultModel       string `json:"default_model"`
	ResponseLanguage   string `json:"response_language"` // "auto", "en", "es", etc.
	GlobalSystemPrompt string `json:"global_system_prompt"`
	SidebarVisible     bool   `json:"sidebar_visible"`
}

// DefaultConfig returns a new AppConfig with default values.
func DefaultConfig() *AppConfig {
	return &AppConfig{
		DefaultModel:       "",
		ResponseLanguage:   "auto",
		GlobalSystemPrompt: "",
		SidebarVisible:     true,
	}
}

// GetConfigFilePath returns the path to the settings file.
func GetConfigFilePath() string {
	return filepath.Join(GetConfigDir(), "settings.json")
}

// LoadConfig loads the application configuration from disk.
// Returns default config if file doesn't exist.
func LoadConfig() (*AppConfig, error) {
	configPath := GetConfigFilePath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, err
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// Save writes the configuration to disk.
func (c *AppConfig) Save() error {
	// Ensure config directory exists
	if err := EnsureDirectories(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(GetConfigFilePath(), data, 0644)
}

// LanguageInstruction returns the system prompt instruction for the configured language.
func (c *AppConfig) LanguageInstruction() string {
	switch c.ResponseLanguage {
	case "en":
		return "Always respond in English."
	case "es":
		return "Siempre responde en español."
	case "pt":
		return "Sempre responda em português."
	case "fr":
		return "Réponds toujours en français."
	case "de":
		return "Antworte immer auf Deutsch."
	default:
		return ""
	}
}

// GetEffectiveSystemPrompt returns the system prompt with language instruction appended.
func (c *AppConfig) GetEffectiveSystemPrompt(chatPrompt string) string {
	// Chat-specific prompt has priority
	prompt := chatPrompt
	if prompt == "" {
		prompt = c.GlobalSystemPrompt
	}

	// Append language instruction if configured
	langInstruction := c.LanguageInstruction()
	if langInstruction != "" {
		if prompt != "" {
			prompt = prompt + "\n\n" + langInstruction
		} else {
			prompt = langInstruction
		}
	}

	return prompt
}
