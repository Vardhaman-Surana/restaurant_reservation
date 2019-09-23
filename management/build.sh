export GO111MODULE=on
export GOOS=linux
export GOARCH=amd64
go mod download
go mod verify
go build -v -o ./bin/server ./cmd/