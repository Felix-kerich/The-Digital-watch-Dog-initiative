# ----------- Build Stage -------------
  FROM golang:1.23-alpine AS builder

  WORKDIR /app
  
  # Install build dependencies for static build
  RUN apk add --no-cache gcc musl-dev
  
  # Copy go.mod and go.sum from src/api
  COPY src/api/go.mod src/api/go.sum ./src/api/
  
  WORKDIR /app/src/api
  
  # Download Go modules
  RUN go mod download
  
  # Copy source code
  COPY src/api .
  
  # Build the binary statically
  RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -tags musl \
      -ldflags="-w -s -extldflags '-static'" -o /app/api .
  
  # ----------- Final Stage -------------
  FROM alpine:latest
  
  WORKDIR /app
  
  # Install tzdata & ca-certificates
  RUN apk add --no-cache tzdata ca-certificates && \
      cp /usr/share/zoneinfo/Africa/Nairobi /etc/localtime && \
      echo "Africa/Nairobi" > /etc/timezone && \
      rm -rf /var/cache/apk/*
  
  # Add app user
  RUN addgroup -S appgroup && adduser -S appuser -G appgroup
  
  # Prepare necessary folders
  RUN mkdir -p /app/logs /app/uploads && \
      chown -R appuser:appgroup /app
  
  # Copy the built binary and .env file
  COPY --from=builder /app/api /app/api
  COPY --from=builder /app/src/api/.env.example /app/.env
  
  # Switch to non-root user
  USER appuser
  
  # Expose port
  EXPOSE 8080
  
  # Health check endpoint (Gin API must expose this!)
  HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/api/health || exit 1
  
  # Run your Gin server
  CMD ["/app/api"]
  