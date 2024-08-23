# grpc-tutorial
nano ~/.bashrc 
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

// ALT
export GOPATH=$(go env GOPATH)
export GOBIN=$(go env GOPATH)/bin


PRE-REQ
source ~/.bashrc

STEPS
1. generate proto buffer
protoc --proto_path=proto --go_out=chatpb --go-grpc_out=chatpb chat.proto
2. build project / run
go build -o client ./client
go build -o server ./server

go run main.go at respective terminal (server and client)

