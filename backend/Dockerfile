FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# SERVICE maps to folders under ./cmd (core, worker-gateway, insurer-gateway, platform-gateway)
ARG SERVICE=core
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o backend ./cmd/${SERVICE}

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/backend .

EXPOSE 8000 8001 8002 8003
CMD ["./backend"]
