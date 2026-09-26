FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o xanbox ./src/xanbox

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/xanbox .

CMD ["./xanbox"]