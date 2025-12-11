#!/bin/bash

# 编译 protoc-gen-gonfig 的多平台二进制文件并打包成zip
# 支持的操作系统和架构组合

set -e  # 遇到错误时退出

# 项目名称
PROJECT_NAME="protoc-gen-gonfig"

# 版本号（可以根据需要修改）
VERSION="v0.0.2"

# 输出目录
DIST_DIR="dist"
mkdir -p "${DIST_DIR}"

# 支持的操作系统和架构
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

echo "开始编译 ${PROJECT_NAME} ${VERSION}"

# 遍历所有平台进行编译
for platform in "${PLATFORMS[@]}"; do
    # 分离操作系统和架构
    OS="${platform%/*}"
    ARCH="${platform#*/}"
    
    echo "正在编译 ${OS}/${ARCH}..."
    
    # 创建输出目录
    OUTPUT_DIR="${DIST_DIR}/${PROJECT_NAME}-${VERSION}-${OS}-${ARCH}"
    BIN_DIR="${OUTPUT_DIR}/bin"
    mkdir -p "${BIN_DIR}"

    # 设置输出文件名
    if [ "$OS" = "windows" ]; then
        OUTPUT_NAME="${PROJECT_NAME}.exe"
    else
        OUTPUT_NAME="${PROJECT_NAME}"
    fi

    OUTPUT_FILE="${BIN_DIR}/${OUTPUT_NAME}"
    
    # 设置环境变量并编译
    CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build -o "${OUTPUT_FILE}" "./cmd/${PROJECT_NAME}"
    
    # 检查编译是否成功
    if [ $? -ne 0 ]; then
        echo "编译 ${OS}/${ARCH} 失败"
        exit 1
    fi
    
    echo "编译完成: ${OUTPUT_FILE}"

    ZIP_FILE="${PROJECT_NAME}-${VERSION}-${OS}-${ARCH}.zip"
    echo "正在创建 ${ZIP_FILE}..."

    (cd "${DIST_DIR}" && zip -r "${ZIP_FILE}" "$(basename ${OUTPUT_DIR})")
    
    echo "删除文件夹 ${OUTPUT_DIR}"
    rm -rf "${OUTPUT_DIR}"
done

echo "所有二进制文件已编译并打包完成！"
echo "输出文件位于: ${DIST_DIR}/"

# 显示生成的文件列表
echo "生成的文件列表:"
ls -la "${DIST_DIR}"