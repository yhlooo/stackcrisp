PKG := github.com/yhlooo/stackcrisp/cmd/stackcrisp
BIN_NAME := stackcrisp
OUTPUT_ROOT := outputs
BIN_ROOT := $(OUTPUT_ROOT)/bin

.PHONY: build
build: clean
	go build -o "$(BIN_ROOT)/$(BIN_NAME)" "$(PKG)"

.PHONY: install
install:
	go install "$(PKG)"

.PHONY: clean
clean:
	rm -rf "$(OUTPUT_ROOT)"

