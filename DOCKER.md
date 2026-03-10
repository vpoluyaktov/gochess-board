# Docker Deployment Guide

This guide explains how to build and run the GoChess Board application in a Docker container with all chess engines pre-installed.

## Included Chess Engines

The Docker image includes the following chess engines:

### UCI Protocol Engines
- **Stockfish 16** - World's strongest open-source engine (ELO: 1320-3190)
- **Fruit 2.1** - Strong tactical engine
- **Toga II 3.0** - Fruit derivative

### CECP/XBoard Protocol Engines
- **GNU Chess 6.2.9** - Classic free chess engine (built from source, fixes CVE-2021-30184)
- **Crafty 23.4** - Strong traditional engine

### Built-in Engine
- **GoChess** - Native Go implementation (~1000-1200 ELO)

## Quick Start

### Using Docker Compose (Recommended)

1. Build and start the container:
   ```bash
   docker-compose up -d
   ```

2. View logs:
   ```bash
   docker-compose logs -f
   ```

3. Stop the container:
   ```bash
   docker-compose down
   ```

### Using Docker CLI

1. Build the image:
   ```bash
   docker build -t gochess-board:latest .
   ```

2. Run the container:
   ```bash
   docker run -d \
     --name gochess-board \
     -p 35256:35256 \
     gochess-board:latest
   ```

3. View logs:
   ```bash
   docker logs -f gochess-board
   ```

4. Stop and remove the container:
   ```bash
   docker stop gochess-board
   docker rm gochess-board
   ```

## Accessing the Application

Once the container is running, open your web browser and navigate to:
```
http://localhost:35256
```

## Default Configuration

The Docker container starts with the following flags:
- `--no-browser` - Don't auto-open browser (not applicable in container)
- `--no-tui` - Disable TUI (required for Docker)
- `--restart` - Kill any existing process before starting
- `--log-level INFO` - Set logging level to INFO
- `--book-file /usr/share/games/fruit/book_small.bin` - Use Fruit opening book (31,467 entries)

## Customizing Configuration

### Override Command Line Arguments

Using Docker Compose, edit `docker-compose.yml`:
```yaml
services:
  gochess-board:
    # ... other settings ...
    command: ["--no-browser", "--log-level", "DEBUG", "--book-file", "/usr/share/games/gnuchess/book.bin"]
```

Using Docker CLI:
```bash
docker run -d \
  --name gochess-board \
  -p 35256:35256 \
  gochess-board:latest \
  --no-browser --log-level DEBUG --book-file /usr/share/games/gnuchess/book.bin
```

### Change Port

Using Docker Compose, edit `docker-compose.yml`:
```yaml
services:
  gochess-board:
    ports:
      - "8080:35256"  # Map host port 8080 to container port 35256
```

Using Docker CLI:
```bash
docker run -d \
  --name gochess-board \
  -p 8080:35256 \
  gochess-board:latest
```

Then access at `http://localhost:8080`

### Enable Persistent Engines

Add the `--persistent-engines` flag:
```bash
docker run -d \
  --name gochess-board \
  -p 35256:35256 \
  gochess-board:latest \
  --no-browser --persistent-engines --book-file /usr/share/games/gnuchess/book.bin
```

## Volume Mounts

### Mount Custom Opening Book

```bash
docker run -d \
  --name gochess-board \
  -p 35256:35256 \
  -v /path/to/your/book.bin:/data/book.bin:ro \
  gochess-board:latest \
  --no-browser --no-tui --book-file /data/book.bin
```

**Note**: Opening books must be in Polyglot binary format (.bin) with proper 16-byte alignment.

### Persist Logs

Using Docker Compose (already configured):
```yaml
volumes:
  - ./logs:/home/chess/logs
```

Using Docker CLI:
```bash
docker run -d \
  --name gochess-board \
  -p 35256:35256 \
  -v $(pwd)/logs:/home/chess/logs \
  gochess-board:latest
```

## Troubleshooting

### Check Container Status
```bash
docker ps -a | grep gochess-board
```

### View Real-time Logs
```bash
docker logs -f gochess-board
```

### Access Container Shell
```bash
docker exec -it gochess-board /bin/bash
```

### Verify Engines are Installed
```bash
docker exec gochess-board which stockfish
docker exec gochess-board which gnuchess
docker exec gochess-board which fruit
docker exec gochess-board which toga2
docker exec gochess-board which crafty
```

### Check Opening Book
```bash
docker exec gochess-board ls -lh /usr/share/games/gnuchess/book.bin
```

## Building for Different Architectures

### Build for ARM64 (e.g., Raspberry Pi, Apple Silicon)
```bash
docker buildx build --platform linux/arm64 -t gochess-board:arm64 .
```

### Build for AMD64 (standard x86_64)
```bash
docker buildx build --platform linux/amd64 -t gochess-board:amd64 .
```

### Multi-architecture Build
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t gochess-board:latest .
```

## Production Deployment

### Using Docker Compose with Restart Policy
```yaml
services:
  gochess-board:
    build: .
    restart: unless-stopped
    ports:
      - "35256:35256"
```

### Behind a Reverse Proxy (Nginx)

Example Nginx configuration:
```nginx
server {
    listen 80;
    server_name chess.example.com;

    location / {
        proxy_pass http://localhost:35256;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## Security Considerations

- The container runs as a non-root user (`chess` with UID 1000)
- Only port 35256 is exposed
- No sensitive data is stored in the container
- Opening book is read-only

## Image Size Optimization

The current image uses Ubuntu 24.04 to ensure all chess engines are available. The approximate size is:
- Base Ubuntu 24.04: ~80MB
- Chess engines: ~50MB
- Go binary: ~30MB
- **Total**: ~160MB

## Notes

- **Dragon by Komodo Chess**: Not automatically installed
   - Download requires manual intervention due to website certificate issues
   - Dragon 1 is free but must be manually added to the container
   - To add manually: Download from https://komodochess.com/ and copy to `/usr/games/dragon`

- **GNU Chess smallbook.bin**: Has alignment issues (not 16-byte aligned)
   - Container uses Fruit's book_small.bin instead (31,467 entries)
   - GNU Chess smallbook.bin is available but not recommended
- **TUI Mode**: The `--no-tui` flag is set by default in Docker
- **Browser Auto-open**: The `--no-browser` flag is set because browser auto-open doesn't work in containers

## Support

For issues related to:
- **Docker setup**: Check this guide
- **Chess engines**: Refer to the main README.md
- **Application features**: See the main README.md
