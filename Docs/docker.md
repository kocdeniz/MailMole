# Docker

MailMole can run in a Docker container for easy deployment.

## Quick Start

### Build and Run

```bash
# Build the image
docker build -t mailmole .

# Run interactively
docker run -it --rm mailmole
```

### Using Docker Compose

```bash
# Run with docker-compose
docker-compose up

# Run in background
docker-compose up -d
```

## Persistent Data

Mount a volume to persist state files and logs:

```bash
docker run -it --rm \
  -v $(pwd):/home/mailmole \
  mailmole
```

## Use Cases

### Bulk Migration with Volume Mount

```bash
# Create accounts file
cat > accounts.csv << EOF
user1@src.com,pass1,user1@dst.com,pass1
user2@src.com,pass2,user2@dst.com,pass2
EOF

# Run bulk migration
docker run -it --rm \
  -v $(pwd):/home/mailmole \
  mailmole
```

### Daemon Mode (Background)

```bash
docker run -d \
  --name mailmole \
  -v $(pwd):/home/mailmole \
  mailmole
```

## Configuration

### Environment Variables (Future)

| Variable | Description |
|----------|-------------|
| `MAILMOLE_LOG_FILE` | Log file path |
| `MAILMOLE_STATE_FILE` | State file path |

### Volume Mounts

| Path | Description |
|------|-------------|
| `/home/mailmole` | Working directory with state files |

## Docker Hub (Future)

Once published, you can pull from Docker Hub:

```bash
docker pull kocdeniz/mailmole:latest
docker run -it --rm kocdeniz/mailmole:latest
```

## System Requirements

- Docker Engine 20.10+
- Linux containers (Alpine-based)

## Building for Different Architectures

```bash
# AMD64 (Intel/AMD)
docker build -t mailmole:amd64 --platform linux/amd64 .

# ARM64 (Apple Silicon, Raspberry Pi)
docker build -t mailmole:arm64 --platform linux/arm64 .

# All platforms
docker buildx build --platform linux/amd64,linux/arm64 -t mailmole:latest .
```

## Troubleshooting

### Permission Denied

If you get permission errors, ensure the volume directory is writable:

```bash
chmod 755 .
```

Or run with specific user:

```bash
docker run -it --rm -u $(id -u):$(id -g) \
  -v $(pwd):/home/mailmole \
  mailmole
```
