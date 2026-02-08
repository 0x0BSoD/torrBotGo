# torrBotGo - Transmission Telegram Bot

![Build](https://github.com/0x0BSoD/torrBotGo/workflows/Build/badge.svg)
![Release](https://github.com/0x0BSoD/torrBotGo/workflows/Release/badge.svg)

A feature-rich Telegram bot that provides a convenient interface for managing Transmission BitTorrent client. Control your torrent downloads remotely through Telegram with an intuitive interface.

## Features

- **Add Torrents**: Add torrents via magnet links or `.torrent` files
- **Smart Categorization**: Automatically categorize torrents based on RuTracker categories
- **Torrent Management**: Start, stop, pause, and remove torrents
- **Queue Control**: Change torrent priority (move up/down/top/bottom in queue)
- **Real-time Status**: Get detailed information about torrent status and progress
- **Search Functionality**: Search torrents by name with prefix `t:` (e.g., `t:Ubuntu`)
- **File Management**: View files within torrents and their download status
- **Notifications**: Receive messages when torrents complete downloading
- **RuTracker Integration**: Automatically fetch torrent images from RuTracker when available
- **Multi-category Support**: Organize downloads into configurable categories

## Quick Start

### Prerequisites

- Go 1.25.4 or later
- Transmission daemon running with RPC enabled
- Telegram Bot Token (from [@BotFather](https://t.me/botfather))

### Configuration

Configuration has the following structure:

```yaml
app:
  error_media: "./tmp/media/error.png"  # Path to error image
  auto_categories: true                 # Enable auto-categorization
  dirs:
    working: "./tmp"                    # Working directory
    images: "./tmp/images"              # Directory for torrent images
    download: "./tmp/downloads/"        # Base download directory
    categories:
      Films:
        path: "Films"                   # Subdirectory for Films
        matcher: "Зарубежное кино,Арт-хаус и авторское кино,HD Video,UHD Video,Мультфильмы"
      TVSeries:
        path: "TVSeries"                # Subdirectory for TV Series
        matcher: "Сериалы,Мультсериалы"
      Anime:
        path: "Anime"                   # Subdirectory for Anime
        matcher: "Аниме"
      Soft:
        path: "Soft"                    # Subdirectory for Software
        matcher: "Программы и Дизайн"
      Education:
        path: "Education"               # Subdirectory for Educational content
        matcher: "Обучение иностранным языкам,Обучающие видео"
      Games:
        path: "Games"                   # Subdirectory for Games
        matcher: "Игры"

telegram:
  token: "YOUR_TELEGRAM_BOT_TOKEN"      # Bot token from @BotFather
  chat_id: YOUR_CHAT_ID                 # Optional: Restrict to specific chat

transmission:
  config:
    uri: "http://localhost:9091/transmission/rpc"  # Transmission RPC URL
    user: "admin"                                  # Transmission username
    password: "admin"                              # Transmission password
```

### Running the Bot

```bash
# Start the bot with your configuration
./torrbot serve --config config.yaml

# Or specify a different config path
./torrbot serve -c /path/to/config.yaml
```

## Usage

### Commands

- `/start` or `/help` - Show help message and main keyboard
- `/status` - Show Transmission server status (active/paused torrents, speeds, free space)
- `/config` - Show Transmission session configuration

### Main Features

#### Adding Torrents
- **Magnet Links**: Send any magnet link directly to the bot
- **Torrent Files**: Upload `.torrent` files directly to the bot
- **RuTracker Integration**: When adding `.torrent` files from RuTracker, the bot automatically:
  - Fetches the torrent image
  - Suggests appropriate category based on RuTracker categories
  - Shows the image in Telegram with category selection buttons

#### Managing Torrents
- **Search**: Type any text to search torrents by name
- **Filter Views**: Use the main keyboard buttons to view:
  - "All torrents" - Show all torrents
  - "Active torrents" - Show only active downloads/uploads
  - "Not Active torrents" - Show only paused/stopped torrents

#### Torrent Controls
When viewing a torrent, you can:
- **Start/Stop** - Control torrent activity
- **Priority** - Change queue position (move up/down/top/bottom)
- **Files** - View files within the torrent
- **Delete** - Remove torrent (with or without data)
- **Refresh** - Update torrent information

### Search Syntax
- `t:searchterm` - Search torrents by name (e.g., `t:Ubuntu`)
- Plain text - Also searches torrents by name

## Docker Deployment

### Building the Docker Image

```bash
# Clone the repository
git clone https://github.com/0x0BSoD/torrBotGo.git
cd torrBotGo

# Build manually
docker build -t torrbotgo .
```

### Quick Docker Run

```bash
# Create config.yaml from the example
cp config/config.yaml.example config.yaml
# Edit config.yaml with your settings

# Run the container
docker run --rm \
  -v "$(pwd)/config.yaml":/app/config.yaml:ro \
  torrbotgo serve
```

### Using Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  torrbotgo:
    build: .
    container_name: torrbotgo
    restart: unless-stopped
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      # Optional: mount download directory for persistence
      # - ./downloads:/downloads
    environment:
      - TZ=UTC
```

Run with:
```bash
docker-compose up -d
```

### Docker Features

- **Multi-stage build**: Small final image (~13MB) using distroless base
- **Non-root user**: Runs as nonroot user for security
- **Build arguments**: Supports version, git commit, and build date injection
- **Timezone support**: Includes timezone data for proper time handling
- **SSL certificates**: Includes CA certificates for secure connections

### Environment Variables

You can override configuration using environment variables:
```bash
docker run --rm \
  -e TELEGRAM_TOKEN="your_token_here" \
  -e TRANSMISSION_HOST="transmission" \
  -e TRANSMISSION_PORT=9091 \
  -e TRANSMISSION_USERNAME="admin" \
  -e TRANSMISSION_PASSWORD="password" \
  torrbotgo serve
```

### Building with Version Information

```bash
# Build manually with build args
docker build \
  --build-arg VERSION="2.0.1" \
  --build-arg GIT_COMMIT="$(git rev-parse --short HEAD)" \
  --build-arg BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -t torrbotgo:2.0.1 .
```

## [UNMAINTAINED] Ansible Deployment

An Ansible role is available for automated deployment:
[https://github.com/0x0BSoD/ansible/tree/master/transmission](https://github.com/0x0BSoD/ansible/tree/master/transmission)

Check the `vars/main.yml` and `defaults/main.yml` directories for configuration options.

Default Ansible variables:
```yaml
user: root
rpc_user: user
rpc_password: pass123
```

## Configuration Example

See `config.example.yaml` for a complete configuration template with detailed comments.

Quick start:
```bash
cp config/config.yaml.example config.yaml
# Edit config.yaml with your settings
```

Key configuration sections:
- `telegram`: Bot token and proxy settings
- `transmission`: RPC connection details
- `app`: Application behavior and directories
- `logging`: Log level and format control

## Architecture

### Project Structure

```
torrBotGo/
├── cmd/                    # CLI commands (Cobra framework)
│   ├── root.go             # Root command definition
│   └── serve.go            # Serve command implementation
├── config/                 # Configuration management
│   ├── config.go           # YAML config parsing and validation
│   └── config.yaml.example # Example configuration file
├── internal/               # Core application logic
│   ├── app/                # Telegram message handlers
│   │   ├── router.go        # Message routing
│   │   ├── handleCommand.go # Command handlers
│   │   ├── handleMessage.go # Message handlers
│   │   ├── handleInline.go  # Inline button handlers
│   │   └── helpers.go       # Helper functions
│   ├── telegram/         # Telegram API integration
│   │   ├── telegram.go   # Telegram client
│   │   ├── keyboards.go  # Inline keyboard definitions
│   │   ├── templates.go  # Message templates
│   │   └── helpers.go    # Telegram helpers
│   ├── transmission/       # Transmission RPC client
│   │   ├── transmission.go # Transmission client
│   │   ├── torrent.go      # Torrent operations
│   │   ├── fetcher.go      # RuTracker web scraping
│   │   ├── watcher.go      # Torrent state watcher
│   │   └── helpers.go      # Transmission helpers
│   ├── events/       # Event bus system
│   │   ├── events.go # Event definitions
│   │   └── bus.go    # Event bus implementation
│   └── cache/       # In-memory cache
│       └── cache.go # Torrent cache implementation
├── pkg/logger/   # Logging utilities
│   └── logger.go # Zap logger configuration
├── main.go         # Application entry point
├── go.mod          # Go module definition
├── go.sum          # Dependency checksums
├── config.yaml     # Configuration template
└── README.MD       # This file
```

### Key Components

1. **Telegram Integration**: Handles all Telegram Bot API interactions
2. **Transmission Client**: Manages all Transmission RPC communications
3. **Event Bus**: Enables decoupled communication between components
4. **Cache System**: Maintains in-memory torrent state for efficient operations
5. **RuTracker Fetcher**: Scrapes RuTracker for torrent metadata and images

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/0x0BSoD/torrBotGo.git
cd torrBotGo

# Install dependencies
go mod download

# Build the application
go build -o torrbot

# Run tests (if available)
go test ./...
```

### Dependencies

- **Cobra**: CLI framework
- **Zap**: Structured logging
- **goquery**: HTML parsing for RuTracker integration
- **transmission**: Transmission RPC client library
- **telegram-bot-api**: Telegram Bot API wrapper

### Code Style

The project follows standard Go conventions:
- `gofmt` for code formatting
- Descriptive package and function names
- Comprehensive error handling
- Context-aware operations for cancellation
- Thread-safe data structures

## Troubleshooting

### Common Issues

1. **Transmission Connection Failed**
   - Verify Transmission RPC is enabled
   - Check URI format: `http://host:port/transmission/rpc`
   - Confirm username and password are correct

2. **Telegram Bot Not Responding**
   - Verify bot token is correct
   - Check that the bot has been started with `/start`
   - Ensure `chat_id` is set if restricting to specific chat

3. **RuTracker Images Not Loading**
   - RuTracker may block requests from certain IPs
   - Check network connectivity to RuTracker
   - Images directory must be writable

4. **Permission Errors**
   - Ensure download directories exist and are writable
   - Docker containers need proper volume mounts with correct permissions

### Logging

The application uses structured logging with Zap. Log level can be controlled via configuration. Check logs for detailed error information.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the terms of the LICENSE file included in the repository.

## Acknowledgments

- [Transmission](https://transmissionbt.com/) for the excellent BitTorrent client
- [RuTracker](https://rutracker.org/) for torrent metadata
- All contributors and users of torrBotGo

## Support

For issues, questions, or feature requests:
- Open an issue on [GitHub](https://github.com/0x0BSoD/torrBotGo/issues)
- Check existing issues for solutions

---

**Note**: This bot is intended for personal use with content you have the rights to download. Always respect copyright laws and terms of service.
