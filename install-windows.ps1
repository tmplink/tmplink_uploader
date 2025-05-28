# 钛盘上传工具 Windows 安装脚本
# TmpLink Uploader Windows Installation Script

param(
    [switch]$Force,
    [string]$InstallPath = "$env:ProgramFiles\TmpLink",
    [switch]$NoStartMenu,
    [switch]$NoDesktop,
    [switch]$NoPath
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 脚本参数
$GitHubRepo = "tmplink/tmplink_uploader"
$DownloadBase = "https://raw.githubusercontent.com/$GitHubRepo/main/build"

# 颜色函数
function Write-Header {
    Write-Host "=====================================" -ForegroundColor Blue
    Write-Host "   钛盘上传工具 Windows 安装程序    " -ForegroundColor Blue
    Write-Host "  TmpLink Uploader Windows Installer" -ForegroundColor Blue
    Write-Host "=====================================" -ForegroundColor Blue
    Write-Host ""
}

function Write-Step {
    param([string]$Message)
    Write-Host "[步骤] $Message" -ForegroundColor Yellow
}

function Write-Success {
    param([string]$Message)
    Write-Host "[成功] $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "[错误] $Message" -ForegroundColor Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "[信息] $Message" -ForegroundColor Blue
}

function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Test-Requirements {
    Write-Step "检查系统要求..."
    
    # 检查是否为 Windows
    if ($env:OS -ne "Windows_NT") {
        Write-Error "此脚本仅适用于 Windows 系统"
        exit 1
    }
    
    # 显示 Windows 版本信息
    $osInfo = Get-CimInstance -ClassName Win32_OperatingSystem
    Write-Info "检测到系统: $($osInfo.Caption) $($osInfo.Version)"
    
    # 检查管理员权限
    if (-not (Test-Administrator)) {
        if (-not $Force) {
            Write-Info "建议以管理员身份运行此脚本以获得最佳体验"
            $response = Read-Host "是否继续？ [y/N]"
            if ($response -notmatch "^[yY]([eE][sS])?$") {
                Write-Info "安装已取消"
                exit 0
            }
        }
        Write-Info "将以当前用户权限继续安装"
        $script:InstallPath = "$env:LOCALAPPDATA\Programs\TmpLink"
    }
    
    Write-Success "系统要求检查通过"
}

function Get-Architecture {
    Write-Step "检测系统架构..."
    
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" {
            $script:ArchDir = "windows-64bit"
            Write-Info "检测到 64位 x86 架构"
        }
        "x86" {
            $script:ArchDir = "windows-32bit"
            Write-Info "检测到 32位 x86 架构"
        }
        default {
            Write-Error "不支持的架构: $arch"
            Write-Info "支持的架构: AMD64 (x64), x86"
            exit 1
        }
    }
}

function Download-Binaries {
    Write-Step "下载二进制文件..."
    
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $script:TempDir = $tempDir.FullName
    
    $guiBinary = "tmplink.exe"
    $cliBinary = "tmplink-cli.exe"
    
    try {
        Write-Info "下载 $guiBinary..."
        $guiUrl = "$DownloadBase/$ArchDir/$guiBinary"
        $guiPath = Join-Path $TempDir $guiBinary
        Invoke-WebRequest -Uri $guiUrl -OutFile $guiPath -UseBasicParsing
        
        Write-Info "下载 $cliBinary..."
        $cliUrl = "$DownloadBase/$ArchDir/$cliBinary"
        $cliPath = Join-Path $TempDir $cliBinary
        Invoke-WebRequest -Uri $cliUrl -OutFile $cliPath -UseBasicParsing
        
        # 验证文件
        if (-not (Test-Path $guiPath) -or -not (Test-Path $cliPath)) {
            throw "下载的文件不存在"
        }
        
        if ((Get-Item $guiPath).Length -eq 0 -or (Get-Item $cliPath).Length -eq 0) {
            throw "下载的文件无效"
        }
        
        Write-Success "二进制文件下载完成"
    } catch {
        Write-Error "下载失败: $($_.Exception.Message)"
        if (Test-Path $TempDir) {
            Remove-Item -Path $TempDir -Recurse -Force
        }
        exit 1
    }
}

