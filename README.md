# Guanaco

A modern GTK4/Libadwaita desktop application for chatting with local AI models powered by Ollama.

## Features

- Stream responses in real-time as the AI generates them
- Beautiful markdown rendering with code highlighting
- Drag and drop documents (PDF, TXT, Markdown) for context
- Persistent chat history stored locally
- Auto-download models when they are not installed
- Native GTK4/Libadwaita interface following GNOME HIG

## Requirements

- Linux with GTK4 and Libadwaita
- [Ollama](https://ollama.ai/) running locally

## Installation

### Flathub (Recommended)

```bash
flatpak install flathub com.github.storo.Guanaco
```

### Snap Store

```bash
sudo snap install guanaco
```

### Debian/Ubuntu (.deb)

Download the latest `.deb` package from [GitHub Releases](https://github.com/storo/guanaco/releases) and install:

```bash
sudo dpkg -i guanaco_*_amd64.deb
sudo apt-get install -f  # Install dependencies if needed
```

### Build from Source

```bash
# Install dependencies (Fedora)
sudo dnf install gtk4-devel libadwaita-devel golang

# Install dependencies (Ubuntu/Debian)
sudo apt install libgtk-4-dev libadwaita-1-dev golang

# Build
make build

# Run
./guanaco
```

## Configuration

Guanaco connects to Ollama at `http://localhost:11434` by default. You can configure this in the settings dialog.

## License

MIT License - see [LICENSE](LICENSE) for details.
