GOMODULE := github.com/yhlooo/stackcrisp
GOPKG := $(MODULE)/cmd/stackcrisp
BIN_NAME := stackcrisp
OUTPUT_ROOT := outputs
BIN_ROOT := $(OUTPUT_ROOT)/bin

.PHONY: build
build:
	go build -o "$(BIN_ROOT)/$(BIN_NAME)" "$(GOPKG)"

.PHONT: test
test:
	go test ./...

.PHONY: install
install:
	go install "$(GOPKG)"

.PHONY: fmt
fmt:
	goimports -local="$(GOMODULE)" -w .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -rf "$(OUTPUT_ROOT)"

