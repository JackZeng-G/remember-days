# 使用最小化基础镜像
FROM alpine:3.19

# 安装 ca-certificates 用于 HTTPS 请求（如需要）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 复制预编译的二进制文件
COPY build/remember .

# 复制静态资源
COPY web/ ./web/

# 复制数据目录（如需初始数据）
COPY data/ ./data/

# 设置权限
RUN chown -R appuser:appuser /app

USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# 启动服务
CMD ["./remember"]