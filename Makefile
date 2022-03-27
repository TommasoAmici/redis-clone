BUILD=main

build:
	go build -o ${BUILD} main.go

run: build
	./${BUILD} -address 127.0.0.1:6380
