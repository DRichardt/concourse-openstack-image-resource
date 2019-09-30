build_linux: export GOOS=linux
build_linux: build
build_darwin: export GOOS=darwin
build_darwin: build

build:
	go build -o bin/check ./cmd/check
	go build -o bin/in ./cmd/in
	go build -o bin/out ./cmd/out