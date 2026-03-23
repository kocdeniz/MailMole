# MailMole

![MailMole Header](mailmole.jpeg)

MailMole is a terminal-first IMAP-to-IMAP migration tool written in Go.
It is designed for practical mailbox migration with a clear TUI, low-memory
message transfer, and compatibility with older Linux systems.

## Current capabilities

- **Interactive TUI** built with Bubble Tea and Lip Gloss
- **Preview Mode** - Review folders and message counts before migration
- **Web Dashboard** - Monitor migrations from any browser (optional)
- Intro/branding screen and mode selection
- Manual mode (single account pair)
- Bulk mode (multiple account pairs from file)
- Real IMAP authentication and folder discovery
- Folder creation on destination when needed
- UID-based message copy from source to destination
- Smart Retry with exponential backoff (`2s`, `5s`, `10s` + jitter)
- O(1) duplicate detection via in-memory `Message-ID` cache
- Batch metadata fetch for faster pre-filtering (`50` per batch)
- Parallel folder workers (up to `3` concurrently)
- Checkpoint persistence (`migration_state.json`) for bulk resume/skip
- Real-time activity log and progress updates
- Per-account fault isolation in bulk mode (continue on errors)
- `CGO_ENABLED=0` compatible build

## Why MailMole is different

### Smart Retry (resilience under real-world server pressure)

MailMole does not fail fast on transient IMAP/network errors. It automatically:

- Detects retryable errors (`timeout`, `server busy`, throttling patterns)
- Retries with exponential backoff (`2s`, `5s`, `10s`) + jitter
- Attempts socket/session recovery to avoid zombie connections
- Continues processing other folders/accounts when one unit fails

This is especially important on SmarterMail and enterprise servers with
throttling/rate-limit behavior.

### O(1) Caching (high-speed duplicate detection)

Instead of running a slow duplicate search per message, MailMole:

1. Loads destination folder `Message-ID` values once into memory
2. Keeps them in a `map[string]bool`
3. Uses O(1) lookup per source message

This removes repeated server-side search overhead and is one of the key reasons
for high throughput in large migrations.

## Requirements

- Go 1.20+ (only for building from source)
- IMAP access to both source and destination servers
- Network access to IMAP ports (usually `993` for TLS)

## 🚀 Installation

Choose one of the following installation methods:

### Option 1: Automatic Installation (Recommended)

Run the install script to automatically install dependencies and build:

```bash
# Clone the repository
git clone https://github.com/kocdeniz/MailMole.git
cd MailMole

# Run installation script
chmod +x install.sh
./install.sh
```

The script will:
- ✅ Check and install Go if needed
- ✅ Download dependencies
- ✅ Build the binary
- ✅ Create example files
- ✅ Set permissions

### Option 2: Using Make

```bash
# Clone and enter directory
git clone https://github.com/kocdeniz/MailMole.git
cd MailMole

# Install and build
make install

# Or just build (if Go is already installed)
make build
```

Available make commands:
```bash
make help       # Show all commands
make build      # Build the binary
make run        # Run with terminal UI
make web        # Run with web dashboard
make install    # Full installation
make clean      # Clean build artifacts
make test       # Run tests
make release    # Build for all platforms
make docker-build  # Build Docker image
make setup      # Setup project directories
```

### Option 3: Using Docker

```bash
# Build and run with Docker
docker-compose up --build

# Or manually
docker build -t mailmole .
docker run -p 8080:8080 mailmole
```

Access at: `http://localhost:8080`

### Option 4: Download Pre-built Binary

```bash
# Linux (amd64)
wget https://github.com/kocdeniz/MailMole/releases/latest/download/mailmole_linux_amd64
chmod +x mailmole_linux_amd64
./mailmole_linux_amd64

# macOS (amd64)
wget https://github.com/kocdeniz/MailMole/releases/latest/download/mailmole_darwin_amd64
chmod +x mailmole_darwin_amd64
./mailmole_darwin_amd64

# Windows
# Download mailmole_windows_amd64.exe from releases page
```

### Option 5: Build from Source

```bash
# Clone repository
git clone https://github.com/kocdeniz/MailMole.git
cd MailMole

# Download dependencies
go mod download

# Build
CGO_ENABLED=0 go build -o mailmole .

# Run
./mailmole
```

## Quick Start

After installation, start MailMole:

```bash
# Terminal UI (default)
./mailmole

# Web Dashboard
./mailmole -web :8080

# Web Dashboard only (no terminal)
./mailmole -web :8080 -web-only
```

Then open `http://localhost:8080` for web interface.

## Usage flow

1. Intro screen: press any key.
2. Select migration mode:
   - `1` Manual Entry
   - `2` Bulk Migration via File

### Manual mode

Fill these fields:

- Source Host:Port
- Source Username
- Source Password
- Destination Host:Port
- Destination Username
- Destination Password

Press `Enter` on the last field to continue.

#### Preview Mode

After connecting, you'll see the **Preview Screen** showing:
- All folders from the source account
- Message count per folder
- Estimated total size
- Estimated migration duration

**Controls:**
- `↑/↓` - Navigate folders
- `Space` - Toggle folder selection
- `a` - Select all folders
- `n` - Select none
- `Enter` - Start migration with selected folders
- `Esc` - Go back

If connection fails, the app returns to the form and shows the exact error.

### Bulk mode

Fill these fields:

- Global Source Host:Port
- Global Destination Host:Port
- Accounts File

Accepted file extensions: `.csv`, `.txt`

File format (one account pair per line):

```text
src_user,src_pass,dst_user,dst_pass
```

Notes:

