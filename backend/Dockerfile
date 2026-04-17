FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# SERVICE maps to folders under ./cmd (core, worker-gateway, insurer-gateway, platform-gateway)
ARG SERVICE=core
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOOS=linux go build -o backend ./cmd/${SERVICE}

# AWS Lambda Go runtime base image
FROM public.ecr.aws/lambda/go:1

# Copy the built Go binary
COPY --from=builder /app/backend ${LAMBDA_TASK_ROOT}

# Set the Lambda handler (binary name)
CMD ["backend"]

EXPOSE 8000 8001 8002 8003
CMD ["./backend"]
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# SERVICE maps to folders under ./cmd (core, worker-gateway, insurer-gateway, platform-gateway)
ARG SERVICE=core
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOOS=linux go build -o backend ./cmd/${SERVICE}

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/backend .

EXPOSE 8000 8001 8002 8003
CMD ["./backend"]
