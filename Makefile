# Guanaco Makefile

APP_ID = com.github.storo.Guanaco
VERSION = 0.1.0
BINARY = guanaco

# Go build flags
LDFLAGS = -s -w -X main.version=$(VERSION)
GOFLAGS = -trimpath

.PHONY: all build clean test lint install uninstall flatpak deb

all: build

# Build the binary
build:
	go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/guanaco

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Lint the code
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Vet code
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -f coverage.out coverage.html
	rm -rf .flatpak-builder

# Install locally
install: build
	install -Dm755 $(BINARY) $(DESTDIR)/usr/bin/$(BINARY)
	install -Dm644 assets/icons/$(APP_ID).svg $(DESTDIR)/usr/share/icons/hicolor/scalable/apps/$(APP_ID).svg
	install -Dm644 assets/$(APP_ID).desktop $(DESTDIR)/usr/share/applications/$(APP_ID).desktop
	install -Dm644 assets/$(APP_ID).metainfo.xml $(DESTDIR)/usr/share/metainfo/$(APP_ID).metainfo.xml

# Uninstall
uninstall:
	rm -f $(DESTDIR)/usr/bin/$(BINARY)
	rm -f $(DESTDIR)/usr/share/icons/hicolor/scalable/apps/$(APP_ID).svg
	rm -f $(DESTDIR)/usr/share/applications/$(APP_ID).desktop
	rm -f $(DESTDIR)/usr/share/metainfo/$(APP_ID).metainfo.xml

# Build Flatpak
flatpak:
	flatpak-builder --force-clean --user --install-deps-from=flathub --repo=repo builddir build/flatpak/$(APP_ID).json

# Run the application
run: build
	./$(BINARY)

# Development: run with hot reload (requires entr)
dev:
	find . -name '*.go' | entr -r go run ./cmd/guanaco

# Build .deb package
deb: build
	rm -rf dist/deb
	mkdir -p dist/deb/DEBIAN
	mkdir -p dist/deb/usr/bin
	mkdir -p dist/deb/usr/share/icons/hicolor/scalable/apps
	mkdir -p dist/deb/usr/share/applications
	mkdir -p dist/deb/usr/share/metainfo
	mkdir -p dist/deb/usr/share/doc/guanaco
	cp $(BINARY) dist/deb/usr/bin/
	cp assets/icons/$(APP_ID).svg dist/deb/usr/share/icons/hicolor/scalable/apps/
	cp assets/$(APP_ID).desktop dist/deb/usr/share/applications/
	cp assets/$(APP_ID).metainfo.xml dist/deb/usr/share/metainfo/
	cp LICENSE dist/deb/usr/share/doc/guanaco/copyright
	echo "Package: guanaco" > dist/deb/DEBIAN/control
	echo "Version: $(VERSION)" >> dist/deb/DEBIAN/control
	echo "Section: utils" >> dist/deb/DEBIAN/control
	echo "Priority: optional" >> dist/deb/DEBIAN/control
	echo "Architecture: amd64" >> dist/deb/DEBIAN/control
	echo "Depends: libgtk-4-1, libadwaita-1-0" >> dist/deb/DEBIAN/control
	echo "Maintainer: storo <storo@voidlab.cl>" >> dist/deb/DEBIAN/control
	echo "Description: Chat with local AI models using Ollama" >> dist/deb/DEBIAN/control
	echo " Guanaco is a modern GTK4/Libadwaita desktop application for chatting" >> dist/deb/DEBIAN/control
	echo " with local AI models powered by Ollama." >> dist/deb/DEBIAN/control
	dpkg-deb --build dist/deb dist/guanaco_$(VERSION)_amd64.deb

# Show help
help:
	@echo "Guanaco Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build the binary"
	@echo "  make test         - Run tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make bench        - Run benchmarks"
	@echo "  make lint         - Lint the code"
	@echo "  make fmt          - Format code"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make install      - Install locally"
	@echo "  make uninstall    - Uninstall"
	@echo "  make flatpak      - Build Flatpak"
	@echo "  make deb          - Build .deb package"
	@echo "  make run          - Run the application"
	@echo "  make dev          - Run with hot reload"
