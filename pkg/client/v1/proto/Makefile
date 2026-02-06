.PHONY: proto install-deps help

proto:
	@echo "Generating gRPC code from proto files using protoc-gen-gogo..."
	@cd ../../../.. && go mod vendor
	@protoc \
		--proto_path=. \
		--proto_path=../../../../vendor \
		--gogo_out=. \
		--gogo_opt=paths=source_relative \
		storage_service.proto
	@echo "Proto generation complete!"
	@echo ""
	@echo "IMPORTANT: Commit the generated files (*.pb.go, *_grpc.pb.go) to the repository"
	@echo "           You can delete the vendor directory after generation"

install-deps:
	@echo "Installing protoc dependencies..."
	go install github.com/gogo/protobuf/protoc-gen-gogo@latest
	@echo "Dependencies installed!"

help:
	@echo "Available targets:"
	@echo "  make install-deps  - Install protoc-gen-gogo"
	@echo "  make proto         - Generate Go code from proto files using gogo (runs go mod vendor first)"
	@echo ""
	@echo "After running 'make proto', commit the generated files to git:"
	@echo "  git add *.pb.go *_grpc.pb.go"
	@echo "  git commit -m 'Generate proto code for storage service'"
	@echo ""
