.PHONY: build clean run

# Переменные
BINARY_NAME=yaga-bhop.exe
BUILD_DIR=build

# Основные команды
build:
	@echo "Building..."
	@if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
	go build -ldflags "-s -w -H windowsgui" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go

run: build
	@echo "Running..."
	@$(BUILD_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)

# Тесты, которых нет :з
test:
	@echo "Running tests..."
	go test ./...

# Ищет потенциальные проблемы в коде
lint:
	@echo "Running linter..."
	golangci-lint run