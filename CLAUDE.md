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
2. **Progress Tracking**: Real-time progress with upload speed calculation via status file updates
3. **Resumable Uploads**: Complete breakpoint resume capability with automatic slice status detection
4. **Quick Upload**: SHA1-based deduplication and instant uploads for duplicate files
5. **Multi-threading**: Concurrent chunk uploads within each CLI process
6. **Error Handling**: Fast-fail approach with clear error messages
7. **Upload Speed Display**: Real-time speed monitoring with weighted averaging algorithm
8. **Server Selection**: Dynamic server list from API with manual selection for sponsored users

### API Integration

The codebase integrates with 钛盘's REST API endpoints:
- Base URL: `https://tmplink-sec.vxtrans.com/api_v2`
- Token-based authentication: Users obtain token from web browser session
- Request format: `application/x-www-form-urlencoded` (form POST) as required by server
- File operations: `/file` endpoint for upload preparation  
- Upload servers: Dynamic server selection for optimal upload performance
- Slice upload: `/app/upload_slice` for chunked uploads
- Server enumeration: Real-time server list retrieval via `upload_request_select2`

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

**Dynamic Server Selection**
- Server list obtained from `upload_request_select2` API response
- Available servers include: Global, JP (Japan), CN (China), HD1/HD2 (High Definition), C2 (Netherlands)
- Each server has `title` (display name) and `url` (upload endpoint)
- GUI automatically refreshes server list after user authentication
- Sponsored users can manually override automatic server selection
- CLI supports forced server selection via `-upload-server` parameter

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
- Enhanced error handling with two-phase JSON parsing for better error messages

### User Permission System

The application implements a tiered permission system based on user sponsor status:

**Regular Users:**
- File upload and download functionality
- Basic settings and standard upload functionality
- Default automatic server selection
- Standard upload speed monitoring

**Sponsored Users (Premium Features):**
- All regular user features
- Advanced upload settings: chunk size and concurrency control
- Manual server selection from dynamic API-provided server list
- Quick upload toggle (enable/disable instant upload checks)
- Priority server access and selection

**Permission Enforcement:**
- Settings interface shows locked (🔒) options for non-sponsored users
- Locked settings display current values as read-only
- Dynamic UI adaptation based on user sponsor status
- Graceful fallback for unsupported operations

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
- Permission-based UI rendering (sponsor vs regular users)
- Dynamic server list management with real-time API updates
- Upload speed calculation with weighted averaging
- Keyboard navigation optimized for arrow keys only

**Enhanced UI Features:**
- Real-time upload speed display (MB/s) for both active and completed uploads
- Server selection interface for sponsored users with left/right navigation
- Quick upload toggle with space bar interaction
- Hidden file visibility toggle with 't' key
- Status bar optimization to ensure key commands remain visible
- Permission-aware settings interface with locked/unlocked indicators

### CLI Operating Modes

The tmplink-cli process automatically selects its operating mode based on the presence of the `-task-id` parameter:

#### CLI Mode (Interactive Use)
**Trigger**: No `-task-id` parameter provided
```bash
./tmplink-cli -file document.pdf  # Automatic CLI mode
```

**Features**:
- 🎯 **Real-time Progress Bar**: Visual progress display similar to wget/curl
- ⚡ **Speed Monitoring**: Current and average upload speed display
- ⏱️ **ETA Calculation**: Estimated time remaining
- 🎨 **Enhanced UI**: Emoji and color-enhanced user experience
- 📊 **Detailed Stats**: File size, total time, and completion summary

#### GUI Mode (Programmatic Use)
**Trigger**: `-task-id` parameter provided
```bash
./tmplink-cli -file document.pdf -task-id upload_123  # GUI mode
```

**Features**:
- 📄 **Status File Output**: Progress written to JSON status files
- 🔄 **Silent Operation**: Suitable for programmatic invocation
- 📡 **IPC Communication**: Communicates with GUI via status files

### CLI Parameters

The tmplink-cli process accepts the following command-line parameters:

**Required Parameters:**
- `-file`: Path to the file to upload

**Token Requirements:** API token must be provided through one of:
- `-set-token`: Pre-save token to configuration file
- `-token`: Provide token temporarily via command line