function Install-Binaries {
    Write-Step "安装程序到系统..."
    
    $guiBinary = "tmplink.exe"
    $cliBinary = "tmplink-cli.exe"
    $guiSource = Join-Path $TempDir $guiBinary
    $cliSource = Join-Path $TempDir $cliBinary
    
    # 创建安装目录
    if (-not (Test-Path $InstallPath)) {
        Write-Info "创建安装目录: $InstallPath"
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }
    
    # 安装二进制文件
    Write-Info "安装 $guiBinary 到 $InstallPath..."
    Copy-Item $guiSource $InstallPath -Force
    Copy-Item $cliSource $InstallPath -Force
    
    # 移除 Windows Defender 的 Mark of the Web
    $guiPath = Join-Path $InstallPath $guiBinary
    $cliPath = Join-Path $InstallPath $cliBinary
    
    try {
        Unblock-File -Path $guiPath -ErrorAction SilentlyContinue
        Unblock-File -Path $cliPath -ErrorAction SilentlyContinue
    } catch {
        # 忽略错误，某些系统可能不支持
    }
    
    # 设置环境变量
    if (-not $NoPath) {
        Set-EnvironmentPath
    }
    
    Write-Success "程序安装完成"
}

function Set-EnvironmentPath {
    Write-Step "配置环境变量..."
    
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    
    if ($currentPath -notlike "*$InstallPath*") {
        Write-Info "添加 $InstallPath 到 PATH"
        $newPath = "$currentPath;$InstallPath"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        
        # 更新当前会话的 PATH
        $env:Path = "$env:Path;$InstallPath"
        
        Write-Success "PATH 环境变量已更新"
    } else {
        Write-Info "PATH 中已包含安装目录"
    }
}

function Add-WindowsDefenderExclusion {
    Write-Step "配置 Windows Defender 排除..."
    
    try {
        if (Test-Administrator) {
            # 添加安装目录到 Windows Defender 排除列表
            Add-MpPreference -ExclusionPath $InstallPath -ErrorAction Stop
            Write-Success "已添加 Windows Defender 排除规则"
        } else {
            Write-Info "需要管理员权限来配置 Windows Defender 排除"
            Write-Info "如果遇到误报，请手动添加 $InstallPath 到 Windows Defender 排除列表"
        }
    } catch {
        Write-Info "无法自动配置 Windows Defender，请手动添加排除规则"
    }
}

function Create-StartMenuShortcut {
    if ($NoStartMenu) {
        return
    }
    
    Write-Step "创建开始菜单快捷方式..."
    
    $startMenuPath = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs"
    $shortcutPath = Join-Path $startMenuPath "钛盘上传工具.lnk"
    
    $WshShell = New-Object -comObject WScript.Shell
    $shortcut = $WshShell.CreateShortcut($shortcutPath)
    $shortcut.TargetPath = Join-Path $InstallPath "tmplink.exe"
    $shortcut.WorkingDirectory = $InstallPath
    $shortcut.Description = "钛盘上传工具 - TmpLink Uploader"
    $shortcut.Save()
    
    Write-Info "开始菜单快捷方式已创建: $shortcutPath"
    Write-Success "开始菜单快捷方式创建完成"
}

