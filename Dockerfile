FROM node:22-bookworm-slim AS node

FROM golang:1.26.3

COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/bin/npm /usr/local/bin/npm
COPY --from=node /usr/local/bin/npx /usr/local/bin/npx
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules

RUN npm install -g @developer-overheid-nl/don-checker@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV GIN_MODE=release
RUN go build -o main ./cmd/main.go

EXPOSE 1337

CMD ["./main"]
