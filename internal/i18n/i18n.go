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
	// General
	translations["New Chat"] = "Nueva conversación"
	translations["Send message (Ctrl+Enter)"] = "Enviar mensaje (Ctrl+Enter)"
	translations["Attach file"] = "Adjuntar archivo"
	translations["Main Menu"] = "Menú principal"
	translations["Chats"] = "Conversaciones"
	translations["Chat"] = "Conversación"
	translations["Error: "] = "Error: "
	translations["Open"] = "Abrir"
	translations["Cancel"] = "Cancelar"
	translations["Save"] = "Guardar"
	translations["Settings"] = "Configuración"
	translations["Loading..."] = "Cargando..."

	// Ollama status
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
	translations["No models found. Use the download button to pull a model."] = "No se encontraron modelos. Usa el botón de descarga para obtener uno."

	// Header bar
	translations["Toggle Sidebar"] = "Mostrar/ocultar barra lateral"
	translations["Download Model"] = "Descargar modelo"
	translations["Chat Settings"] = "Configuración del chat"

	// Sidebar
	translations["Delete chat"] = "Eliminar conversación"
	translations["Delete Chat?"] = "¿Eliminar conversación?"
	translations["This conversation will be permanently deleted. This action cannot be undone."] = "Esta conversación se eliminará permanentemente. Esta acción no se puede deshacer."
	translations["Delete"] = "Eliminar"
	translations["No conversations yet"] = "Aún no hay conversaciones"
	translations["Start a new chat to begin"] = "Inicia una nueva conversación para comenzar"

	// Input area
	translations["Select model"] = "Seleccionar modelo"
	translations["Stop generation"] = "Detener generación"
	translations["Type a message..."] = "Escribe un mensaje..."

	// Chat view - Welcome screen
	translations["Good morning"] = "Buenos días"
	translations["Good afternoon"] = "Buenas tardes"
	translations["Good evening"] = "Buenas noches"
	translations["How can I help you today?"] = "¿Cómo puedo ayudarte hoy?"
	translations["Explain"] = "Explícame"
	translations["Write"] = "Escribe"
	translations["Summarize"] = "Resume"
	translations["Translate"] = "Traduce"
	translations["a concept"] = "un concepto"
	translations["code for me"] = "código para mí"
	translations["this article"] = "este artículo"
	translations["to English"] = "al español"

	// File attachments
	translations["Select Document"] = "Seleccionar documento"
	translations["Supported Documents"] = "Documentos soportados"
	translations["Text Files"] = "Archivos de texto"
	translations["PDF Documents"] = "Documentos PDF"
	translations["All Supported Files"] = "Todos los archivos soportados"
	translations["Images"] = "Imágenes"
	translations["Remove attachment"] = "Eliminar adjunto"
	translations["unsupported file type: %s"] = "tipo de archivo no soportado: %s"
	translations["file too large: %s (max %dMB)"] = "archivo demasiado grande: %s (máx %dMB)"
	translations["failed to process %s: %v"] = "error al procesar %s: %v"
	translations["%s (%d chars)"] = "%s (%d caracteres)"

	// Model dialog
	translations["Download Model"] = "Descargar Modelo"
	translations["Available Models:"] = "Modelos disponibles:"
	translations["Or enter custom model:"] = "O ingresa un modelo personalizado:"
	translations["Model name..."] = "Nombre del modelo..."
	translations["Download"] = "Descargar"
	translations["Downloading..."] = "Descargando..."
	translations["Starting download..."] = "Iniciando descarga..."
	translations["Download cancelled"] = "Descarga cancelada"
	translations["Download complete!"] = "¡Descarga completa!"
	translations["Downloading model %s..."] = "Descargando modelo %s..."
	translations["please enter a model name (e.g., llama3.2)"] = "por favor ingresa un nombre de modelo (ej., llama3.2)"

	// System prompt dialog
	translations["System Prompt"] = "Prompt del sistema"
	translations["Set instructions that define how the AI should behave in this chat."] = "Define instrucciones sobre cómo debe comportarse la IA en esta conversación."

	// Settings dialog
	translations["Default Model:"] = "Modelo predeterminado:"
	translations["Response Language:"] = "Idioma de respuesta:"
	translations["Global System Prompt:"] = "Prompt global del sistema:"
	translations["Applied to all new chats (chat-specific prompts take priority)"] = "Se aplica a todas las conversaciones nuevas (los prompts específicos tienen prioridad)"
	translations["(None - use first available)"] = "(Ninguno - usar el primero disponible)"

	// Toast messages
	translations["Model %s downloaded!"] = "¡Modelo %s descargado!"
	translations["System prompt saved"] = "Prompt del sistema guardado"
	translations["Settings saved"] = "Configuración guardada"

	// User-friendly error messages
	translations["Could not connect to Ollama. Please check if it's running."] = "No se pudo conectar a Ollama. Verifica que esté en ejecución."
	translations["Failed to load the list of models. Please try again."] = "Error al cargar la lista de modelos. Intenta de nuevo."
	translations["Could not start Ollama. Please start it manually."] = "No se pudo iniciar Ollama. Por favor, inícialo manualmente."
	translations["Model download failed. Please check your connection."] = "Error al descargar el modelo. Verifica tu conexión."
	translations["Response timed out. The model took too long to respond."] = "Tiempo de espera agotado. El modelo tardó demasiado en responder."

	// Copy button
	translations["Copy code"] = "Copiar código"
	translations["Copied!"] = "¡Copiado!"

	// Model download
	translations["Failed to download model: "] = "Error al descargar modelo: "
}
