# V2bX-Patched-By-Haman v1.0

[English Document](README.en.md) | [中文文档](README.md)

本仓库提供基于 **Sing-Box 1.14+** 核心、针对 **Xboard** 机器模式精简重构的 **V2bX-Patched-By-Haman v1.0** 自动化管理脚本与编译源码。

**V2bX** 是专为 **Xboard** 打造的高性能节点服务端（修改自 XrayR 并基于 Sing-Box 内核），支持单实例通过机器模式对接面板。

---

## ⚡ 支持面板与节点协议

### 1. 对接面板支持
- **Xboard**（推荐，完美支持 Machines 机器模式）

### 2. 支持协议
本重构版本专注于现代高效协议，已剔除旧协议（VMess/Trojan/Shadowsocks）：
- **VLESS** (支持 Reality / XTLS / AnyTLS)
- **Hysteria 2**
- **AnyTLS**

### 3. 节点管理功能矩阵

| 功能特性 | VLESS | Hysteria 2 | AnyTLS |
| :--- | :---: | :---: | :---: |
| **自动申请/续签 TLS** | √ | √ | √ |
| **在线人数统计** | √ | √ | √ |
| **审计与规则路由** | √ | √ | √ |
| **在线 IP 限制** | √ | √ | √ |
| **TCP 连接数限制** | √ | √ | √ |
| **跨节点 IP 限制** | √ | √ | √ |
| **用户级/动态限速** | √ | √ | X |

---

## 🚀 一键安装与快速启动

在任何 Linux（Ubuntu / Debian / CentOS / AlmaLinux 等）系统上，只需以 root 权限运行以下命令即可完成一键安装：

```bash
bash <(curl -fsSL "https://raw.githubusercontent.com/zwgundam/V2bX-Patched-By-Haman/main/v2bx.sh")
```

安装完成后，可以在系统中任何路径直接输入命令 `v2bx` 呼出交互菜单。

---

## 🛠️ 管理菜单功能说明 (`v2bx`)

输入 `v2bx` 后，系统将弹出全功能菜单：

```text
=================================================
 V2bX-Patched-By-Haman v1.0 (Sing-Box 1.14+ 重构版)
 物理运行状态: [ 运行中 ] (已开启开机自启)
=================================================
  1. 安装 V2bX
  2. 更新 V2bX (内核 + 脚本 + 规则)
  3. 卸载 V2bX
-------------------------------------------------
  4. 启动 V2bX
  5. 停止 V2bX
  6. 重启 V2bX
  7. 查看 V2bX 运行日志
-------------------------------------------------
  8. 查看 当前配置文件
  9. 交互式修改 配置参数
-------------------------------------------------
  0. 退出脚本
=================================================
```

### 命令行快捷参数 (CLI Direct Access)

- `v2bx start` - 启动 V2bX 服务并实时查看日志
- `v2bx stop` - 停止 V2bX 服务
- `v2bx restart` - 重启 V2bX 服务并实时查看证书与运行日志
- `v2bx status` - 查看服务当前物理运行状态与开机自启状态
- `v2bx log` - 实时追踪 V2bX 服务日志 (`journalctl -u V2bX -f -n 50`)
- `v2bx config` - 呼出交互式配置菜单

---

## 🌟 v1.0 核心黑科技与自动化特性

1. **80 端口智能借用与自动恢复引擎**
   - 申请 HTTP-01 证书时，Go 内核会自动探测 80 端口占用情况（支持识别 Nginx, Caddy, Apache, Lighttpd, Docker 容器等）。
   - 若 80 端口被占用，会自动毫秒级挂起占用服务，完成 ACME 证书申请后通过 Go 语言 `defer` 保证 100% 自动重启恢复原服务，实现小白“零配置无感申请证书”。

2. **智能证书生命周期管理**
   - 启动时自动扫描证书状态。若存在过期的自签证书、域名不匹配证书，会自动清理并向 Let's Encrypt 重新申请合法 ACME 证书。
   - 申请失败时自动无缝降级为自签名证书，确保节点绝对不会因证书问题崩溃宕机。

3. **Bash `exec` 原子的无缝自我更新**
   - 菜单选项 `2` 支持内核二进制、管理脚本本体及 Sing-Box 1.14+ 规则的三合一更新。
   - 更新过程附带三重完整性校验（文件大小、语法检测、特征标识），并通过 `exec` 进行进程原子重载，杜绝自我覆盖导致的字节错乱。

---

## ☕ 赞赏 (Buy Me A Coffee)

如果你觉得本项目对你有帮助，欢迎赞赏支持开发者的持续维护！

[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-Donate-orange.svg?style=flat&logo=buy-me-a-coffee)](https://www.buymeacoffee.com)

---

## 🙏 致谢 (Thanks)

本项目基于开源社区优秀项目二次重构，特别感谢以下开源先驱：

- [wyx2685/V2bX](https://github.com/wyx2685/V2bX) & [MoeclubM/V2bX](https://github.com/MoeclubM/V2bX) - 原 V2bX 服务端架构
- [SagerNet/sing-box](https://github.com/SagerNet/sing-box) - 强大的通用网络代理引擎
- [XrayR-project/XrayR](https://github.com/XrayR-project/XrayR) - 优秀的后端架构参考
- [cedar2025/Xboard](https://github.com/cedar2025/Xboard) - 现代化面板支持

---

## ⚠️ 免责声明 (Disclaimer)

- 本项目为开源免费项目，仅供个人技术交流与学习使用，严禁用于任何违法违规用途。
- 不对任何人使用本项目造成的任何直接或间接损失承担责任。
