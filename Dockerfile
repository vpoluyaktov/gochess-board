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
    # CECP Engines - skip gnuchess from apt (has buffer overflow bug)
    crafty \
    # Build dependencies
    build-essential \
    wget \
    unzip \
    tar \
    ca-certificates && \
    # Clean up apt cache
    rm -rf /var/lib/apt/lists/*

# Build GNU Chess 6.2.9 from source (fixes CVE-2021-30184 buffer overflow)
RUN cd /tmp && \
    wget https://ftp.gnu.org/gnu/chess/gnuchess-6.2.9.tar.gz && \
    tar xzf gnuchess-6.2.9.tar.gz && \
    cd gnuchess-6.2.9 && \
    ./configure --prefix=/usr && \
    make && \
    make install && \
    cd / && \
    rm -rf /tmp/gnuchess-6.2.9*

# Install Dragon by Komodo Chess (Dragon 1 is free)
RUN cd /tmp && \
    (wget --no-check-certificate -O dragon.zip https://komodochess.com/downloads/Dragon1-Linux.zip && \
    unzip -q dragon.zip && \
    find . -name "dragon*" -type f -executable -exec mv {} /usr/games/dragon \; && \
    chmod +x /usr/games/dragon 2>/dev/null && \
    rm -rf dragon.zip dragon* || true) && \
    (test -f /usr/games/dragon && echo "Dragon installed successfully" || echo "Dragon download failed - skipping")

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

# Add /usr/games to PATH for chess engines
ENV PATH="/usr/games:${PATH}"

# Expose the default port
EXPOSE 35256

# Set the entrypoint with the specified flags
ENTRYPOINT ["/usr/local/bin/gochess-board"]

# Default arguments as specified (using Fruit's book - properly formatted)
CMD ["--no-browser", "--no-tui", "--restart", "--log-level", "INFO", "--book-file", "/usr/share/games/fruit/book_small.bin"]
