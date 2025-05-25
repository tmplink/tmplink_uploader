# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a dual-process uploader tool for [钛盘](https://tmp.link/), written in Go. The project consists of a GUI process and independent CLI processes that handle individual file uploads with inter-process communication via JSON status files.

## Architecture

### Core Components

**GUI Process (cmd/tmplink/):**
- **main.go**: GUI entry point with Bubble Tea TUI framework
- **internal/gui/tui/model.go**: Terminal User Interface implementation
- File selection, progress monitoring, and CLI process management

**CLI Process (cmd/tmplink-cli/):**
- **main.go**: Independent CLI uploader for single files
- No config files, all parameters via command line
- One-time execution, exits after upload completion
- Status communication via JSON files

**Reference Files:**
- **reference/tmplink.js & uploader.js**: JavaScript reference implementation for API analysis

### Key Features

**Dual-Process Architecture:**
1. **GUI Process**: Bubble Tea TUI for file selection and progress monitoring
2. **CLI Processes**: Independent upload processes for each file
3. **File-based Communication**: JSON status files for inter-process communication
4. **Process Independence**: CLI processes are completely self-contained

**Upload Features:**
1. **Chunked Upload**: Configurable chunk size (default 3MB, max 80MB)
2. **Progress Tracking**: Real-time progress via status file updates
3. **Resumable Uploads**: SHA1-based deduplication and instant uploads
4. **Multi-threading**: Concurrent chunk uploads within each CLI process
5. **Error Handling**: Automatic retry with configurable retry count

### API Integration

The codebase integrates with 钛盘's REST API endpoints:
- Base URL: `https://tmplink-sec.vxtrans.com/api_v2`
- Token-based authentication: Users obtain token from web browser session
- Request format: `application/x-www-form-urlencoded` (form POST) as required by server
- File operations: `/file` endpoint for upload preparation  
- Upload servers: Dynamic server selection for optimal upload performance
- Slice upload: `/app/upload_slice` for chunked uploads

#### Key API Endpoints

**Upload Token Request (`upload_request_select2`)**
- Endpoint: `POST /api_v2/file`
- Action: `upload_request_select2`
- **Important**: No captcha verification required for this endpoint
- Purpose: Get upload token (utoken) and available upload servers
- Parameters: `action`, `token`, `filesize` (captcha field can be empty)
- Returns: `{status: 1, data: {utoken: "...", servers: [...]}}`

**Slice Upload Preparation**
- Endpoint: `POST {server_url}/app/upload_slice`
- Action: `prepare`
- Purpose: Query which slices need uploading and handle upload state machine
- Status codes: 1=complete, 2=wait, 3=upload_slice, 7=error, 9=reset

### Authentication Flow

Instead of traditional login, the application uses a token-based approach:
1. User visits https://tmp.link/ and logs in via web browser
2. User obtains API token from browser localStorage
3. Token is entered into CLI application and saved to config
4. Token is validated against the API before enabling features
5. Token is re-validated before critical operations (uploads)
6. If token becomes invalid, user is prompted to re-enter token

### Token Validation

The application includes comprehensive token validation:
- Initial validation on startup using `/user` endpoint with `get_detail` action
- Uses form POST requests (`application/x-www-form-urlencoded`) matching web interface
- Re-validation before upload operations to ensure token is still valid
- Automatic re-prompting if token expires or becomes invalid
- User information (email, UID, sponsor status) is retrieved during validation
- Detailed error reporting including debug information from API

### TUI Architecture

The terminal user interface is built with bubbletea following the Elm architecture:

**State Management:**
- `TUIModel` struct contains all UI state and components
- States: `stateLoading`, `stateMain`, `stateSettings`, `stateFileSelect`, `stateUploadList`, etc.
- Navigation stack for proper back button functionality

**UI Components:**
- `list.Model` - Menu navigation and file selection
- `textinput.Model` - Number input for settings (chunk size, concurrency)
- `viewport.Model` - Scrollable content display for upload lists
- `progress.Model` - Progress bars for upload status
- `spinner.Model` - Loading indicators

**Key Patterns:**
- Custom `itemDelegate` for styled list items
- State-driven rendering with dedicated render functions
- Component focus management and keyboard handling
- Responsive sizing that adapts to terminal dimensions

### CLI Parameters

The tmplink-cli process accepts the following command-line parameters:

**Required Parameters (Minimal Set):**
- `-file`: Path to the file to upload
- `-token`: 钛盘 API token  
- `-task-id`: Unique task identifier
- `-status-file`: Path to JSON status file for progress communication

