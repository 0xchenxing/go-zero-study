#!/usr/bin/env bash
# 宿主机交叉编译两个服务，产物落在 build/。
#
# 为什么不在镜像里编译：仓库里的 .go 是 E-SafeNet 透明加密后的密文，只有白名单进程
# （go.exe / git.exe / 编辑器）读到的是明文，Docker 不在白名单里，镜内 go build 必然失败。
# 详见 api/Dockerfile 顶部注释。
#
# 用法：
#     bash build.sh && docker compose up -d
set -euo pipefail
cd "$(dirname "$0")"

export CGO_ENABLED=0 GOOS=linux GOARCH=amd64

mkdir -p build
go build -trimpath -ldflags="-s -w" -o build/user-api ./api
go build -trimpath -ldflags="-s -w" -o build/user-rpc ./rpc

ls -lh build
