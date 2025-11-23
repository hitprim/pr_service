FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o pr_service ./cmd/app


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/pr_service ./

CMD ["./pr_service"]