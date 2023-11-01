default:help

##@ General

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: run
run: ## Run application in docker
	docker compose up --build

.PHONY: pre-commit
pre-commit: ## Run linters and formatters via pre-commit
	@pre-commit run --all-files

.PHONY: gen-proto
gen-proto: ## Generate protobuf files
	@protoc \
		--go_opt=paths=source_relative --go_out=. \
		--go-grpc_opt=paths=source_relative --go-grpc_out=. \
		proto/*.proto
