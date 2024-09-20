.PHONY: todo test build clean

# Run the todo app
todo:
	go run ./cmd/todo

# Run tests
test:
	go test ./test -v

# Build app into binary
build:
	go build -o ./bin/todo ./cmd/todo

# Run clean
clean:
	go clean
