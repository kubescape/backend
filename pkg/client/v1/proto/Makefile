.PHONY: proto install-deps help

proto:
	@echo "Generating gRPC code from proto files..."
	@cd ../../../.. && go mod vendor
	@protoc \
		--proto_path=. \
		--proto_path=../../../../vendor \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		storage_service.proto
	@echo "Proto generation complete!"
	@echo ""
	@echo "IMPORTANT: Commit the generated files (*.pb.go, *_grpc.pb.go) to the repository"
	@echo "           You can delete the vendor directory after generation"

install-deps:
	@echo "Installing protoc dependencies..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Dependencies installed!"

help:
	@echo "Available targets:"
	@echo "  make install-deps  - Install protoc-gen-go and protoc-gen-go-grpc"
	@echo "  make proto         - Generate Go code from proto files (runs go mod vendor first)"
	@echo ""
	@echo "After running 'make proto', commit the generated files to git:"
	@echo "  git add *.pb.go *_grpc.pb.go"
	@echo "  git commit -m 'Generate proto code for storage service'"
	@echo ""
