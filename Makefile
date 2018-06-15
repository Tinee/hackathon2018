build:
	dep ensure
	env GOOS=linux go build -ldflags="-s -w" -o bin/trigger-time-options cmd/trigger-time-options/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/trigger cmd/trigger/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/status cmd/status/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/test-setup cmd/test-setup/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/set-mock-data cmd/set-mock-data/main.go