# Multi-stage build for GoChess Board with all chess engines
FROM golang:1.24-bookworm AS builder

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o gochess-board ./cmd/gochess-board

# Final stage - Ubuntu 24.04 for chess engines
FROM ubuntu:24.04

# Install all chess engines and dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    # UCI Engines
    stockfish \
    fruit \
    toga2 \
    # CECP Engines
    gnuchess \
    crafty \
    # Build dependencies for additional engines
    wget \
    unzip \
    ca-certificates && \
    # Clean up apt cache
    rm -rf /var/lib/apt/lists/*

# Install Dragon by Komodo Chess (if available via direct download)
# Note: Dragon/Komodo might require manual installation or licensing
# This is a placeholder - adjust based on actual availability
RUN mkdir -p /usr/local/bin/dragon

# Create non-root user for running the application
RUN useradd -m chess && \
    mkdir -p /home/chess/data && \
    chown -R chess:chess /home/chess

# Copy the built binary from builder
COPY --from=builder /build/gochess-board /usr/local/bin/gochess-board

# Make it executable
RUN chmod +x /usr/local/bin/gochess-board

# Switch to non-root user
USER chess
WORKDIR /home/chess

# Expose the default port
EXPOSE 35256

# Set the entrypoint with the specified flags
ENTRYPOINT ["/usr/local/bin/gochess-board"]

# Default arguments as specified
CMD ["--no-browser", "--restart", "--log-level", "INFO", "--book-file", "/usr/share/games/gnuchess/book.bin"]
