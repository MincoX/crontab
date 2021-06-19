# crontab

# 构建镜像
docker build -t go/crontab .

# 使用脚本默认启动 app 和 worker 两个服务
docker run -d \
    -p 10002:10002 \
    -v /root/project/crontab/config.yaml:/src/crontab/config.yaml \
    -v /root/project/crontab/logs:/src/crontab/logs \
    --name crontab go/crontab:latest

# 指定启动 app 服务，覆盖 Dockerfile 中 CMD 命令执行脚本启动服务
docker run -d \
    -p 10002:10002 \
    -v /root/project/crontab/config.yaml:/src/crontab/config.yaml \
    -v /root/project/crontab/logs:/src/crontab/logs \
    --name crontab go/crontab:latest \
    ./app -config config.yaml

# 指定启动 worker 服务，覆盖 Dockerfile 中 CMD 命令执行脚本启动服务
docker run -d \
    -v /root/project/crontab/config.yaml:/src/crontab/config.yaml \
    -v /root/project/crontab/logs:/src/crontab/logs \
    --name crontab go/crontab:latest \
    ./worker -config config.yaml
