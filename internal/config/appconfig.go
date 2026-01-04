package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// AppConfig holds the application-wide settings.
type AppConfig struct {
	DefaultModel       string `json:"default_model"`
	ResponseLanguage   string `json:"response_language"` // "auto", "en", "es", etc.
	GlobalSystemPrompt string `json:"global_system_prompt"`
	SidebarVisible     bool   `json:"sidebar_visible"`
}

// BaseFormatPrompts contains formatting instructions that are always prepended
// to the system prompt to guide the model toward clean Markdown output.
var BaseFormatPrompts = map[string]string{
	"en": `Use standard Markdown formatting:
- Headings: ## Title (with blank line after)
- Lists: use dash (- item), one per line
- Code: triple backticks with language
- Tables: | Col1 | Col2 | format
No decorative lines (___), no ASCII art.`,

	"es": `Usa formato Markdown estándar:
- Títulos: ## Título (con línea en blanco después)
- Listas: usar guión (- elemento), uno por línea
- Código: triple backticks con lenguaje
- Tablas: formato | Col1 | Col2 |
Sin líneas decorativas (___), sin arte ASCII.`,

	"pt": `Use formatação Markdown padrão:
- Títulos: ## Título (com linha em branco depois)
- Listas: usar hífen (- item), um por linha
- Código: três crases com linguagem
- Tabelas: formato | Col1 | Col2 |
Sem linhas decorativas (___), sem arte ASCII.`,

	"fr": `Utilisez le format Markdown standard:
- Titres: ## Titre (avec ligne vide après)
- Listes: utiliser tiret (- élément), un par ligne
- Code: triple backticks avec langage
- Tableaux: format | Col1 | Col2 |
Pas de lignes décoratives (___), pas d'art ASCII.`,

	"de": `Verwende Standard-Markdown-Format:
- Überschriften: ## Titel (mit Leerzeile danach)
- Listen: Bindestrich verwenden (- Element), eins pro Zeile
- Code: dreifache Backticks mit Sprache
- Tabellen: | Spalte1 | Spalte2 | Format
Keine dekorativen Linien (___), keine ASCII-Kunst.`,
}

// getBaseFormatPrompt returns the base formatting prompt for the given language.
// Falls back to English if the language is not supported.
func getBaseFormatPrompt(lang string) string {
	if prompt, ok := BaseFormatPrompts[lang]; ok {
		return prompt
	}
	return BaseFormatPrompts["en"]
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

	return os.WriteFile(GetConfigFilePath(), data, 0600)
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

// GetEffectiveSystemPrompt returns the system prompt with base formatting
// instructions prepended and language instruction appended.
func (c *AppConfig) GetEffectiveSystemPrompt(chatPrompt string) string {
	// Determine effective language
	effectiveLang := c.ResponseLanguage
	if effectiveLang == "" || effectiveLang == "auto" {
		effectiveLang = "en"
	}

	// Start with base formatting prompt
	parts := []string{getBaseFormatPrompt(effectiveLang)}

	// Add user's custom prompt (chat-specific has priority over global)
	customPrompt := chatPrompt
	if customPrompt == "" {
		customPrompt = c.GlobalSystemPrompt
	}
	if customPrompt != "" {
		parts = append(parts, customPrompt)
	}

	// Add language instruction if configured
	if langInstruction := c.LanguageInstruction(); langInstruction != "" {
		parts = append(parts, langInstruction)
	}

	return strings.Join(parts, "\n\n")
}
