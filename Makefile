build:
	go fmt *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -x -ldflags="-w -s" -o transmissionBot
	mv transmissionBot ./dist
