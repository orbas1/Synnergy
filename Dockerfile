# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /src
COPY . .
RUN cd synnergy-network && go mod download && GOFLAGS="-trimpath" go build -o /synnergy ./cmd/synnergy

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /synnergy ./synnergy
COPY docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x docker-entrypoint.sh synnergy
ENTRYPOINT ["./docker-entrypoint.sh"]
