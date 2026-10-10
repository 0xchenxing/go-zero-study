# 只打包，不在镜像里编译。原因同 api/Dockerfile：仓库里的 .go 是 E-SafeNet 密文，
# Docker 读不到明文，镜内 go build 必然失败。二进制由宿主机 `bash build.sh` 交叉编译。
#
# 构建上下文是仓库根目录（COPY 的路径都相对仓库根），因此根目录那份 .dockerignore 生效。
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 app \
 && mkdir -p /app/etc \
 && chown -R app:app /app

WORKDIR /app
COPY build/user-rpc /app/user-rpc
COPY --chown=app:app rpc/etc/user.docker.yaml /app/etc/user.yaml
# Windows 文件系统没有执行位，COPY 进来是 0644，必须显式补回
RUN chmod 0755 /app/user-rpc

USER app
CMD ["./user-rpc", "-f", "etc/user.yaml"]
