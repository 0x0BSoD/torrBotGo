build:
	go fmt *.go
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o transmissionBot
	mv transmissionBot ./dist