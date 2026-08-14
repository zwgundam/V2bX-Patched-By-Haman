# v2bx-custom 使用与部署说明

> 基于 sing-box 1.14+ 内核的 Xboard / V2board 节点服务端
> 仅保留三种核心协议：**VLESS**、**AnyTLS**、**Hysteria2**

---

## 目录

1. [项目简介](#项目简介)
2. [保留的协议](#保留的协议)
3. [编译构建](#编译构建)
4. [运行时指令](#运行时指令)
5. [配置文件详解](#配置文件详解)
6. [对接 Xboard / V2board](#对接-xboard--v2board)
7. [证书配置](#证书配置)
8. [systemd 服务配置](#systemd-服务配置)
9. [常用操作](#常用操作)
10. [故障排查](#故障排查)

---

## 项目简介

`v2bx-custom` 是一个针对自有部署场景裁剪过的 V2bX 节点服务端。它对接
Xboard / V2board 面板，从面板拉取节点配置与用户列表，动态生成 sing-box
入站协议，并将用户流量上报回面板。

相比上游 V2bX，本仓库砍掉了 Trojan、Shadowsocks、VMess、Naive、TUIC 等
非必要协议，只保留三个当下最常用、对网络环境最友好的协议：

| 协议       | 传输           | 主要场景                           |
| ---------- | -------------- | ---------------------------------- |
| VLESS      | TCP / WS / gRPC / H2 / HTTPUpgrade / QUIC | Reality / TLS / WebSocket 反代     |
| AnyTLS     | TCP            | Padding 抗指纹 TLS                 |
| Hysteria2  | QUIC           | 基于 UDP 的高速伪装传输，支持 Obfs  |

内核已升级并锁定到 sing-box 1.14+ 兼容接口（`github.com/MoeclubM/sing-box_mod v1.14.0-v2bx.3`）。

---

## 保留的协议

`core/sing/sing.go` 中 `Protocols()` 返回值即运行时真正接受的协议种类：

```go
func (b *Sing) Protocols() []string {
    return []string{
        "vless",
        "anytls",
        "hysteria2",
    }
}
```

面板下发的节点 `protocol` 字段只接受上面三种值（含 hysteria legacy 自动归一化为 hysteria2、`v2ray` alias 归一化为 vless）。

---

## 编译构建

### 环境要求

- Go **1.25+**（toolchain 自动拉取，推荐 `go1.25.0` 或更高）
- Linux / macOS（已在 `linux/arm64` 验证）
- 需要联网拉取依赖（`go mod download`）

### 一键编译

在工作区根目录执行：

```bash
cd /root/.openclaw/workspace/v2bx-custom
go build -o v2bx-custom main.go
```

成功后会得到单文件二进制 `v2bx-custom`：

```bash
$ ls -la v2bx-custom
-rwxr-xr-x 1 root root 98800689 Aug 11 22:43 v2bx-custom
```

交叉编译示例：

```bash
# linux/amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o v2bx-custom-linux-amd64 main.go

# linux/arm64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o v2bx-custom-linux-arm64 main.go
```

> 上游面板通信使用 Resty，单二进制无外部运行时依赖，可直接扔到目标机器上跑。

---

## 运行时指令

```
V2bX [command]

Available Commands:
  completion   生成 shell 自动补全脚本
  help         帮助
  log          输出 v2bx-custom 日志
  restart      重启 v2bx-custom 服务
  server       前台运行节点服务端
  start        启动 v2bx-custom 服务（systemd / init.d）
  stop         停止 v2bx-custom 服务
  synctime     从 NTP 服务器同步时间
  uninstall    卸载 v2bx-custom
  update       在线更新 v2bx-custom
  version      打印版本信息
  x25519       生成 X25519 密钥对
```

最常用：

```bash
./v2bx-custom version                       # 查版本
./v2bx-custom server -c /etc/V2bX/config.json    # 前台跑（调试用）
sudo ./v2bx-custom install                  # 安装为系统服务
sudo ./v2bx-custom start                    # 启动
sudo ./v2bx-custom status                   # 状态
```

`./v2bx-custom version` 输出示例：

```
  _/      _/    _/_/    _/        _/      _/
 _/      _/  _/    _/  _/_/_/      _/  _/
_/      _/      _/    _/    _/      _/
 _/  _/      _/      _/    _/    _/  _/
  _/      _/_/_/_/  _/_/_/    _/      _/

V2bX TempVersion (A V2board backend based on multi core)
```

---

## 配置文件详解

节点程序配置 = 一个 JSON，支持 JSON5 注释与尾逗号，默认路径 `/etc/V2bX/config.json`。

### 完整示例

```jsonc
{
  "Log": {
    "Level": "info",          // debug / info / warn / error
    "Output": ""              // 留空走 stderr
  },

  "Cores": [
    {
      "Name": "v2bx-custom-core",

      "Log": {
        "Disable": false,
        "Level": "info",
        "Timestamp": true,
        "Output": ""
      },

      "NTP": {
        "Enable": false,
        "Server": "time.apple.com",
        "ServerPort": 0
      },

      // 指定一份可选的 “sing-box 原生配置” 作为基线，
      // 留空则所有 outbounds 由节点程序按路由规则动态生成
      "OriginalPath": "/etc/V2bX/sing_origin.json"
    }
  ],

  "Machines": [
    {
      // ---- 面板相关 ----
      "ApiHost": "https://your-xboard.example.com",
      "ApiKey":  "panel-api-key",
      "APISendIP": "203.0.113.10",   // 可选：节点向面板上报使用的源 IP
      "NodeID":  42,                  // 节点 ID（v2board 必填）
      "NodeType": "vless",            // 仅支持 vless / anytls / hysteria2

      // ---- sing-box 高级选项（可选） ----
      "EnableTFO": false,
      "EnableSniff": true,
      "SniffOverrideDestination": true,
      "EnableDNS": false,
      "DomainStrategy": "prefer_ipv4",

      // 多路复用（仅对 vless 生效）
      "MultiplexConfig": {
        "Enable": false,
        "Padding": true,
        "Brutal": { "Enable": false, "UpMbps": 0, "DownMbps": 0 }
      },

      // ---- 证书 ----
      "CertConfig": {
        "CertMode":     "self",              // none / file / dns / http / self
        "CertDomain":   "node.example.com",  // 申请证书的域名
        "Email":        "[email protected]",
        "Provider":     "cloudflare",        // dns 模式使用
        "DNSEnv": {
          "CLOUDFLARE_EMAIL":   "...",
          "CLOUDFLARE_API_KEY": "..."
        },
        "CertFile": "/etc/V2bX/cert/fullchain.pem",
        "KeyFile":  "/etc/V2bX/cert/privkey.pem"
      }
    }
  ]
}
```

### 字段说明

#### `Machines[i].NodeType`

仅接受 `vless` / `anytls` / `hysteria2`。其它值（vmess / trojan / shadowsocks / naive / tuic / hysteria 1）会被拒绝启动并显式报错。

#### `Machines[i].EnableSniff`

开启后 sing-box 会按 SNI / HTTP Host 嗅探目标地址。建议打开。

#### `Machines[i].CertConfig.CertMode`

| 模式   | 行为 |
| ------ | ---- |
| `none` | 不会自动签发 / 申请证书，必须显式给 `CertFile` + `KeyFile` |
| `file` | 直接使用 `CertFile` / `KeyFile` 指向的本地证书，不签发 |
| `self` | 用当前机器信息自签一份 30 天的临时证书，调试用 |
| `dns`  | 通过 acme DNS-01 申请（需要 `Provider` 与 `DNSEnv`） |
| `http` | 通过 acme HTTP-01 申请（需要 80 端口可达） |

---

## 对接 Xboard / V2board

### 在面板侧创建节点

1. 进入管理员后台 → 节点管理 → 添加节点。
2. **协议** 选 `vless` / `anytls` / `hysteria2` 之一（其余协议已被裁剪）。
3. 端口、传输、TLS、Reality / 证书等参数按协议填好，记下 `节点 ID`。
4. 复制节点所属的 `Machine Token`。

### 在节点机器侧配置

把上面的 `ApiHost` / `ApiKey` / `NodeID` / `NodeType` 填进 `/etc/V2bX/config.json` 的 `Machines[0]`，然后：

```bash
sudo mkdir -p /etc/V2bX
sudo cp config.json /etc/V2bX/config.json
sudo ./v2bx-custom install
sudo ./v2bx-custom start
./v2bx-custom log
```

首次启动会：

1. 调面板 `/api/v2/server/config` 拉取本节点的协议 / 端口 / TLS / 路由配置
2. 调 `/api/v2/server/user` 拉取用户列表 → 写入 sing-box inbound
3. 按 `pull_interval` 周期重拉配置，按 `push_interval` 周期上报流量

### 三种协议各自需要的最小面板参数

#### VLESS（Reality）

- `tls_settings.reality_settings`：dest / server_name / private_key / short_ids
- `network`：tcp / ws / grpc / httpupgrade / http / quic
- `flow`：xtls-rprx-vision（按需）

#### AnyTLS

- `padding_scheme`：字符串数组，例如 `["100-200"]`、`null`

#### Hysteria2

- `tls_settings`：证书 / Reality
- `up_mbps` / `down_mbps`：带宽限制（可填 `0` + `ignore_client_bandwidth: true` 关闭）
- `obfs`：salamander（最常用）
- `obfs-password`：混淆密码

---

## 证书配置

只列 Hysteria2 / VLESS-TLS / AnyTLS 用得到证书的链路。

### self（最快，本地测试用）

```json
"CertConfig": {
  "CertMode": "self",
  "CertDomain": "node.example.com"
}
```

自签证书 30 天有效，仅适合内部测试。

### dns（生产推荐）

通过 acme + Cloudflare（示例）：

```json
"CertConfig": {
  "CertMode": "dns",
  "CertDomain": "node.example.com",
  "Email": "[email protected]",
  "Provider": "cloudflare",
  "DNSEnv": {
    "CLOUDFLARE_EMAIL":   "[email protected]",
    "CLOUDFLARE_API_KEY": "your-api-key"
  }
}
```

支持 `cloudflare` / `alidns` / `dnspod` / `godaddy` / `namecheap` / `namecom` / `tencentcloud` 等。

### file（你已经签好了）

```json
"CertConfig": {
  "CertMode": "file",
  "CertFile": "/etc/letsencrypt/live/node.example.com/fullchain.pem",
  "KeyFile":  "/etc/letsencrypt/live/node.example.com/privkey.pem"
}
```

---

## systemd 服务配置

`./v2bx-custom install` 会自动写 `/etc/systemd/system/V2bX.service`（参见 `cmd/install_linux.go`）。
如果你想自己手写一份，参考模板：

```ini
[Unit]
Description=v2bx-custom - V2board backend node server (sing-box 1.14+, vless/anytls/hysteria2)
After=network-online.target nss-lookup.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/V2bX
ExecStart=/usr/local/bin/v2bx-custom server -c /etc/V2bX/config.json
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=5s
LimitNOFILE=1048576
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

# 优雅退出
KillSignal=SIGTERM
TimeoutStopSec=30s

# 日志走 journald，可通过 `journalctl -u V2bX` 查看
StandardOutput=journal
StandardError=journal
SyslogIdentifier=V2bX

[Install]
WantedBy=multi-user.target
```

启用 / 启动：

```bash
sudo cp ./v2bx-custom /usr/local/bin/v2bx-custom
sudo chmod +x /usr/local/bin/v2bx-custom
sudo cp ./v2bx-custom.service /etc/systemd/system/V2bX.service
sudo systemctl daemon-reload
sudo systemctl enable --now V2bX
sudo systemctl status V2bX
sudo journalctl -u V2bX -f   # 跟日志
```

### 一键运行（不开 systemd）

适合临时调试或容器场景：

```bash
# 直接前台跑
./v2bx-custom server -c ./config.json

# 用 screen / tmux 放后台
screen -dmS v2bx ./v2bx-custom server -c ./config.json

# 用 nohup
nohup ./v2bx-custom server -c ./config.json > /var/log/v2bx.log 2>&1 &
```

容器场景（docker）示例 `Dockerfile` 片段：

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot
COPY v2bx-custom /v2bx-custom
COPY config.json /etc/V2bX/config.json
ENTRYPOINT ["/v2bx-custom", "server", "-c", "/etc/V2bX/config.json"]
```

---

## 常用操作

```bash
# 查版本
./v2bx-custom version

# 时间同步（每次启动会自动同步，可手动再触发）
./v2bx-custom synctime

# 生成 X25519 密钥（用于 Reality 配置）
./v2bx-custom x25519

# 服务管理（等价于 systemctl）
./v2bx-custom install
./v2bx-custom start
./v2bx-custom stop
./v2bx-custom restart
./v2bx-custom status
./v2bx-custom log
./v2bx-custom uninstall
./v2bx-custom update      # 在线更新同架构二进制
```

---

## 故障排查

### 启动报 `core config is nil`

`config.json` 里没有 `Cores` 段，或 `Cores` 是空数组。

### 启动报 `unsupported Node type`

`Machines[0].NodeType` 不是 `vless` / `anytls` / `hysteria2`。检查大小写。

### 启动报 `decode xxx params error: ...`

面板侧返回的协议 / 字段与本地协议模型不匹配，确认面板版本支持 `vless` / `anytls` / `hysteria2` 下发。

### 报 `unsupported node type returned by panel`

面板返回了被裁剪的协议（vmess / trojan / shadowsocks / naive / tuic / hysteria v1）。请在面板端把节点协议改成三种保留协议之一。

### Hysteria2 没有流量

1. 检查服务器 UDP 入站是否被云厂商防火墙拦了（UDP 不是 TCP，security group 必须放 UDP）；
2. 检查客户端 obfs 是否与 `obfs-password` 一致。

### VLESS Reality 连不上

1. dest 目标得是真实可达的 TLS 服务；
2. 确认 `private_key` / `short_id` 与面板一致；
3. 时间差没超过 `MaxTimeDiff`（默认建议 ≥ 30s，可放宽）。

---

## 升级迁移说明

如果之前跑过上游 V2bX，要把所有节点协议改成 `vless` / `anytls` / `hysteria2`，`vmess` / `trojan` / `shadowsocks` / `naive` / `tuic` / `hysteria v1` 节点会被服务端拒绝启动并报：

```
unsupported node type returned by panel: <protocol>
(only vless/anytls/hysteria2 are supported in this build)
```

升级步骤：

1. 在面板上把旧协议节点改成 `vless` 或 `anytls` 或 `hysteria2`，让用户改订阅；
2. 替换节点二进制，重新加载服务。

---

## 版本 & 依赖

- 内核：`github.com/MoeclubM/sing-box_mod v1.14.0-v2bx.3`（sing-box 1.14+ 接口）
- Go toolchain：`go1.25.0`（已在 `go1.26.0` 编译验证）
- 主要依赖：resty / fsnotify / lego / cobra / msgpack / logrus / ntp