**Optional Configuration Parameters:**
- `-chunk-size`: Chunk size in MB (default: 3MB, range: 1-99MB)
- `-model`: File expiration period (default: saved value or 0=24 hours)
  - `0`: 24 hours (default)
  - `1`: 3 days
  - `2`: 7 days  
  - `99`: Permanent (no expiration)
- `-mr-id`: Directory ID (default: saved value or "0" for root directory)
- `-skip-upload`: Enable instant upload check (default: 1=enabled)
- `-upload-server`: Force specific upload server URL (optional, overrides API selection)
- `-server-name`: Upload server display name (optional, for display only)
- `-task-id`: Task identifier (default: auto-generated)
- `-status-file`: Status file path (default: auto-generated)
- `-debug`: Enable debug mode for detailed logging

**Configuration Management Parameters:**
- `-set-token`: Set and save API token to configuration file
- `-set-model`: Set and save default file expiration period
- `-set-mr-id`: Set and save default directory ID

**API Server Architecture:**
- **API Server**: Fixed at `https://tmplink-sec.vxtrans.com/api_v2` (hardcoded, not configurable)
- **Upload Servers**: Dynamic allocation via API, can be manually overridden with `-upload-server`

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
9. `mr_id` - Directory ID (default "0" for root directory)
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

## Documentation Structure

This project maintains comprehensive documentation in the following structure:

- **README.md**: Installation and basic usage guide for end users
- **docs/usage.md**: Detailed usage guide with comprehensive examples
- **docs/technical.md**: Technical documentation for developers and advanced users
- **docs/api.md**: API integration specifications and endpoints
- **docs/design.md**: System design philosophy and architecture decisions
- **CLAUDE.md**: Development context and guidelines for Claude Code

## Code Structure

### Error Handling
- All functions return errors using Go conventions
- User-friendly error messages in Chinese
- Fast-fail error handling for immediate feedback

### Concurrency
- Upload workers run in separate goroutines
- Mutex-protected shared state in uploader
- Channel-based task queue system

### Configuration
- JSON-based config file at `~/.tmplink_config.json`
- Default values: 3MB chunks, 5 concurrent uploads
- Settings persist between sessions
- Fast-fail error handling for quick error identification

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

### Testing Rules and Guidelines

**File Organization:**
- **ALL test files and temporary files MUST be placed in the `test/` directory**
- Test input files: `test/small_test.txt`, `test/medium_test.bin`, `test/large_test.bin`
- Status files: `test/upload_status_*.json` 
- Test logs: `test/test_*.log`
- Temporary data: `test/temp_*`

**Test File Naming Convention:**
- Test input files: `test/<size>_test.<ext>` (e.g., `test/small_test.txt`)
- Status files: `test/status_<test-id>.json` (e.g., `test/status_001.json`)
- Log files: `test/test_<feature>_<date>.log` (e.g., `test/test_upload_20241231.log`)

**CLI Testing Parameters:**
- Use `-status-file test/status_<test-id>.json` for all CLI tests
- Use unique task IDs: `test-<sequential-number>` (e.g., `test-001`, `test-002`)
- Clean up test status files after validation

### Test Environment
- Test files organized in `test/` directory
- Sample configuration files for development
- Upload test logs and status files
- Test data includes 10MB files with 1MB chunk testing

### Known Test Results
- **10MB file with 1MB chunks**: Successfully uploads with status 8 completion
- **500MB large file testing**: Successfully completed with retry logic fixes
- **mr_id parameter testing**: Confirmed "0" default prevents folder lookup errors
- **Status code validation**: Status 7(data=8) correctly identified as error, status 8 as success
- **API endpoint compatibility**: All endpoints tested with proper form POST requests
- **Upload speed calculation**: Verified weighted averaging algorithm accuracy
- **Server selection testing**: Dynamic server list retrieval from API confirmed
- **Permission system testing**: Sponsor vs regular user feature access validated
- **Token validation**: Enhanced error handling for expired/invalid tokens
- **Navigation improvements**: Arrow-key-only navigation and hidden file toggle verified

