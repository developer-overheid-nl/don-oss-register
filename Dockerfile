FROM node:24-bookworm-slim AS don-checker

RUN npm install -g @developer-overheid-nl/don-checker@latest && \
    npm cache clean --force

FROM golang:1.26.5

COPY --from=don-checker /usr/local/ /usr/local/

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV GIN_MODE=release
RUN go build -p=1 -o main ./cmd/main.go

EXPOSE 1337

CMD ["./main"]
