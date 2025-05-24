# 变量定义
BINARY_DIR = ./
CLI_BINARY = tmplink-cli
GUI_BINARY = tmplink
CLI_SOURCE = ./cmd/tmplink-cli
GUI_SOURCE = ./cmd/tmplink
GO_VERSION = $(shell go version | awk '{print $$3}')
BUILD_TIME = $(shell date +"%Y-%m-%d %H:%M:%S")
GIT_COMMIT = $(shell git rev-parse --short HEAD)

# 构建标记
LDFLAGS = -ldflags "-X main.Version=$(GIT_COMMIT) -X 'main.BuildTime=$(BUILD_TIME)'"

.PHONY: all build clean test run help deps lint vet fmt check install build-all

# 默认目标
all: build

# 构建所有可执行文件
build: build-cli build-gui

# 构建CLI程序
build-cli:
	@echo "构建CLI程序..."
	go build $(LDFLAGS) -o $(BINARY_DIR)$(CLI_BINARY) $(CLI_SOURCE)

# 构建GUI程序  
build-gui:
	@echo "构建GUI程序..."
	go build $(LDFLAGS) -o $(BINARY_DIR)$(GUI_BINARY) $(GUI_SOURCE)

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
	@echo "清理构建文件..."
	rm -f $(CLI_BINARY)
	rm -f $(GUI_BINARY)
	rm -f coverage.out
	rm -f coverage.html
	rm -f test-tmplink-cli
	rm -f test-tmplink-gui
	rm -rf dist/

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

# 构建所有平台
build-all: build-linux build-windows build-darwin

# 构建Linux版本
build-linux:
	@echo "构建Linux版本..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/linux/$(CLI_BINARY) $(CLI_SOURCE)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/linux/$(GUI_BINARY) $(GUI_SOURCE)

# 构建Windows版本
build-windows:
	@echo "构建Windows版本..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/windows/$(CLI_BINARY).exe $(CLI_SOURCE)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/windows/$(GUI_BINARY).exe $(GUI_SOURCE)

# 构建macOS版本
build-darwin:
	@echo "构建macOS版本..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/darwin/$(CLI_BINARY) $(CLI_SOURCE)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/darwin/$(GUI_BINARY) $(GUI_SOURCE)

# 创建发布包
dist: clean build-all
	@echo "创建发布包..."
	@mkdir -p dist
	@for os in linux windows darwin; do \
		echo "打包 $$os..."; \
		cd dist/$$os && tar -czf ../tmplink-$$os-amd64.tar.gz *; \
		cd ../..; \
	done

# 开发环境设置
dev-setup:
	@echo "设置开发环境..."
	@echo "安装开发依赖..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "开发环境设置完成"

# 显示帮助信息
help:
	@echo "TmpLink Uploader 构建系统"
	@echo ""
	@echo "可用命令:"
	@echo "  build       - 构建所有可执行文件"
	@echo "  build-cli   - 只构建CLI程序"
	@echo "  build-gui   - 只构建GUI程序"
	@echo "  deps        - 安装依赖"
	@echo "  test        - 运行测试"
	@echo "  test-coverage - 运行测试并生成覆盖率报告"
	@echo "  run         - 运行GUI程序（开发模式）"
	@echo "  run-cli     - 运行CLI程序"
	@echo "  run-gui     - 运行GUI程序"
	@echo "  fmt         - 格式化代码"
	@echo "  vet         - 运行go vet"
	@echo "  lint        - 运行golangci-lint"
	@echo "  check       - 运行所有检查"
	@echo "  clean       - 清理构建文件"
	@echo "  install     - 安装到系统"
	@echo "  uninstall   - 从系统卸载"
	@echo "  build-all   - 构建所有平台版本"
	@echo "  dist        - 创建发布包"
	@echo "  dev-setup   - 设置开发环境"
	@echo "  help        - 显示此帮助信息"
	@echo ""
	@echo "构建信息:"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"