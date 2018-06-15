build:
	dep ensure
	env GOOS=linux go build -ldflags="-s -w" -o bin/trigger-time-options cmd/trigger-time-options/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/trigger cmd/trigger/main.go