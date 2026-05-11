# mini-go-im
基于 Go 的轻量级即时通信系统

## v0.1.0（2026-05-11）

首个可用版本，搭建基础 TCP 服务框架。

- 实现 TCP Server 结构体，支持指定 IP 和端口启动
- 采用 goroutine-per-connection 并发模型处理连接
- 基本的连接建立确认（连接创建时打印提示信息）
- 项目模块初始化（go.mod、.gitignore）
