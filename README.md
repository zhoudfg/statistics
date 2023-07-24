# 后台统计服务

## 1. statistics-server
有关用户相关的统计程序
### 1.1 启动
```
mkdir -p /usr/local/statistics/statistics-server/etc/
mkdir -p /usr/local/statistics/bin/
mkdir -p /usr/local/statistics/log/
将配置文件server.yaml放置在/usr/local/statistics/statistics-server/etc下，执行程序statistics-server建议放在/usr/local/statistics/bin/下，nohup ./statistics-server &即可，或可配置supervisor
```