default: build build-docker
build:
	@echo "building"
	@echo "GOPATH=${GOPATH}"
	go build .

build-docker:
	docker build -t arykalin/getsecretfromvault .

push:
	docker push arykalin/getsecretfromvault