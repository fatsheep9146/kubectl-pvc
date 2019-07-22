

.PHONY: all clean

all: kubectl-pvc

kubectl-pvc:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o _output/kubectl-pvc ./cmd/plugin

clean:
	go clean -r -x
	rm _output/kubectl-pvc
