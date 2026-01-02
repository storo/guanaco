// Package i18n provides internationalization support for Guanaco.
package i18n

import (
	"os"
	"strings"
	"sync"
)

// translations holds the current translation map.
var (
	translations = make(map[string]string)
	mu           sync.RWMutex
	currentLang  string
)

// Init initializes the i18n system.
// It detects the current locale from environment variables.
func Init(localeDir string) {
	// Detect locale from environment
	lang := os.Getenv("LANGUAGE")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}
	if lang == "" {
		lang = os.Getenv("LANG")
	}

	// Extract language code (e.g., "es_ES.UTF-8" -> "es")
	if idx := strings.Index(lang, "_"); idx > 0 {
		lang = lang[:idx]
	}
	if idx := strings.Index(lang, "."); idx > 0 {
		lang = lang[:idx]
	}

	SetLanguage(lang)
}

// SetLanguage sets the current language.
func SetLanguage(lang string) {
	mu.Lock()
	defer mu.Unlock()

	currentLang = lang

	// Load translations for the language
	translations = make(map[string]string)

	switch lang {
	case "es":
		loadSpanish()
	case "en":
		// English is the default, no translation needed
	default:
		// Default to source strings (English)
	}
}

// T translates a string.
// If no translation is found, returns the original string.
func T(msgid string) string {
	mu.RLock()
	defer mu.RUnlock()

	if trans, ok := translations[msgid]; ok {
		return trans
	}
	return msgid
}

// N translates a string with plural forms.
func N(singular, plural string, n uint) string {
	if n == 1 {
		return T(singular)
	}
	return T(plural)
}

// CurrentLanguage returns the current language code.
func CurrentLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// loadSpanish loads Spanish translations.
func loadSpanish() {
	translations["New Chat"] = "Nueva conversación"
	translations["Send message (Ctrl+Enter)"] = "Enviar mensaje (Ctrl+Enter)"
	translations["Attach file"] = "Adjuntar archivo"
	translations["Main Menu"] = "Menú principal"
	translations["Chats"] = "Conversaciones"
	translations["Chat"] = "Conversación"
	translations["Start Ollama"] = "Iniciar Ollama"
	translations["Retry Connection"] = "Reintentar conexión"
	translations["Ollama Not Detected"] = "Ollama no detectado"
	translations["Guanaco requires Ollama to be running.\nClick the button below to start Ollama."] = "Guanaco requiere que Ollama esté ejecutándose.\nHaz clic en el botón de abajo para iniciar Ollama."
	translations["Starting Ollama..."] = "Iniciando Ollama..."
	translations["Ollama started successfully!"] = "¡Ollama iniciado correctamente!"
	translations["Failed to start Ollama: "] = "Error al iniciar Ollama: "
	translations["Failed to load models: "] = "Error al cargar modelos: "
	translations["Loaded %d models"] = "Cargados %d modelos"
	translations["No models found. Run: ollama pull llama3.2"] = "No se encontraron modelos. Ejecuta: ollama pull llama3.2"
	translations["Error: "] = "Error: "
	translations["Select Document"] = "Seleccionar documento"
	translations["Supported Documents"] = "Documentos soportados"
	translations["Text Files"] = "Archivos de texto"
	translations["PDF Documents"] = "Documentos PDF"
	translations["Open"] = "Abrir"
	translations["Cancel"] = "Cancelar"
	translations["Remove attachment"] = "Eliminar adjunto"
	translations["Downloading model %s..."] = "Descargando modelo %s..."
	translations["please enter a model name (e.g., llama3.2)"] = "por favor ingresa un nombre de modelo (ej., llama3.2)"
}
