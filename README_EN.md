# Minas Project Documentation
[简体中文](README.md) | [English](README_EN.md)

## Project Introduction

Minas is a multifunctional system management tool focused on providing powerful data management and automated operations capabilities. Its core features include file backup system, log and backup file cleaning, scheduled task management, WebDAV service, and super terminal management. It adopts a front-end and back-end separation architecture, with the backend developed in Go and the frontend developed using Vue 3 + TypeScript.

## Features

- **Data File Backup**: Supports various backup modes (incremental backup, one-way mirror, full backup), can be scheduled to execute automatically to protect important data
- **Log & Data Cleaning**: Intelligently manages system logs and backup data, automatically cleans expired files according to time, quantity, and other rules to free up storage space
- **Scheduled Script Tasks**: Powerful task scheduling system supporting cron expressions, can execute custom scripts and system commands
- **WebDAV Service**: Provides WebDAV protocol support for convenient file access and management
- **Super Terminal Management**: Provides Web terminal functionality for remote command execution
- **Multi-platform Support**: Supports multiple operating systems including Linux and Windows
- **Docker Support**: Provides Docker images for easy deployment

## Developer Guide

### Environment Requirements

#### Backend Development Environment

- Go 1.19 or higher
- Supported databases: SQLite, MySQL, PostgreSQL
- rclone tool (for file backup functionality)

#### Frontend Development Environment

- Node.js 14 or higher
- npm or yarn package manager

### Project Structure

```
minas/
├── build.sh                # Build script
├── build-docker.sh         # Docker build script
├── Dockerfile              # Docker configuration file
├── server/                 # Backend code
│   ├── app/                # Application layer code
│   ├── cmd/                # Command line tools
│   ├── core/               # Core code
│   ├── data/               # Data and configuration
│   ├── middleware/         # Middleware
│   ├── route/              # Routes
│   ├── service/            # Service layer
│   ├── utils/              # Utility classes
│   ├── www/                # Compiled frontend files
│   ├── go.mod              # Go module dependencies
│   └── main.go             # Main entry
├── web/                    # Frontend code
│   ├── public/             # Static resources
│   ├── src/                # Source code
│   ├── index.html          # HTML entry
│   ├── package.json        # Dependencies configuration
│   └── vite.config.ts      # Vite configuration
└── airgo.sh                # Development mode startup script
```

### Setting Up Development Environment

#### Clone Code

```bash
git clone git@github.com:sorchk/minas.git
cd minas
```

#### Backend Development

1. Install Go dependencies

```bash
cd server
go mod download
```

2. Install rclone tool (for file backup functionality)

```bash
# Linux
curl https://rclone.org/install.sh | sudo bash

# Or download and install manually
# Download link: https://rclone.org/downloads/
```

3. Install air tool (for hot reload)

Air is a hot reload tool that automatically recompiles and runs your program when code changes, making it ideal for development environments.

```bash
# Install using go install
go install github.com/air-verse/air@latest

# Make sure $GOPATH/bin is in your PATH environment variable
# For Linux/macOS, you can add to ~/.bashrc or ~/.zshrc
# export PATH=$PATH:$GOPATH/bin

# Verify installation
air -v
```

Or install using alternative methods:

```bash
# Install using curl
# Linux/macOS
curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

4. Start in development mode

```bash
# Use air tool for hot reload
cd ..
chmod +x *.sh
./airgo.sh
```

#### Frontend Development

1. Install dependencies

```bash
cd web
npm install
# or
yarn
```

2. Start in development mode

```bash
npm run dev
# or
yarn dev
```

The frontend development server will start at http://localhost:3002 and automatically proxy API requests to the backend service.

### Building the Project

Use the provided build script to build the entire project with one command:

```bash
./build.sh
```

This will:
1. Build the frontend code and place it in the `server/www/dist` directory
2. Build the backend code, generating multi-platform executable files in the `dist` directory

### Docker Build

```bash
./build-docker.sh [version]
```

If no version is specified, the default version 1.3.3 will be used.

### Custom Development

#### Modifying Configuration

The main configuration file is located at `server/data/config.yaml`, which can be modified as needed.

#### Adding New Features

1. Create a new functional module in the `server/app` directory
2. Register new routes in `server/route/Index.go`
3. Add corresponding frontend code in `web/src`

## User Guide

### Installation Methods

#### Binary Installation

1. Download the binary file suitable for your system from the [releases page](https://github.com/sorchk/minas/releases)
2. Extract the file
3. Run the executable file

```bash
# Linux
chmod +x minas_amd64  # or minas_arm64
./minas_amd64 server

