nano ~/.bashrc 
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

// ALT
export GOPATH=$(go env GOPATH)
export GOBIN=$(go env GOPATH)/bin


PRE-REQ
source ~/.bashrc

STEPS
go build -o client ./client
go build -o server ./server
