.RECIPEPREFIX := >
GO_MODULE_DIR := synnergy-network

.PHONY: go-build go-test go-cycle node-install node-test build-matrix all

go-build:
>cd $(GO_MODULE_DIR) && go build ./...

go-test:
>cd $(GO_MODULE_DIR) && go test ./...

go-cycle:
>./scripts/check_circular_imports.sh

node-install:
>cd synnergy-network/GUI/token-creation-tool/server && npm ci
>cd synnergy-network/GUI/dao-explorer/backend && npm ci
>cd synnergy-network/GUI/smart-contract-marketplace && npm ci
>cd synnergy-network/GUI/storage-marketplace/backend && npm ci
>cd synnergy-network/GUI/nft_marketplace/backend && npm ci

node-test:
>cd synnergy-network/GUI/token-creation-tool/server && npm run test --if-present
>cd synnergy-network/GUI/dao-explorer/backend && npm run test --if-present
>cd synnergy-network/GUI/smart-contract-marketplace && npm run test --if-present
>cd synnergy-network/GUI/storage-marketplace/backend && npm run test --if-present
>cd synnergy-network/GUI/nft_marketplace/backend && npm run test --if-present

all: go-build node-install

build-matrix:
>scripts/build_matrix.sh

