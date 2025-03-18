LOCAL_BIN:=$(CURDIR)/bin

install-deps:
	set GOBIN=%LOCAL_BIN%
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

get-deps:
	go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc


generate:
	make generate-dish-api

generate-dish-api:
	protoc --proto_path=api/dish_v1 --go_out=pkg/dish_v1 --go_opt=paths=source_relative --plugin=protoc-gen-go=bin/protoc-gen-go.exe --go-grpc_out=pkg/dish_v1 --go-grpc_opt=paths=source_relative --plugin=protoc-gen-go-grpc=bin/protoc-gen-go-grpc.exe api/dish_v1/dish.proto

build:
	set GOOS=linux
	set GOARCH=amd64
	go build -o service_linux cmd/grpc_server/main.go

copy-to-server:
	scp service_linux root@45.12.231.86:

docker-build-and-push:
	docker buildx build --no-cache --platform linux/amd64 -t cr.selcloud.ru/registry/test-server:v0.0.1 .
	docker login -u token -p CRgAAAAA25585DDnneMa6u0Wbc8Qk43iRLoTvDJM cr.selcloud.ru/registry
	docker push cr.selcloud.ru/registry/test-server:v0.0.1
