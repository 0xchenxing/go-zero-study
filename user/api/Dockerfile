# 只打包，不在镜像里编译。
#
# goctl 生成的原版是多阶段构建（COPY . . + RUN go build），本机跑不通：装了 E-SafeNet
# 透明加密，仓库里的 .go 是密文，只有白名单进程（go.exe / git.exe / 编辑器）读到的是明文，
# Docker 不在白名单里，镜内 go build 只会读到密文并报 `expected 'package', found 'b'`。
#
# 所以二进制改由宿主机交叉编译好再 COPY 进来：
#
#     bash build.sh          # 产物 build/user-api、build/user-rpc
#     docker compose up -d
#
# 构建上下文是仓库根目录（COPY 的路径都相对仓库根），因此根目录那份 .dockerignore 生效。
FROM alpine:3.20

# /app/uploads 必须先建好并 chown：命名卷首次挂载时用镜像里该路径的属主去初始化，
# 否则非 root 的 app 用户写上传目录会 permission denied（对外表现为 HTTP 200 + 业务码 9001，
# 真实错误只出现在容器日志里）。
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 app \
 && mkdir -p /app/etc /app/uploads \
 && chown -R app:app /app

WORKDIR /app
COPY build/user-api /app/user-api
COPY --chown=app:app api/etc/user-api.docker.yaml /app/etc/user-api.yaml

# Windows 文件系统没有执行位，COPY 进来是 0644，必须显式补回
RUN chmod 0755 /app/user-api

USER app
CMD ["./user-api", "-f", "etc/user-api.yaml"]
