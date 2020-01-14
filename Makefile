.PHONY: all clean

all: fmt vet build install-local

build:
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o _output/kubectl-captain ./cmd/plugin
	cd _output && tar -zcvf kubectl-captain.tar.gz kubectl-captain

clean:
	go clean -r -x
	rm _output/kubectl-captain


mod:
	GO111MODULE=on go mod tidy


install-local:
	CGO_ENABLED=0 GO111MODULE=on  go build -o _output/kubectl-captain ./cmd/plugin
	cp _output/kubectl-captain /usr/local/bin

test:
	bash test.sh

vet:
	go vet ./...

fmt:
	go fmt ./...
