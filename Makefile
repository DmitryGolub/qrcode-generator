.PHONY: generate

PROTO_DIR = proto
OUT_DIR = backend/pkg/api/v1

generate:
	mkdir -p $(OUT_DIR)
	protoc \
		-I $(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/qrcodegen/v1/*.proto