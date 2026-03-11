# GitHub Actions Workflows

This directory contains GitHub Actions workflows for automated testing and Docker image building.

## Workflows

### docker-build.yaml

Runs on every push and pull request to build multi-platform Docker images for testing purposes.

**Platforms:**
- linux/amd64
- linux/arm64
- linux/arm/v7

**Process:**
1. Runs Go tests first (fail-fast)
2. Builds Docker images for each platform using matrix strategy
3. Uses Docker Buildx with QEMU for cross-platform builds
4. Images are built but not pushed (testing only)

### docker-release.yaml

Runs when a new release is published to build and push multi-platform Docker images.

**Platforms:**
- linux/amd64
- linux/arm64
- linux/arm/v7

**Process:**
1. Runs Go tests first (fail-fast)
2. Builds multi-platform Docker images in a single job
3. Pushes to both Docker Hub and GitHub Container Registry
4. Tags images with version, major.minor, major, and latest

**Required Secrets:**
- `DOCKER_USERNAME`: Docker Hub username
- `DOCKER_PASSWORD`: Docker Hub password or access token
- `GITHUB_TOKEN`: Automatically provided by GitHub Actions

**Image Locations:**
- Docker Hub: `${DOCKER_USERNAME}/gochess-board:${VERSION}`
- GitHub Container Registry: `ghcr.io/${GITHUB_REPOSITORY}:${VERSION}`

## Setup Instructions

### For Docker Hub Publishing

1. Go to repository Settings → Secrets and variables → Actions
2. Add the following secrets:
   - `DOCKER_USERNAME`: Your Docker Hub username
   - `DOCKER_PASSWORD`: Your Docker Hub password or access token (recommended)

### For GitHub Container Registry

No additional setup required. The workflow uses the automatically provided `GITHUB_TOKEN`.

## Usage

### Testing Builds

Push to any branch or create a pull request. The `docker-build.yaml` workflow will:
- Run all Go tests
- Build Docker images for all platforms
- Verify the build process works correctly

### Publishing Releases

1. Create and push a new tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. Create a GitHub release from the tag

3. The `docker-release.yaml` workflow will automatically:
   - Run all Go tests
   - Build multi-platform Docker images
   - Push to Docker Hub and GitHub Container Registry
   - Tag with appropriate version labels

## Docker Image Tags

For a release `v1.2.3`, the following tags are created:
- `1.2.3`
- `1.2`
- `1`
- `latest` (if on default branch)
- `1.2.3-${GIT_SHA}`

## Running Published Images

```bash
# From Docker Hub
docker run -p 35256:35256 ${DOCKER_USERNAME}/gochess-board:latest

# From GitHub Container Registry
docker run -p 35256:35256 ghcr.io/${GITHUB_REPOSITORY}:latest

# Specific version
docker run -p 35256:35256 ${DOCKER_USERNAME}/gochess-board:1.2.3

# Specific platform
docker run --platform linux/arm64 -p 35256:35256 ${DOCKER_USERNAME}/gochess-board:latest
```
