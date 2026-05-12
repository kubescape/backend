.PHONY: proto install-deps help

proto:
	@echo "Generating message bindings with protoc-gen-gogo..."
	@cd ../../../.. && go mod vendor
	@protoc \
		--proto_path=. \
		--proto_path=../../../../vendor \
		--gogo_out=. \
		--gogo_opt=paths=source_relative \
		storage_service.proto
	@echo "Generating gRPC service stubs with protoc-gen-go-grpc..."
	@protoc \
		--proto_path=. \
		--proto_path=../../../../vendor \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		storage_service.proto
	@cd ../../../.. && rm -rf vendor
	@echo "Proto generation complete!"
	@echo ""
	@echo "IMPORTANT: Commit the generated files (*.pb.go, *_grpc.pb.go) to the repository"

install-deps:
	@echo "Installing protoc dependencies..."
	go install github.com/gogo/protobuf/protoc-gen-gogo@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Dependencies installed!"

help:
	@echo "Available targets:"
	@echo "  make install-deps  - Install protoc-gen-gogo + protoc-gen-go-grpc"
	@echo "  make proto         - Generate Go code from proto files (messages via gogo, services via go-grpc)"
	@echo ""
	@echo "After running 'make proto', commit the generated files to git:"
	@echo "  git add *.pb.go *_grpc.pb.go"
	@echo "  git commit -m 'Generate proto code for storage service'"
	@echo ""
