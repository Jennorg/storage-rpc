.PHONY: all server client clean protos test

SERVER_BIN = lbserver
CLIENT_BIN = lbclient
BIN_DIR = ./bin
LOGS_DIR = ./logs

all: server client

server:
	@echo "Building server..."
	go build -o $(BIN_DIR)/$(SERVER_BIN) ./cmd/lbserver

client:
	@echo "Building client..."
	go build -o $(BIN_DIR)/$(CLIENT_BIN) ./cmd/lbclient

clean:
	@echo "Cleaning up..."
	rm -f $(BIN_DIR)/$(SERVER_BIN) $(BIN_DIR)/$(CLIENT_BIN)
	rm -rf $(LOGS_DIR)
	@echo "Clean complete."

protos:
	@echo "Generating protobuf files..."
	@bash ./scripts/generate_proto.sh

test:
	@echo "Running tests..."
	@bash ./scripts/test.sh

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(BIN_DIR)/$(SERVER_BIN): $(BIN_DIR)
$(BIN_DIR)/$(CLIENT_BIN): $(BIN_DIR)
