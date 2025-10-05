VERSION ?= dev
BUILD_DIR = build

.PHONY: all
all: build

.PHONY: build
build:
	go build -o dnote-pg2sqlite

.PHONY: test
test:
	go test -v

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -f dnote-pg2sqlite

.PHONY: release
release: clean
ifndef VERSION
	$(error VERSION is required. Usage: make VERSION=1.0.0 release)
endif
	@echo "==> Building binaries for version $(VERSION)"
	@mkdir -p $(BUILD_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/dnote-pg2sqlite-linux-amd64
	cd $(BUILD_DIR) && tar -czf dnote-pg2sqlite-$(VERSION)-linux-amd64.tar.gz dnote-pg2sqlite-linux-amd64
	cd $(BUILD_DIR) && shasum -a 256 dnote-pg2sqlite-$(VERSION)-linux-amd64.tar.gz >> checksums.txt

	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/dnote-pg2sqlite-linux-arm64
	cd $(BUILD_DIR) && tar -czf dnote-pg2sqlite-$(VERSION)-linux-arm64.tar.gz dnote-pg2sqlite-linux-arm64
	cd $(BUILD_DIR) && shasum -a 256 dnote-pg2sqlite-$(VERSION)-linux-arm64.tar.gz >> checksums.txt

	# macOS AMD64 (Intel)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/dnote-pg2sqlite-darwin-amd64
	cd $(BUILD_DIR) && tar -czf dnote-pg2sqlite-$(VERSION)-darwin-amd64.tar.gz dnote-pg2sqlite-darwin-amd64
	cd $(BUILD_DIR) && shasum -a 256 dnote-pg2sqlite-$(VERSION)-darwin-amd64.tar.gz >> checksums.txt

	# macOS ARM64 (M1/M2)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/dnote-pg2sqlite-darwin-arm64
	cd $(BUILD_DIR) && tar -czf dnote-pg2sqlite-$(VERSION)-darwin-arm64.tar.gz dnote-pg2sqlite-darwin-arm64
	cd $(BUILD_DIR) && shasum -a 256 dnote-pg2sqlite-$(VERSION)-darwin-arm64.tar.gz >> checksums.txt

	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/dnote-pg2sqlite-windows-amd64.exe
	cd $(BUILD_DIR) && zip dnote-pg2sqlite-$(VERSION)-windows-amd64.zip dnote-pg2sqlite-windows-amd64.exe
	cd $(BUILD_DIR) && shasum -a 256 dnote-pg2sqlite-$(VERSION)-windows-amd64.zip >> checksums.txt

	@echo "==> Binaries built in $(BUILD_DIR)/"
	@echo "==> Checksums:"
	@cat $(BUILD_DIR)/checksums.txt
