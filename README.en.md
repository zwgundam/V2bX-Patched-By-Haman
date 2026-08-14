# V2bX v1.0 Management Script & Automated Deployment Guide

This repository provides the customized, streamlined **V2bX v1.0** automated management script and compiled binaries built on the **Sing-Box 1.14+** core for lightweight Xboard backend integration.

**V2bX** is a high-performance node backend designed for **Xboard** (modified from XrayR and powered by the Sing-Box core), supporting single-instance multi-node deployment via Machine Mode.

---

## ⚡ Supported Panels & Node Protocols

### 1. Panel Support
- **Xboard** (Recommended, fully supports Machines mode)

### 2. Supported Protocols in this Custom Build
This streamlined fork focuses on modern, high-performance protocols:
- **VLESS** (Supports Reality / XTLS / AnyTLS)
- **Hysteria 2**
- **AnyTLS**

> ⚠️ **Note**: Legacy protocols from original V2bX (such as VMess, Trojan, Shadowsocks) have been pruned in this build to keep runtime efficiency and minimal footprint under Sing-Box 1.14+.

### 3. Feature Matrix

| Feature | VLESS | Hysteria 2 | AnyTLS |
| :--- | :---: | :---: | :---: |
| **Auto ACME TLS Cert Issue/Renew** | √ | √ | √ |
| **Online User Statistics** | √ | √ | √ |
| **Audit & Routing Rules** | √ | √ | √ |
| **Online IP Limit** | √ | √ | √ |
| **TCP Connection Limit** | √ | √ | √ |
| **Cross-node IP Limit** | √ | √ | √ |
| **User-level / Dynamic Speed Limit** | √ | √ | X |

---

## 🚀 Quick Start / One-Click Install

On any Linux distribution (Ubuntu / Debian / CentOS / AlmaLinux, etc.), run the following command as root:

```bash
bash <(curl -fsSL "https://jp.671152.xyz/p/armjp%EF%BC%88local%EF%BC%89/root/.openclaw/workspace/haman-pub/v2bx/v2bx.sh")
```

Once installed, type `v2bx` anywhere in your shell to open the interactive management menu.

---

## 🛠️ Management Menu (`v2bx`)

Typing `v2bx` launches the interactive CLI menu:

```text
=================================================
 V2bX Management Script v1.0 (Sing-Box 1.14+ Custom)
 Status: [ Running ] (Auto-start enabled)
=================================================
  1. Install V2bX
  2. Update V2bX (Core + Script + Rules)
  3. Uninstall V2bX
-------------------------------------------------
  4. Start V2bX
  5. Stop V2bX
  6. Restart V2bX
  7. View V2bX Real-time Logs
-------------------------------------------------
  8. View Current Configuration
  9. Interactive Configuration Wizard
-------------------------------------------------
  0. Exit
=================================================
```

### CLI Shortcut Commands

- `v2bx start` - Start V2bX service
- `v2bx stop` - Stop V2bX service
- `v2bx restart` - Restart V2bX service and view live certificates & logs
- `v2bx status` - Check current service running status and systemd auto-start status
- `v2bx log` - Tail live logs (`journalctl -u V2bX -f -n 50`)
- `v2bx config` - Open interactive configuration wizard

---

## 🌟 Key Features & Innovations in v1.0

1. **Port 80 Intelligent Borrowing & Auto-Recovery**
   - Automatically detects Port 80 occupancy during ACME HTTP-01 certificate challenges (supports Nginx, Caddy, Apache, Lighttpd, Docker containers, etc.).
   - Temporarily suspends the occupying process, performs the ACME validation, and guarantees 100% automatic recovery via Go `defer` mechanisms.

2. **Smart TLS Certificate Lifecycle Management**
   - Scans certificate validity on startup. Auto-cleans expired self-signed or mismatched certificates and requests official ACME certificates from Let's Encrypt.
   - Seamlessly falls back to self-signed certificates upon unexpected ACME failures to ensure node uptime.

3. **Atomic Self-Update Mechanism via Bash `exec`**
   - Menu Option `2` updates the binary core, management script, and Sing-Box 1.14+ rule assets simultaneously.
   - Atomic replacement via `exec` prevents script pointer offset corruption during self-overwrites.

4. **Robust Systemd Service Monitoring**
   - Multi-path detection across service units and `pgrep` process lists for rock-solid health monitoring.

---

## 📁 Shared Directory & Files

- `v2bx.sh` — V2bX v1.0 management script
- `v2bx-linux-arm64` — Pre-built ARM64 binary (QUIC, gRPC, gVisor, WireGuard enabled)
- `v2bx-linux-amd64` — Pre-built x86_64 binary (QUIC, gRPC, gVisor, WireGuard enabled)
- `README.md` — Chinese documentation
- `README.en.md` — English documentation
