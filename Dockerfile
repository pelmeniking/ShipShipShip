# Multi-stage build for admin panel and backend (AMD64 & ARM64)

# Stage 1: Build the admin panel
FROM node:22-slim AS admin-build
WORKDIR /app/admin

# Copy admin package files
COPY admin/package*.json ./

# Install the exact dependency versions from package-lock.json
RUN npm ci

# Copy admin source
COPY admin/ ./

# Build the admin panel
RUN npm run build

# Stage 2: Build the backend
FROM golang:1.27-trixie AS backend-build
WORKDIR /app/backend

# Install dependencies for CGO + SQLite
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files and download deps
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source
COPY backend/ ./

# Build the backend (CGO enabled for SQLite)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 3: Final runtime image
FROM debian:trixie-slim
WORKDIR /app

# Runtime deps (include sqlite3 which pulls libsqlite3-0)
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Data dir
RUN mkdir -p /app/data

# Copy backend binary
COPY --from=backend-build /app/backend/main /app/

# Copy built admin panel
COPY --from=admin-build /app/admin/build /app/admin/build

# Environment
ENV GIN_MODE=release
ENV PORT=8080
ENV DB_PATH=/app/data/changelog.db

# Expose port and persist data
EXPOSE 8080
VOLUME ["/app/data"]

# Start app
CMD ["./main"]
