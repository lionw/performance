# 系统监控报警工具 (System Monitor Tool)

这是一个基于 Go 语言开发的轻量级服务器系统监控工具。它可以定期检查服务器的 CPU、内存和磁盘使用情况，并在指标超过设定阈值时通过钉钉发送报警通知。同时支持将检查结果记录到本地日志文件。

## 主要功能

- **多维度监控**: 实时监控 CPU 使用率、内存使用率、磁盘使用率。
- **钉钉报警**: 当监控指标超过配置的阈值时，自动发送钉钉群消息。
  - 报警信息包含：当前时间、服务器 IP 地址、具体的资源使用详情（总量、已用、剩余）。
- **灵活配置**:
  - 支持自定义检查间隔（秒）。
  - 支持自定义 CPU、内存、磁盘的报警阈值（百分比）。
  - 支持配置报警生效的时间段（例如只在 09:00 - 18:00 报警）。
  - 提供全局开关，可一键开启或关闭监控。
- **日志记录**: 所有的检查结果（无论是否报警）都会被记录到本地日志文件中，按天归档。

## 目录结构

```
.
├── config.json          # 配置文件
├── main.go              # 程序入口
├── monitor/             # 监控逻辑包
├── dingtalk/            # 钉钉发送逻辑包
├── logs/                # 日志目录 (自动生成)
└── go.mod               # Go 模块文件
```

## 快速开始

### 1. 配置 `config.json`

在程序运行目录下创建或修改 `config.json` 文件：

```json
{
    "dingtalk_access_token": "你的钉钉机器人AccessToken",
    "at_mobiles": [
        "13800138000"  // 需要@的手机号
    ],
    "check_interval_seconds": 60,   // 检查间隔（秒）
    "enable_check": true,           // 是否开启检查
    "cpu_threshold": 90.0,          // CPU 报警阈值 (%)
    "mem_threshold": 90.0,          // 内存报警阈值 (%)
    "disk_threshold": 90.0,         // 磁盘报警阈值 (%)
    "alert_start_time": "09:00",    // 报警生效开始时间
    "alert_end_time": "18:00"       // 报警生效结束时间
}
```

### 2. 运行程序

```bash
go run main.go
```

或者编译后运行：

```bash
go build -o monitor.exe main.go
./monitor.exe
```

## 日志说明

程序运行时会自动在 `logs` 目录下生成日志文件，文件名格式为 `YYYY-MM-DD.log`。

日志内容示例：
```
[2026-02-12 14:00:01] IP: 192.168.1.100
CPU Usage: 5.08%
Memory Usage: 58.00% (Total: 31.85 GB, Used: 18.75 GB, Available: 13.09 GB)
Disk Usage (C:): 99.49% (Total: 199.23 GB, Used: 198.21 GB, Free: 1.01 GB)
----------------------------------------
```

## 报警示例

当资源使用率超过阈值且当前时间在配置的时间段内时，钉钉群会收到如下消息：

```
Time: 2026-02-12 14:00:01
IP: 192.168.1.100
Memory Usage High: 92.50% (Threshold: 90.00%, Total: 32.00 GB, Used: 29.60 GB, Available: 2.40 GB)
```
