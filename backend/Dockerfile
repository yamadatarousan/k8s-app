FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /todo-api ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /todo-api /todo-api

EXPOSE 8080

ENTRYPOINT ["/todo-api"]
