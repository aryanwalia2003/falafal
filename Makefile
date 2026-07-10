BINARY := falafal
VERSION ?= dev
DIST := dist
DRIVE_PKG := github.com/aryanwalia2003/falafal/internal/drive
# Pulled from the environment (CI secrets, or your own shell for local
# testing) — never hardcoded here, never committed.
DRIVE_CLIENT_ID ?=
DRIVE_CLIENT_SECRET ?=

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: build test clean release $(PLATFORMS)

build:
	go build -o $(BINARY) .

test:
	go test ./...

clean:
	rm -rf $(DIST) $(BINARY)

release: clean $(PLATFORMS)
	cd $(DIST) && sha256sum * > checksums.txt

$(PLATFORMS):
	$(eval OS := $(word 1,$(subst /, ,$@)))
	$(eval ARCH := $(word 2,$(subst /, ,$@)))
	$(eval EXT := $(if $(filter windows,$(OS)),.exe,))
	$(eval OUT := $(DIST)/$(BINARY)_$(OS)_$(ARCH))
	mkdir -p $(OUT)
	GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags "-s -w -X main.version=$(VERSION) -X $(DRIVE_PKG).ClientID=$(DRIVE_CLIENT_ID) -X $(DRIVE_PKG).ClientSecret=$(DRIVE_CLIENT_SECRET)" -o $(OUT)/$(BINARY)$(EXT) .
	cp README.md $(OUT)/
	if [ "$(OS)" = "windows" ]; then \
		cd $(DIST) && zip -qr -j $(BINARY)_$(OS)_$(ARCH).zip $(BINARY)_$(OS)_$(ARCH)/; \
	else \
		tar -czf $(DIST)/$(BINARY)_$(OS)_$(ARCH).tar.gz -C $(DIST) $(BINARY)_$(OS)_$(ARCH); \
	fi
	rm -rf $(OUT)
