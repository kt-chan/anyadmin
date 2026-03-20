# --- Stage 1: Build Backend ---
FROM golang:1.25-bookworm AS backend-builder

WORKDIR /app/backend
# Copy go.mod and go.sum first to leverage Docker cache
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the backend source
COPY backend/ ./
# Build the backend binary
RUN CGO_ENABLED=0 GOOS=linux go build -o anyadmin-server ./cmd/server/main.go
# Build the agent binary
RUN CGO_ENABLED=0 GOOS=linux go build -o anyadmin-agent ./cmd/agent/main.go

# --- Stage 2: Build Frontend ---
FROM node:22-bookworm AS frontend-builder

WORKDIR /app/frontend
# Copy package.json and package-lock.json first
COPY frontend/package*.json ./
RUN npm install

# Copy the rest of the frontend source
COPY frontend/ ./
# Build frontend assets (Tailwind CSS, etc.)
RUN npm run build

# --- Stage 3: Final Production Image ---
FROM ubuntu:22.04

# Build arguments and environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV NODE_ENV=production
ENV GO_VERSION=1.25.6

# Install essential prerequisites
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    ca-certificates \
    gnupg \
    procps \
    git \
    && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends \
    nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# Add Go to PATH
ENV PATH=$PATH:/usr/local/go/bin

# Create 'anyadmin' user and group with UID 1000
RUN groupadd -g 1000 anyadmin && useradd -u 1000 -m -r -g anyadmin -s /bin/bash anyadmin

# Set working directory
WORKDIR /home/anyadmin/app

# Create necessary directories
RUN mkdir -p /home/anyadmin/logs /home/anyadmin/app/backend/dist /home/anyadmin/app/frontend \
    && chown -R anyadmin:anyadmin /home/anyadmin

# Copy backend binary from stage 1
COPY --from=backend-builder --chown=anyadmin:anyadmin /app/backend/anyadmin-server /home/anyadmin/app/backend/
# Copy agent binary from stage 1 to dist folder
COPY --from=backend-builder --chown=anyadmin:anyadmin /app/backend/anyadmin-agent /home/anyadmin/app/backend/dist/
# Copy backend data.json if exists
COPY --chown=anyadmin:anyadmin backend/data.json* /home/anyadmin/app/backend/
# Copy backend deployments (tars, templates, etc.)
COPY --chown=anyadmin:anyadmin backend/deployments /home/anyadmin/app/backend/deployments

# Copy frontend source and built assets from stage 2
# Note: we copy everything from frontend source because app.js needs it
COPY --from=frontend-builder --chown=anyadmin:anyadmin /app/frontend /home/anyadmin/app/frontend/
# Ensure we have the built public folder
COPY --from=frontend-builder --chown=anyadmin:anyadmin /app/frontend/public /home/anyadmin/app/frontend/public/

# Copy the entrypoint script
COPY --chown=anyadmin:anyadmin docker/entrypoint.sh /home/anyadmin/app/entrypoint.sh
RUN chmod +x /home/anyadmin/app/entrypoint.sh

# Switch to 'anyadmin' user
USER anyadmin

# Expose ports: 3000 (Frontend), 8080 (Backend)
EXPOSE 3000 8080

# Environment variables for configuration (can be overridden)
ENV PORT=3000
ENV MgmtPort=8080
ENV MgmtHost=0.0.0.0

# Start script
ENTRYPOINT ["/home/anyadmin/app/entrypoint.sh"]