function Create-DesktopShortcut {
    if ($NoDesktop) {
        return
    }
    
    Write-Step "创建桌面快捷方式..."
    
    $desktopPath = [Environment]::GetFolderPath("Desktop")
    $shortcutPath = Join-Path $desktopPath "钛盘上传工具.lnk"
    
    $response = "y"
    if (Test-Path $shortcutPath) {
        $response = Read-Host "桌面快捷方式已存在，是否覆盖？ [y/N]"
    }
    
    if ($response -match "^[yY]([eE][sS])?$") {
        $WshShell = New-Object -comObject WScript.Shell
        $shortcut = $WshShell.CreateShortcut($shortcutPath)
        $shortcut.TargetPath = Join-Path $InstallPath "tmplink.exe"
        $shortcut.WorkingDirectory = $InstallPath
        $shortcut.Description = "钛盘上传工具 - TmpLink Uploader"
        $shortcut.Save()
        
        Write-Info "桌面快捷方式已创建: $shortcutPath"
        Write-Success "桌面快捷方式创建完成"
    } else {
        Write-Info "跳过桌面快捷方式创建"
    }
}

function Test-Installation {
    Write-Step "验证安装..."
    
    $guiPath = Join-Path $InstallPath "tmplink.exe"
    $cliPath = Join-Path $InstallPath "tmplink-cli.exe"
    
    if ((Test-Path $guiPath) -and (Test-Path $cliPath)) {
        Write-Success "安装验证成功"
        Write-Info "GUI 程序: $guiPath"
        Write-Info "CLI 程序: $cliPath"
        
        # 尝试获取版本信息
        try {
            $guiVersion = & $guiPath --version 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Info "GUI 程序版本: $guiVersion"
            } else {
                Write-Info "GUI 程序: 可用"
            }
        } catch {
            Write-Info "GUI 程序: 可用"
        }
        
        try {
            $cliVersion = & $cliPath --version 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Info "CLI 程序版本: $cliVersion"
            } else {
                Write-Info "CLI 程序: 可用"
            }
        } catch {
            Write-Info "CLI 程序: 可用"
        }
    } else {
        Write-Error "安装验证失败，请检查安装过程"
        exit 1
    }
}

function Create-UninstallInfo {
    Write-Step "创建卸载信息..."
    
    $uninstallScript = @"
# 钛盘上传工具卸载脚本
# 删除程序文件
Remove-Item -Path "$InstallPath" -Recurse -Force -ErrorAction SilentlyContinue

# 删除快捷方式
Remove-Item -Path "`$env:APPDATA\Microsoft\Windows\Start Menu\Programs\钛盘上传工具.lnk" -ErrorAction SilentlyContinue
Remove-Item -Path "$([Environment]::GetFolderPath('Desktop'))\钛盘上传工具.lnk" -ErrorAction SilentlyContinue

# 清理环境变量
`$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
`$newPath = `$currentPath -replace [regex]::Escape("$InstallPath"), ""
`$newPath = `$newPath -replace ";;", ";"
[Environment]::SetEnvironmentVariable("Path", `$newPath, "User")

# 移除 Windows Defender 排除规则（需要管理员权限）
try {
    Remove-MpPreference -ExclusionPath "$InstallPath" -ErrorAction Stop
    Write-Host "已移除 Windows Defender 排除规则" -ForegroundColor Green
} catch {
    Write-Host "无法移除 Windows Defender 排除规则，请手动移除" -ForegroundColor Yellow
}

Write-Host "钛盘上传工具已卸载" -ForegroundColor Green
"@
    
    $uninstallPath = Join-Path $InstallPath "uninstall.ps1"
    $uninstallScript | Out-File -FilePath $uninstallPath -Encoding UTF8
    
    Write-Info "卸载脚本已创建: $uninstallPath"
    Write-Success "卸载信息创建完成"
}

