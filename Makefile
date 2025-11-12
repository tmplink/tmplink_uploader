# 变量定义
BINARY_DIR = ./
CLI_BINARY = tmplink-cli
GUI_BINARY = tmplink
CLI_SOURCE = ./cmd/tmplink-cli
GUI_SOURCE = ./cmd/tmplink
GO_VERSION = $(shell go version | awk '{print $$3}')
BUILD_TIME = $(shell date +"%Y-%m-%d %H:%M:%S")
GIT_COMMIT = $(shell git rev-parse --short HEAD)
BUILD_DIR = ./build

# 从version.json读取版本号
CLI_VERSION = $(shell grep -o '"cli_version"[[:space:]]*:[[:space:]]*"[^"]*"' version.json | grep -o '"[^"]*"$$' | tr -d '"')
GUI_VERSION = $(shell grep -o '"gui_version"[[:space:]]*:[[:space:]]*"[^"]*"' version.json | grep -o '"[^"]*"$$' | tr -d '"')

# 构建标记
CLI_LDFLAGS = -ldflags "-X main.Version=$(CLI_VERSION) -X 'main.BuildTime=$(BUILD_TIME)' -X main.GitCommit=$(GIT_COMMIT)"
GUI_LDFLAGS = -ldflags "-X main.Version=$(GUI_VERSION) -X 'main.BuildTime=$(BUILD_TIME)' -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: all build clean test run help deps lint vet fmt check install build-all build-release release

# 默认目标
all: build

# 构建当前平台版本
build: build-cli build-gui

# 构建所有平台发布版本
release: build-release

# 构建CLI程序
build-cli:
	@echo "构建CLI程序..."
	go build $(CLI_LDFLAGS) -o $(BINARY_DIR)$(CLI_BINARY) $(CLI_SOURCE)

# 构建静态链接的 Linux CLI 程序（兼容旧版 GLIBC）
build-cli-linux-static:
	@echo "构建静态链接的 Linux CLI 程序..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(CLI_LDFLAGS) -o $(BINARY_DIR)$(CLI_BINARY)-linux-amd64-static $(CLI_SOURCE)
	@echo "完成！二进制文件: $(BINARY_DIR)$(CLI_BINARY)-linux-amd64-static"

# 构建GUI程序  
build-gui:
	@echo "构建GUI程序..."
	go build $(GUI_LDFLAGS) -o $(BINARY_DIR)$(GUI_BINARY) $(GUI_SOURCE)

# 构建所有平台发布版本
build-release: clean-build
	@echo "构建所有平台发布版本..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "构建 macOS Intel..."
	GOOS=darwin GOARCH=amd64 go build $(GUI_LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY)-darwin-amd64 $(GUI_SOURCE)
	GOOS=darwin GOARCH=amd64 go build $(CLI_LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY)-darwin-amd64 $(CLI_SOURCE)
	
	@echo "构建 macOS ARM64..."
	GOOS=darwin GOARCH=arm64 go build $(GUI_LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY)-darwin-arm64 $(GUI_SOURCE)
	GOOS=darwin GOARCH=arm64 go build $(CLI_LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY)-darwin-arm64 $(CLI_SOURCE)
	
	@echo "构建 Windows 64位..."
	GOOS=windows GOARCH=amd64 go build $(GUI_LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY)-windows-amd64.exe $(GUI_SOURCE)
	GOOS=windows GOARCH=amd64 go build $(CLI_LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY)-windows-amd64.exe $(CLI_SOURCE)
	
	@echo "构建 Windows 32位..."
	GOOS=windows GOARCH=386 go build $(GUI_LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY)-windows-386.exe $(GUI_SOURCE)
	GOOS=windows GOARCH=386 go build $(CLI_LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY)-windows-386.exe $(CLI_SOURCE)
	
	@echo "构建 Linux 64位..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GUI_LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY)-linux-amd64 $(GUI_SOURCE)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(CLI_LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY)-linux-amd64 $(CLI_SOURCE)
	
	@echo "构建 Linux 32位..."
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build $(GUI_LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY)-linux-386 $(GUI_SOURCE)
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build $(CLI_LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY)-linux-386 $(CLI_SOURCE)
	
	@echo "构建 Linux ARM64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GUI_LDFLAGS) -o $(BUILD_DIR)/$(GUI_BINARY)-linux-arm64 $(GUI_SOURCE)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(CLI_LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY)-linux-arm64 $(CLI_SOURCE)
	
	@echo "所有平台构建完成！"

# 兼容旧的build-all命令
build-all: build-release

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod download
	go mod tidy

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "运行测试并生成覆盖率..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成: coverage.html"

# 运行CLI程序
run-cli: build-cli
	@echo "运行CLI程序..."
	./$(CLI_BINARY) -h

# 运行GUI程序  
run-gui: build-gui
	@echo "运行GUI程序..."
	./$(GUI_BINARY)

# 直接运行（开发模式）
run:
	@echo "运行GUI程序（开发模式）..."
	go run $(GUI_SOURCE)

# 代码格式化
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
vet:
	@echo "运行go vet..."
	go vet ./...

# 代码规范检查
lint:
	@echo "运行golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint未安装，跳过检查"; \
	fi

# 综合检查
check: fmt vet lint test

