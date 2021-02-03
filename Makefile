run:
	go run . -conf=dev.env.json
build:
	go build -o runner.bin
test:
	go test
fmt:
	go fmt ./...

wire:
	cd sync && wire