function Show-Usage {
    Write-Step "使用说明"
    Write-Host ""
    Write-Host "安装完成！您现在可以使用以下方式启动程序：" -ForegroundColor White
    Write-Host ""
    Write-Host "  1. 双击桌面快捷方式 '钛盘上传工具'" -ForegroundColor White
    Write-Host "  2. 从开始菜单搜索并启动 '钛盘上传工具'" -ForegroundColor White
    Write-Host "  3. 在命令提示符或 PowerShell 中运行：" -ForegroundColor White
    Write-Host "     tmplink      - 启动图形界面版本" -ForegroundColor Cyan
    Write-Host "     tmplink-cli  - 使用命令行版本" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "获取帮助：" -ForegroundColor White
    Write-Host "  tmplink --help" -ForegroundColor Cyan
    Write-Host "  tmplink-cli --help" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "安装位置：" -ForegroundColor White
    Write-Host "  程序文件: $InstallPath" -ForegroundColor Cyan
    Write-Host "  配置文件: $env:USERPROFILE\.tmplink_config.json" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "卸载程序：" -ForegroundColor White
    Write-Host "  运行: $InstallPath\uninstall.ps1" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "更多信息请查看：" -ForegroundColor White
    Write-Host "  https://github.com/$GitHubRepo" -ForegroundColor Cyan
    Write-Host ""
    
    # Windows 特定提示
    Write-Host "Windows 用户提示：" -ForegroundColor Yellow
    Write-Host "  - 如果防火墙弹出提示，请允许程序访问网络" -ForegroundColor White
    Write-Host "  - 如果 Windows Defender 误报，已自动添加到白名单" -ForegroundColor White
    Write-Host "  - 重启命令提示符或 PowerShell 以使用命令行版本" -ForegroundColor White
    Write-Host ""
}

function Invoke-Cleanup {
    Write-Step "清理临时文件..."
    if ($script:TempDir -and (Test-Path $script:TempDir)) {
        Remove-Item -Path $script:TempDir -Recurse -Force
    }
    Write-Success "清理完成"
}

function Main {
    Write-Header
    
    try {
        Test-Requirements
        Get-Architecture
        Download-Binaries
        Install-Binaries
        Add-WindowsDefenderExclusion
        Create-StartMenuShortcut
        Create-DesktopShortcut
        Create-UninstallInfo
        Test-Installation
        Show-Usage
        Invoke-Cleanup
        
        Write-Host ""
        Write-Success "钛盘上传工具安装完成！"
        Write-Host ""
        
        # 询问是否立即启动
        $response = Read-Host "是否立即启动钛盘上传工具？ [y/N]"
        if ($response -match "^[yY]([eE][sS])?$") {
            Start-Process (Join-Path $InstallPath "tmplink.exe")
        }
        
    } catch {
        Write-Error "安装过程中发生错误: $($_.Exception.Message)"
        Write-Host "详细错误信息:" -ForegroundColor Red
        Write-Host $_.Exception.ToString() -ForegroundColor Red
        exit 1
    }
}

# 显示帮助信息
if ($args -contains "-h" -or $args -contains "--help" -or $args -contains "/?") {
    Write-Host "钛盘上传工具 Windows 安装脚本" -ForegroundColor Blue
    Write-Host ""
    Write-Host "用法: .\install-windows.ps1 [选项]" -ForegroundColor White
    Write-Host ""
    Write-Host "选项:" -ForegroundColor Yellow
    Write-Host "  -Force           强制安装，跳过确认提示" -ForegroundColor White
    Write-Host "  -InstallPath     指定安装路径 (默认: $env:ProgramFiles\TmpLink)" -ForegroundColor White
    Write-Host "  -NoStartMenu     不创建开始菜单快捷方式" -ForegroundColor White
    Write-Host "  -NoDesktop       不创建桌面快捷方式" -ForegroundColor White
    Write-Host "  -NoPath          不修改 PATH 环境变量" -ForegroundColor White
    Write-Host "  -h, --help, /?   显示此帮助信息" -ForegroundColor White
    Write-Host ""
    Write-Host "示例:" -ForegroundColor Yellow
    Write-Host "  .\install-windows.ps1" -ForegroundColor Cyan
    Write-Host "  .\install-windows.ps1 -Force" -ForegroundColor Cyan
    Write-Host "  .\install-windows.ps1 -InstallPath 'C:\MyApps\TmpLink'" -ForegroundColor Cyan
    Write-Host ""
    exit 0
}

# 捕获中断信号
trap {
    Write-Error "安装被中断"
    exit 1
}

# 运行主函数
Main