FROM golang:1.27 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /go-app ./cmd/api

# -- Stage 2: Runtime --
FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /go-app /go-app

# Ensure the container runs as the non-root user provided by distroless
USER nonroot:nonroot

EXPOSE 8080

ENV APP_ENV=production
ENV PORT=8080
ENTRYPOINT ["/go-app"]
