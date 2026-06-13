.PHONY: proto build test vet fmt tidy lint all

# Regenerate the protobuf + gRPC stubs into internal/pb.
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc (go install ...).
proto:
	protoc \
	  --proto_path=proto \
	  --go_out=. --go_opt=module=github.com/ZeroRuntimeAI/zrt-go \
	  --go_opt=Mzrt_runtime.proto=github.com/ZeroRuntimeAI/zrt-go/internal/pb \
	  --go-grpc_out=. --go-grpc_opt=module=github.com/ZeroRuntimeAI/zrt-go \
	  --go-grpc_opt=Mzrt_runtime.proto=github.com/ZeroRuntimeAI/zrt-go/internal/pb \
	  proto/zrt_runtime.proto

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w zrt plugins tests

tidy:
	go mod tidy

all: fmt vet build test
