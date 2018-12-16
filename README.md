# hrproxy

A http reverse proxy with qps limitation by uri pattern

## Versions

## Five Minute Tutorial
    ```shell
    ./hrproxy -b192.168.0.1:8091,192.168.0.2:8091 -l127.0.0.1:80 --limit='*\/_bulk',90
    ```
## Features
    1. 快速构建反向代理服务
    2. 对特定访问url请求模式进行限流
## Usage scenarios
    1. 全局特定PATH限流：hive-elasticsearch 交互中（hive外表存储在elasticsearch中），对hive批量写入elasticsearch的 /*/*/_bulk 方法进行全局限流保障elasticsearch的服务稳定性。（注：hive和elasticsearch集群交互是使用http协议）

