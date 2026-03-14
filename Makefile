.PHONY: build test lint install deploy clean

build: homelab-mcp

homelab-mcp:
	go build -o homelab-mcp .

test:
	go test ./...

lint:
	@if which golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, falling back to go vet"; \
		go vet ./...; \
	fi

install: homelab-mcp
	sudo cp homelab-mcp /usr/local/bin/homelab-mcp

deploy: install
	sudo systemctl restart homelab-mcp

clean:
	rm -f homelab-mcp