**Optional Configuration Parameters:**
- `-server`: Upload server URL (default: https://tmplink-sec.vxtrans.com/api_v2)
- `-chunk-size`: Chunk size in bytes (default: 3MB, max: 80MB)
- `-max-retries`: Maximum retry attempts (default: 3)
- `-timeout`: Request timeout in seconds (default: 300)
- `-model`: File expiration period (default: 0, 24 hours)
  - `0`: 24 hours (default)
  - `1`: 3 days
  - `2`: 7 days  
  - `99`: Permanent (no expiration)
- `-mr-id`: Resource ID (default: "0" for root directory, for specific upload contexts)
- `-skip-upload`: Skip upload flag (default: 1, enables instant upload check)
- `-uid`: User ID (optional, auto-obtained from token validation if not provided)

**API Upload Parameters:**

The CLI generates API calls with the following complete parameter set for upload operations:

1. `token` - 钛盘 API token
2. `uptoken` - Client-generated SHA1(uid + filename + filesize + chunk_size)  
3. `action` - API action ("prepare" for chunk queries, "upload_slice" for data upload)
4. `sha1` - File SHA1 hash for deduplication
5. `filename` - Original filename
6. `filesize` - Total file size in bytes
7. `slice_size` - Chunk size in bytes
8. `utoken` - Server-provided upload token from upload_request_select2
9. `mr_id` - Resource ID (default "0" for root directory)
10. `model` - File expiration period (0 = 24 hours, 1 = 3 days, 2 = 7 days, 99 = permanent)

**Internal Parameter Processing:**

The CLI automatically handles the following internally:
- `sha1`: Calculated from file content
- `filename`: Extracted from file path
- `filesize`: Retrieved from file stats
- `utoken`: Obtained from upload_request_select2 API response
- `uptoken`: Generated using SHA1(uid + filename + filesize + chunk_size)
- `uid`: Auto-obtained from token validation (/user API endpoint) if not provided

**Design Philosophy:**
- **Minimal user input**: Only 4 required parameters for basic operation
- **Auto-discovery**: CLI handles token validation, file analysis, and API tokens internally
- **Reasonable defaults**: All configuration parameters have sensible defaults
- **No config files**: All settings passed via command line for process independence

**Note:** 
- `captcha` parameter is NOT used as 钛盘 API does not require captcha verification
- `mr-id` parameter defaults to "0" (root directory) and is always included in API calls
- Users don't need to know internal API details like utoken/uptoken generation

**Critical Bug Fix (2024-12-31):**
- Fixed mr_id parameter default value from empty string "" to "0"
- Empty mr_id caused status 7(data=8) "folder not found" errors
- Proper mr_id="0" now returns status 8 with successful merge completion

## Development Commands

### Build and Run
- `make build` - Build for current platform
- `make release` - Build for all platforms (Linux, Windows, macOS)
- `make run` - Run directly with go run
- `make dist` - Create distribution packages

### Development
- `make deps` - Install dependencies
- `make fmt` - Format code
- `make vet` - Run go vet
- `make test` - Run tests
- `make clean` - Clean build artifacts

### Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/charmbracelet/bubbles` - TUI components (list, textinput, viewport, progress)
- `github.com/charmbracelet/lipgloss` - TUI styling
- `github.com/mattn/go-runewidth` - Unicode width calculation for Chinese characters
- `github.com/schollz/progressbar/v3` - Progress bars

## Code Structure

### Error Handling
- All functions return errors using Go conventions
- User-friendly error messages in Chinese
- Retry logic for network failures

### Concurrency
- Upload workers run in separate goroutines
- Mutex-protected shared state in uploader
- Channel-based task queue system

### Configuration
- JSON-based config file at `~/.tmplink_config.json`
- Default values: 3MB chunks, 5 concurrent uploads, 300s upload timeout
- Settings persist between sessions
- Configurable upload timeout for large files (default 5 minutes)
- Automatic retry mechanism with exponential backoff (max 3 retries)

### Security
- Token-based authentication
- HTTPS for all API calls
- File validation and size limits
- No sensitive data in logs

## API Response Handling

### Upload Preparation (`prepare_v4`)
The `prepare_v4` API has flexible response structure:
- **Status 1**: Quick upload successful, `data` contains `{ukey: "..."}` object
- **Status 0**: Need chunk upload, `data` is `false` (not an error)
- **Other status**: Actual errors with message details

### Upload Status Codes
- Status 1: Upload completed successfully
- Status 2: Waiting for other chunks  
- Status 3: Ready to upload next chunk
- Status 6: File already exists (instant upload)
- Status 7: Upload error - data field contains error code (e.g., data=8 means "folder not found")
- Status 8: Upload merge completed successfully (normal completion after chunked upload)

### File Preparation
- SHA1 calculation for deduplication
- Chunk size validation (1-80MB)
- Server availability checking
- Upload token generation

### Response Structure Handling
The Go client handles dynamic API responses:
- `PrepareResponse.Data` uses `interface{}` to handle both object and boolean values
- Type assertion is used to safely extract specific fields
- Fallback logic ensures robust error handling for unexpected response formats

## Testing and Validation

### Test Environment
- Test files organized in `test/` directory
- Sample configuration files for development
- Upload test logs and status files
- Test data includes 10MB files with 1MB chunk testing

### Known Test Results
- **10MB file with 1MB chunks**: Successfully uploads with status 8 completion
- **mr_id parameter testing**: Confirmed "0" default prevents folder lookup errors
- **Status code validation**: Status 7(data=8) correctly identified as error, status 8 as success
- **API endpoint compatibility**: All endpoints tested with proper form POST requests