### Recent Bug Fixes and Improvements
- **Retry mechanism removal**: Removed all retry logic for faster error feedback and simplified code
- **Token validation enhancement**: Two-phase JSON parsing for better error messages when tokens expire
- **Auto-update functionality (2025-06-06)**: Implemented version checking and automatic update system
  - Added `version.json` with CLI and GUI version tracking
  - Created `internal/updater` package for version management
  - CLI supports `--version`, `--check-update`, `--auto-update` parameters
  - GUI supports `--version`, `--check-update`, `--auto-update` parameters
  - **Automatic startup check**: Programs automatically check for updates when starting normal operations
  - Background update checking with non-blocking goroutines
  - Silent failure handling to avoid interrupting user experience
  - Downloads from GitHub releases based on detected platform
  - Supports all platforms: Linux (32/64/ARM64), Windows (32/64), macOS (Intel/ARM64)
- **Upload speed implementation**: Added SpeedCalculator with weighted averaging for accurate speed monitoring
- **Server address handling**: Fixed GUI server selection to properly pass upload server addresses to CLI
- **Dynamic server lists**: Replaced hardcoded server lists with real-time API retrieval
- **Permission system**: Implemented sponsor-only features with graceful degradation for regular users
- **UI navigation**: Streamlined keyboard shortcuts and improved status bar visibility

### Critical Bug Fix (2025-05-25): GUI Upload Process Communication
**Problem**: GUI file selection showed "启动中" (starting) status indefinitely with no actual upload progress or network traffic.

**Root Cause**: Variable shadowing bug in CLI `uploadFile` function (line 317):
```go
// WRONG: This creates a new local variable, leaving outer uploadInfo as nil
uploadInfo, err := getUTokenOnly(ctx, config, sha1Hash, fileName, fileInfo.Size())

// CORRECT: This assigns to the already declared variable
uploadInfo, err = getUTokenOnly(ctx, config, sha1Hash, fileName, fileInfo.Size())
```

**Impact**: 
- CLI processes launched by GUI would crash with null pointer dereference
- Status files remained in "starting" state
- No actual file upload occurred despite appearing in task list

**Solution Applied**:
1. **Fixed Variable Shadowing**: Corrected assignment in both GUI and CLI code paths
2. **Enhanced Error Handling**: CLI now properly handles uploadInfo initialization 
3. **Improved Process Monitoring**: GUI checkProgress function already handled missing status files correctly

**Verification**:
- CLI independent mode: ✅ Works correctly
- GUI to CLI communication: ✅ Fixed and functional
- Process tracking: ✅ PID correctly recorded in status files
- Upload progress: ✅ Real-time progress monitoring restored

**Files Modified**:
- `cmd/tmplink-cli/main.go`: Lines 317, 328 - Fixed variable shadowing
- `internal/gui/tui/model.go`: Improved process startup coordination

This fix resolves the core issue where GUI users experienced phantom upload tasks that never actually uploaded files.

### Breakpoint Resume Implementation (2025-05-26)

**Complete Resume Capability**: The Go implementation now matches the JavaScript version's full breakpoint resume functionality.

**Key Implementation Details**:

1. **Slice Status Detection**: 
   - Parses complete API response including `total`, `wait`, and `next` fields
   - Automatically calculates uploaded vs pending slices
   - Initializes resume state only once per upload session

2. **Progress Recovery Algorithm**:
   ```go
   // Resume detection and initialization
   if !resumeTracker.initialized && uploadedSlices > 0 {
       estimatedBytes := int64(uploadedSlices) * int64(chunkSize)
       progressCallback(estimatedBytes, fileSize) // Restore progress display
       resumeTracker.initialized = true
   }
   ```

3. **Enhanced Progress Calculation**:
   - Accounts for already-uploaded slices in progress display
   - Prevents progress from restarting at 0% during resume
   - Uses simplified slice-based calculation for accuracy

**API Response Handling**:
- **Status 3**: Enhanced to parse `total` and `wait` fields for resume detection
- **Resume Tracker**: New `ResumeTracker` struct prevents duplicate initialization
- **Debug Output**: Comprehensive logging for resume verification

**Testing**: 
- Use `./test_resume.sh` to verify resume functionality
- Supports interruption and restart scenarios
- Compatible with all chunk sizes (1MB-99MB)

**Compatibility**: Fully backward compatible with existing upload flows, resume detection is transparent to users.