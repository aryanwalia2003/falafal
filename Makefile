BINARY := falafal
VERSION ?= dev
DIST := dist

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
	GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(OUT)/$(BINARY)$(EXT) .
	cp README.md $(OUT)/
	if [ "$(OS)" = "windows" ]; then \
		cd $(DIST) && zip -qr -j $(BINARY)_$(OS)_$(ARCH).zip $(BINARY)_$(OS)_$(ARCH)/; \
	else \
		tar -czf $(DIST)/$(BINARY)_$(OS)_$(ARCH).tar.gz -C $(DIST) $(BINARY)_$(OS)_$(ARCH); \
	fi
	rm -rf $(OUT)
