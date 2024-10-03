build: 
	go build -o ./bin/vm-manager

run: build
	sudo ./bin/vm-manager

test: build
	go test -v ./...