# Windows
minas.exe server
```

#### Docker Installation

```bash
docker run -d --name minas \
  -p 8002:8002 \
  -p 8003:8003 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/backup:/backup \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /usr/bin/docker:/usr/bin/docker \
  sorc/minas:latest
```

### Configuration Description

When running for the first time, the system will generate a default configuration file `config.yaml` in the `data` directory. Main configuration items:

```yaml
app:
  host: 0.0.0.0            # Listening address
  port: 8002               # HTTP port
  context-path: "/minas"   # Context path
  ssl-port: 8003           # HTTPS port
  enable-ui: true          # Whether to enable UI
  type: "all"              # Service type: all or webdav
term:
  command: "bash"          # Terminal command
db:
  type: sqlite             # Database type: sqlite, mysql, postgres
  # Other database configurations...
rclone:
  bin-path: "rclone"       # Path to rclone executable
  port: 5572               # rclone service port
log:
  level: "debug"           # Log level
  log-path: "logs"         # Log path
  # Other log configurations...
```

### Service Management

Minas provides various service management commands:

```bash
# Start service
minas server

# Install as system service
minas install

# Start system service
minas start

# Stop system service
minas stop

# Uninstall system service
minas uninstall

# View version
minas -v
```

### Feature Usage

#### Initialization

When accessing the system for the first time, you need to perform initialization settings to create an administrator account.

#### Data Backup and Synchronization

1. Go to the "Scheduled Tasks" page, select the "File Synchronization" function
2. Create a new backup/synchronization task
3. Select the synchronization type:
   - **Backup Mode**: One-way file copying, suitable for data backup
   - **Mirror Mode**: Make the target directory exactly the same as the source directory, will delete files in the target that don't exist in the source
   - **Two-way Sync**: Keep two directories consistent, changes in either directory will be synchronized to the other
4. Set source path and target path (supports local paths, remote FTP/SFTP, cloud storage, etc.)
5. Configure advanced options (such as file filtering rules, bandwidth limits, etc.)
6. Set execution schedule (daily, weekly, monthly, or custom cron expression)
7. Save and activate the task

#### Log and Data Cleaning

1. Go to the "Scheduled Tasks" page, select the "File Cleaning" function
2. Create a new cleaning task
3. Set the target directory for cleaning
4. Configure cleaning rules:
   - Clean by time (e.g., delete files older than 30 days)
   - Clean by number (e.g., keep only the 10 most recent backup files)
   - Clean by file size (e.g., clean the oldest files when the directory exceeds 10GB)
   - Clean by filename pattern (supports wildcards and regular expressions)
5. Set execution schedule
6. Optionally enable logging and cleaning reports
7. Save and activate the task

#### Scheduled Script Tasks

1. Go to the "Scheduled Tasks" page, select the "Script Tasks" function
2. Create a new script task
3. Enter task name and description
4. Write or upload script content to execute (supports shell, python, etc.)
5. Set execution parameters and environment variables
6. Configure execution schedule (supports cron expressions)
7. Set failure retry strategy and timeout
8. Configure task completion notification (optional)
9. Save and activate the task

#### WebDAV Service

1. After logging in to the system, go to the "WebDAV Management" page
2. Click "New" to create a WebDAV account
3. Set account name, home directory, and permissions
4. After saving, you can connect using a WebDAV client:
   - Address: `http(s)://server-address:port/minas/dav`
   - Username: Created account
   - Password: System-generated token


#### Super Terminal

1. Go to the "Terminal" page
2. The system will automatically open a terminal session
3. You can execute command-line operations

### Common Issues

1. **Unable to start service**
   - Check if the port is occupied
   - Check if the configuration file is correct
   - View log files for detailed error information

2. **WebDAV connection failure**
   - Confirm WebDAV account status is enabled
   - Check if username and token are correct
   - Confirm network connection and firewall settings

3. **File backup/synchronization failure**
   - Check if rclone is correctly installed
   - Confirm source and target paths have access permissions
   - Verify network connection and remote storage configuration
   - Check task logs for detailed error information

4. **Scheduled task not executing**
   - Check if system time is correct
   - Confirm cron expression syntax is correct
   - Check if task status is "Enabled"
   - View task execution logs

5. **Log cleaning not working**
   - Confirm cleaning rule configuration is correct
   - Check target directory permissions
   - Verify file matching rules correctly match target files

## License

This project is licensed under the MIT License. See the LICENSE file for details.
