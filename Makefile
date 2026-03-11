SHELL:=/bin/bash

.PHONY: all
all: build

.PHONY: certs
certs:
	@echo "Generating certificates..."
	@mkdir -p certs
	@openssl req -x509 -newkey rsa:4096 -keyout certs/ca.key -out certs/ca.crt -days 365 -nodes -subj "/CN=Test CA"
	@openssl req -newkey rsa:4096 -keyout certs/server.key -out certs/server.csr -nodes -subj "/CN=localhost"
	@openssl x509 -req -in certs/server.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/server.crt -days 365 -extfile <(printf "subjectAltName=DNS:localhost")
	@openssl req -newkey rsa:4096 -keyout certs/client.key -out certs/client.csr -nodes -subj "/CN=Test Client"
	@openssl x509 -req -in certs/client.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/client.crt -days 365

.PHONY: build
build: build-server build-client

.PHONY: build-server
build-server:
	@echo "Building server..."
	@go build -o bin/server server/main.go

.PHONY: build-client
build-client:
	@echo "Building client..."
	@go build -o bin/client client/main.go

.PHONY: run-server
run-server:
	@echo "Starting server..."
	@./bin/server > server.log 2>&1

.PHONY: run-client
run-client:
	@echo "Running client..."
	@./bin/client

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf certs bin