- Lines starting with `#` are treated as comments.
- Empty lines are ignored.
- On validation success, migration starts immediately.

## How migration works

For each account pair:

1. Connect and authenticate to source and destination IMAP servers.
2. List source folders and process folders with a worker pool.
3. Ensure each folder exists on destination (`CREATE` if needed).
4. Preload destination `Message-ID` cache (single batched pass).
5. Fetch source metadata in batches (`UID`, `Message-ID`, size).
6. O(1) map lookup to skip duplicates before body transfer.
7. Transfer only required messages (`FETCH BODY[]` -> `APPEND`).
8. Emit live status/speed updates to the TUI via channel messages.

If a folder or account fails, the error is logged and migration continues with
the next item.

## Folder naming behavior

MailMole preserves server folder names exactly as returned by IMAP.

Example: if source uses `INBOX.Sent`, `INBOX.Drafts`, `INBOX.Archive`, those
exact names are created/copied on destination. This is normal IMAP behavior.

## Web Dashboard (Optional)

MailMole includes an optional web dashboard for monitoring and managing migrations from any browser.

### Usage

Run with web dashboard enabled:
```bash
# TUI + Web Dashboard
./mailmole -web :8080

# Web Dashboard only (no TUI)
./mailmole -web :8080 -web-only
```

Then open `http://localhost:8080` in your browser.

### Web Dashboard Features

#### 1. Connection Setup
- **Quick Templates**: Pre-configured settings for Gmail, Outlook, Yandex, iCloud, Yahoo
- **Single Account Mode**: Migrate one account at a time
- **Bulk Migration Mode**: Migrate multiple accounts simultaneously
- **Import/Export**: Load accounts from CSV/JSON files

#### 2. Migration Templates
Choose from 19 pre-configured migration templates:
- Same provider migrations (Gmail→Gmail, Outlook→Outlook, etc.)
- Cross-provider migrations (Gmail→Outlook, Outlook→Yandex, etc.)

#### 3. Connection Validation
- **Test Connection**: Quick connectivity check
- **Detailed Test**: Comprehensive validation with per-account results
- **Visual Status**: Success/error indicators with detailed error messages
- **Export Results**: Download validation reports as CSV

#### 4. Preview Mode
- Review all folders before migration
- View message counts and size estimates
- Select/deselect specific folders
- See estimated duration

#### 5. Real-time Monitoring
- Live progress tracking via Server-Sent Events (SSE)
- Toast notifications for immediate feedback
- Per-account migration status
- Transfer speed and ETA

#### 6. Scheduling
- Schedule migrations for later execution
- Repeat options: Daily, Weekly, Monthly
- View and manage scheduled jobs

#### 7. Activity Logs
- Real-time log streaming
- Filter by log level (INFO, WARN, ERROR)
- Download complete logs

### Ideal Use Cases

The web dashboard is perfect for:
- **Remote monitoring** from another device
- **Team collaboration** - Share migration progress
- **Headless servers** - Run without TUI
- **Mobile access** - Monitor from your phone
- **Bulk operations** - Manage hundreds of accounts

## Security notes

- Credentials are entered directly in the terminal UI.
- For TLS connections made with a raw IP address, certificate verification is
  relaxed only when MAILMOLE_ALLOW_INSECURE_IP_TLS=1 is set.
- For hostname-based connections, normal TLS hostname verification applies.
- **Web Dashboard**: No authentication by default - only run on trusted networks
  or use firewall rules to restrict access.
  If binding to a non-local address, set MAILMOLE_WEB_TOKEN and open the UI with
  ?token=<value> to authenticate API requests.

## Project structure

```text
main.go                  # App entrypoint
internal/ui/             # Bubble Tea model, update loop, rendering
internal/sync/           # Queue parser and migration engine
internal/imap/           # IMAP client wrapper and transfer operations
internal/web/            # Web dashboard (HTTP server, SSE, HTML frontend)
```

### Web Dashboard Features

The web dashboard (`internal/web/`) provides a modern browser-based interface:

- **Real-time Monitoring**: Server-Sent Events (SSE) for live updates
- **Migration Templates**: Gmail, Outlook, Yandex, iCloud, Yahoo presets
- **Import/Export**: CSV/JSON support for bulk account management
- **Connection Validation**: Detailed test results with visual feedback
- **Toast Notifications**: User-friendly status messages
- **Responsive Design**: Works on desktop and mobile devices
- **Scheduling**: Schedule migrations for later execution
- **Multi-language Support**: English and Turkish (extensible)

#### API Endpoints

- `GET /` - Web dashboard interface
- `POST /api/test-connection` - Test single account connection
- `POST /api/preview` - Get folder preview for single account
- `POST /api/bulk-preview` - Get folder preview for multiple accounts
- `POST /api/validate` - Validate account connections with detailed results
- `POST /api/schedule` - Schedule a migration job
- `GET /api/schedules` - List scheduled jobs
- `DELETE /api/schedule/delete?id=<job_id>` - Delete scheduled job
- `GET /ws` - SSE endpoint for real-time updates

## Multi-Language Support

MailMole supports multiple languages in the web dashboard:

- 🇬🇧 **English** (Default)
- 🇹🇷 **Turkish** (Türkçe)

Language is automatically detected from browser settings. To add a new language:

1. Create a new translation file in `internal/web/locales/`
2. Add the language code to the supported languages list
3. Restart the web dashboard

## Known limitations

- Web dashboard has no built-in authentication (run on trusted networks only)
- Some advanced IMAP features (e.g., custom flags) are not preserved during migration

## Status

This project is under active development. Interfaces and behavior may evolve as
new migration and reporting features are added.