# 清理构建文件
clean:
	@echo "清理本地构建文件..."
	rm -f $(CLI_BINARY)
	rm -f $(GUI_BINARY)
	rm -f coverage.out
	rm -f coverage.html
	rm -f test-tmplink-cli
	rm -f test-tmplink-gui
	rm -rf dist/

# 清理发布构建文件
clean-build:
	@echo "清理发布构建文件..."
	rm -f $(BUILD_DIR)/$(CLI_BINARY)-*
	rm -f $(BUILD_DIR)/$(GUI_BINARY)-*

# 安装到系统
install: build
	@echo "安装程序到系统..."
	sudo cp $(CLI_BINARY) /usr/local/bin/
	sudo cp $(GUI_BINARY) /usr/local/bin/

# 卸载
uninstall:
	@echo "从系统卸载..."
	sudo rm -f /usr/local/bin/$(CLI_BINARY)
	sudo rm -f /usr/local/bin/$(GUI_BINARY)

# 创建发布包
dist: build-release
	@echo "创建发布包..."
	@mkdir -p $(BUILD_DIR)/packages
	@echo "打包 Darwin AMD64..."
	@tar -czf $(BUILD_DIR)/packages/tmplink-darwin-amd64.tar.gz -C $(BUILD_DIR) $(GUI_BINARY)-darwin-amd64 $(CLI_BINARY)-darwin-amd64
	@echo "打包 Darwin ARM64..."
	@tar -czf $(BUILD_DIR)/packages/tmplink-darwin-arm64.tar.gz -C $(BUILD_DIR) $(GUI_BINARY)-darwin-arm64 $(CLI_BINARY)-darwin-arm64
	@echo "打包 Windows AMD64..."
	@tar -czf $(BUILD_DIR)/packages/tmplink-windows-amd64.tar.gz -C $(BUILD_DIR) $(GUI_BINARY)-windows-amd64.exe $(CLI_BINARY)-windows-amd64.exe
	@echo "打包 Windows 386..."
	@tar -czf $(BUILD_DIR)/packages/tmplink-windows-386.tar.gz -C $(BUILD_DIR) $(GUI_BINARY)-windows-386.exe $(CLI_BINARY)-windows-386.exe
	@echo "打包 Linux AMD64..."
	@tar -czf $(BUILD_DIR)/packages/tmplink-linux-amd64.tar.gz -C $(BUILD_DIR) $(GUI_BINARY)-linux-amd64 $(CLI_BINARY)-linux-amd64
	@echo "打包 Linux 386..."
	@tar -czf $(BUILD_DIR)/packages/tmplink-linux-386.tar.gz -C $(BUILD_DIR) $(GUI_BINARY)-linux-386 $(CLI_BINARY)-linux-386
	@echo "打包 Linux ARM64..."
	@tar -czf $(BUILD_DIR)/packages/tmplink-linux-arm64.tar.gz -C $(BUILD_DIR) $(GUI_BINARY)-linux-arm64 $(CLI_BINARY)-linux-arm64
	@echo "发布包创建完成，位于 $(BUILD_DIR)/packages/ 目录"

# 开发环境设置
dev-setup:
	@echo "设置开发环境..."
	@echo "安装开发依赖..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "开发环境设置完成"

# 显示构建状态
status:
	@echo "构建状态检查..."
	@echo "本地构建文件:"
	@ls -la $(CLI_BINARY) $(GUI_BINARY) 2>/dev/null || echo "  未找到本地构建文件"
	@echo ""
	@echo "发布构建文件:"
	@ls -la $(BUILD_DIR)/$(CLI_BINARY)-* $(BUILD_DIR)/$(GUI_BINARY)-* 2>/dev/null | head -10 || echo "  未找到发布构建文件"

# 显示帮助信息
help:
	@echo "钛盘上传工具构建系统"
	@echo ""
	@echo "构建命令:"
	@echo "  build       - 构建当前平台版本"
	@echo "  build-cli   - 只构建CLI程序"
	@echo "  build-cli-linux-static - 构建静态链接的Linux CLI程序（兼容旧版GLIBC）"
	@echo "  build-gui   - 只构建GUI程序"
	@echo "  run         - 运行GUI程序（开发模式）"
	@echo "  run-cli     - 运行CLI程序"
	@echo "  run-gui     - 运行GUI程序"
	@echo ""
	@echo "发布命令:"
	@echo "  release     - 构建所有平台发布版本"
	@echo "  build-release - 同 release"
	@echo "  build-all   - 同 release"
	@echo "  dist        - 创建发布包"
	@echo ""
	@echo "测试命令:"
	@echo "  test        - 运行测试"
	@echo "  test-coverage - 运行测试并生成覆盖率报告"
	@echo "  fmt         - 格式化代码"
	@echo "  vet         - 运行go vet"
	@echo "  lint        - 运行golangci-lint"
	@echo "  check       - 运行所有检查"
	@echo ""
	@echo "管理命令:"
	@echo "  deps        - 安装依赖"
	@echo "  clean       - 清理本地构建文件"
	@echo "  clean-build - 清理发布构建文件"
	@echo "  install     - 安装到系统"
	@echo "  uninstall   - 从系统卸载"
	@echo "  status      - 显示构建状态"
	@echo "  dev-setup   - 设置开发环境"
	@echo "  help        - 显示此帮助信息"
	@echo ""
	@echo "构建信息:"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
