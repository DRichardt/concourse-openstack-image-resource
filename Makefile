IMAGE := drichardt/concourse-openstack-image-resource
TAG   := 0.0.2

build_linux: export GOOS=linux
build_linux: build
build_darwin: export GOOS=darwin
build_darwin: build

build:
	go build -o bin/check ./cmd/check
	go build -o bin/in ./cmd/in
	go build -o bin/out ./cmd/out

image:
	docker build -t $(IMAGE):$(TAG) $(BUILD_ARGS) .