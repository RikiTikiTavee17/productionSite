FROM golang:1.23.5-alpine AS builder

COPY . /github.com/RikiTikiTavee17/course/grpc/source/
WORKDIR /github.com/RikiTikiTavee17/course/grpc/source/

RUN go mod download
RUN go build -o ./bin/crud_server cmd/grpc_server/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /github.com/RikiTikiTavee17/course/grpc/source/bin/crud_server .

CMD ["./crud_server"]