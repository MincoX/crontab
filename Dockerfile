FROM golang:alpine

# 为我们的镜像设置必要的环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

# 移动到工作目录：/build
WORKDIR /build

# 将代码复制到容器中
COPY . .

# 将我们的代码编译成二进制可执行文件app
RUN go build -o app master/main.go && go build -o main worker/main.go

# 移动到工作目录：/build
WORKDIR /src/crontab

# 将二进制文件从 /build 目录复制到这里
RUN cp /build/app . && cp /build/main . && cp /build/service_launch.sh . && chmod +x app main service_launch.sh

# 声明服务端口
EXPOSE 10002

# 启动容器时运行的命令
CMD ./service_launch.sh
