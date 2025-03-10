#!/bin/bash

# Exit on error
set -e

# 设置基础变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DIST_DIR="${PROJECT_ROOT}/dist"
PACKAGE_BINARY=${PACKAGE_BINARY:-"package"}
PACKAGE_TOOL="${PROJECT_ROOT}/scripts/package/${PACKAGE_BINARY}"
VENV_DIR="${PROJECT_ROOT}/.venv"

# 获取版本信息
VERSION=$(grep "VERSION :=" "${PROJECT_ROOT}/Makefile" | cut -d "=" -f2 | tr -d " ")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 获取目标平台
TARGET_OS=${1:-$(go env GOOS)}
TARGET_ARCH=${2:-$(go env GOARCH)}
TARGET_SUFFIX="_${TARGET_OS}_${TARGET_ARCH}"

# 显示构建信息
echo "Building AE CLI:"
echo "  Version:    ${VERSION}"
echo "  Commit:     ${GIT_COMMIT}"
echo "  Build Time: ${BUILD_TIME}"
echo "  Target OS:  ${TARGET_OS}"
echo "  Target Arch:${TARGET_ARCH}"
echo

# 检查跨平台构建的警告
if [ "$TARGET_OS" != "$(go env GOOS)" ] || [ "$TARGET_ARCH" != "$(go env GOARCH)" ]; then
    echo "Warning: Cross-platform build detected!"
    echo "  Host OS:    $(go env GOOS)"
    echo "  Host Arch:  $(go env GOARCH)"
    echo "  Target OS:  ${TARGET_OS}"
    echo "  Target Arch:${TARGET_ARCH}"
    echo
    echo "Note: The resulting binaries can only be executed on ${TARGET_OS}/${TARGET_ARCH} systems."
    echo
fi

# 创建发布目录
mkdir -p "${DIST_DIR}"

# 设置 Python 虚拟环境
if [ -d "${VENV_DIR}" ]; then
    echo "Removing existing virtual environment..."
    rm -rf "${VENV_DIR}"
fi

echo "Creating Python virtual environment..."
# 确保使用 Python 3.9
PYTHON_CMD=$(which python3.9)
if [ -z "$PYTHON_CMD" ]; then
    echo "Error: Python 3.9 not found. Please ensure python3.9 is installed and in your PATH"
    exit 1
fi

PYTHON_VERSION=$($PYTHON_CMD -V | cut -d' ' -f2)
echo "Using Python version: $PYTHON_VERSION"

$PYTHON_CMD -m venv "${VENV_DIR}"

# 激活虚拟环境
source "${VENV_DIR}/bin/activate"

# 验证虚拟环境中的 Python 版本
VENV_PYTHON_VERSION=$(python3 -V)
echo "Virtual environment Python version: $VENV_PYTHON_VERSION"

# 升级 pip 和基础包
echo "Installing Python dependencies..."
python3 -m pip install --upgrade pip setuptools wheel

# 编译打包工具（仅在本地构建时）
if [ -z "$PACKAGE_BINARY" ] || [ "$PACKAGE_BINARY" = "package" ]; then
    echo "Building package tool..."
    cd "${PROJECT_ROOT}/scripts/package"
    GOFLAGS="-buildmode=default" go build -o "${PACKAGE_TOOL}"
fi

# 编译主程序
echo "Building main program..."
cd "${PROJECT_ROOT}"
GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" -o "${DIST_DIR}/ae${TARGET_SUFFIX}.tmp" main.go

# 编译 ESK (Go)
echo "Building ESK component..."
cd "${PROJECT_ROOT}/cmd/esk"
GOOS=$TARGET_OS GOARCH=$TARGET_ARCH go build -o "${DIST_DIR}/esk${TARGET_SUFFIX}.tmp"

# 编译 EI2 (Python)
echo "Building EI2 component..."
cd "${PROJECT_ROOT}/cmd/ei2"

# 安装所有依赖
pip install -e .
pip install pyinstaller

# 构建 PyInstaller 命令的基础部分
PYINSTALLER_CMD="pyinstaller --onefile \
    --name \"ei2${TARGET_SUFFIX}.tmp\" \
    --distpath \"${DIST_DIR}\" \
    --workpath \"${DIST_DIR}/build\" \
    --specpath \"${DIST_DIR}\""

# 自动发现所有 Python 包和模块
for pkg in $(python3 -c "
from setuptools import find_packages
import os
packages = find_packages()
standalone_modules = [f[:-3] for f in os.listdir('.') if f.endswith('.py') and f != 'setup.py']
print(' '.join(packages + standalone_modules)
"); do
    PYINSTALLER_CMD="${PYINSTALLER_CMD} --collect-all ${pkg}"
done

# 执行构建命令
eval "${PYINSTALLER_CMD} ei2.py"

cd "${PROJECT_ROOT}"

# 清理 PyInstaller 生成的临时文件
rm -rf "${DIST_DIR}/build" "${DIST_DIR}/ei2${TARGET_SUFFIX}.tmp.spec"

# 编译 SYS (Rust)
echo "Building SYS component..."
cd "${PROJECT_ROOT}/cmd/sys"

# Clean previous builds
cargo clean

# Build based on target platform
if [ "$TARGET_OS" = "linux" ]; then
    echo "Building for Linux with musl..."
    # Add musl target
    rustup target add x86_64-unknown-linux-musl
    
    # Build with musl
    RUSTFLAGS="-C target-feature=+crt-static" \
    cargo build --release --target x86_64-unknown-linux-musl
    
    cp "target/x86_64-unknown-linux-musl/release/sys" "${DIST_DIR}/sys${TARGET_SUFFIX}.tmp"
else
    echo "Building for ${TARGET_OS}..."
    # Build for native platform
    GOOS=$TARGET_OS GOARCH=$TARGET_ARCH \
    cargo build --release
    
    cp "target/release/sys" "${DIST_DIR}/sys${TARGET_SUFFIX}.tmp"
fi

# 使用打包工具将所有组件打包到主程序中
echo "Packaging components..."
cd "${DIST_DIR}"
"${PACKAGE_TOOL}" "ae${TARGET_SUFFIX}" "ae${TARGET_SUFFIX}.tmp" "esk${TARGET_SUFFIX}.tmp" "ei2${TARGET_SUFFIX}.tmp" "sys${TARGET_SUFFIX}.tmp"

# 清理临时文件
rm -f "ae${TARGET_SUFFIX}.tmp" "esk${TARGET_SUFFIX}.tmp" "ei2${TARGET_SUFFIX}.tmp" "sys${TARGET_SUFFIX}.tmp"

# 设置可执行权限
chmod +x "${DIST_DIR}/ae${TARGET_SUFFIX}"

# 退出虚拟环境
deactivate

echo
echo "Build complete! Binary is available at: ${DIST_DIR}/ae${TARGET_SUFFIX}"
if [ "$TARGET_OS" != "$(go env GOOS)" ] || [ "$TARGET_ARCH" != "$(go env GOARCH)" ]; then
    echo "Note: This binary can only be executed on ${TARGET_OS}/${TARGET_ARCH} systems."
fi
