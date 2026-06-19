APP_NAME := tictactoe
MAIN_PATH := ./cmd/tictactoe/main.go
BIN_DIR := ./bin

.PHONY: build run clean

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

clean:
	rm -rf $(BIN_DIR)
