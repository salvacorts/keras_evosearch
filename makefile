.PHONY: default
default: build;

install_py_deps:
	python -m pip install -r client/requirements.txt

install_go_deps:
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto_go:
	protoc \
		--go_out=server/  \
		--go-grpc_out=server/ \
		protobuf/api.proto

proto_py:
	python -m grpc_tools.protoc \
			-I. \
			--python_out=client \
			--grpc_python_out=client \
			 protobuf/api.proto

proto: proto_go proto_py

build_server: proto_go
	cd server; \
		go mod tidy; \
		mkdir -p bin; \
		go build -o bin/main main.go

build_client: proto_py

build: build_server build_client

clean:
	rm -r server/gen
	rm -r server/bin
	rm -r client/gen

run_server:
	server/bin/main

run_client:
	python client/main.py