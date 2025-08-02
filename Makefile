.RECIPEPREFIX := >
GO_MODULE_DIR := synnergy-network

.PHONY: go-build go-test node-install node-test all

go-build:
>cd $(GO_MODULE_DIR) && go build ./...

go-test:
>cd $(GO_MODULE_DIR) && go test ./...

node-install:
>cd synnergy-network/GUI/token-creation-tool/server && npm install
>cd synnergy-network/GUI/dao-explorer/backend && npm install
>cd synnergy-network/GUI/smart-contract-marketplace && npm install
>cd synnergy-network/GUI/storage-marketplace/backend && npm install
>cd synnergy-network/GUI/nft_marketplace/backend && npm install

node-test:
>cd synnergy-network/GUI/token-creation-tool/server && npm run test --if-present
>cd synnergy-network/GUI/dao-explorer/backend && npm run test --if-present
>cd synnergy-network/GUI/smart-contract-marketplace && npm run test --if-present
>cd synnergy-network/GUI/storage-marketplace/backend && npm run test --if-present
>cd synnergy-network/GUI/nft_marketplace/backend && npm run test --if-present

all: go-build node-install
