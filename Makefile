build:
	go fmt *.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -race -ldflags="-w -s" -o transmissionBot
	mv transmissionBot ./dist