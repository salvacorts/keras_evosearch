.PHONY: default
default: build;

install_py_deps:
	python -m pip install -r client/requirements.txt

install_go_deps:
	echo TODO

proto_go:
	protoc \
		--go_out=server/  \
		--go-grpc_out=server/ \
		protobuf/api.proto

proto_py:
	mkdir -p client/gen
	python -m grpc_tools.protoc \
			 -Iprotobuf \
			 --python_out=client/gen \
			 --grpc_python_out=client/gen \
			 protobuf/api.proto

proto: proto_go proto_py

build_server: proto_go
	cd server; \
		mkdir -p bin; \
		go build -o bin/main main.go

build_client: proto_py
	echo TODO

build: build_server build_client

clean:
	rm -r server/gen
	rm -r server/bin
	rm -r client/gen

run_server:
	server/bin/main