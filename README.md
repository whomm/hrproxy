# hrproxy

A http reverse proxy with qps limitation by uri pattern

根据uri的模式（支持正则）限制qps的反向代理服务器

## Versions
    1. 第一个版本写完了，设置配置文件就可以快速使用
## Five Minute Tutorial
    
    ./hrproxy -c def.YAML
    
## Features
    1. 快速构建反向代理服务
    2. 对特定访问url请求模式进行限流
## Usage scenarios
    1. 全局特定PATH限流

