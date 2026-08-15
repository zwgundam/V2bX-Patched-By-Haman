# V2bX-Patched-By-Haman v1.0

[English Document](README.en.md) | [中文文档](README.md)

This repository provides the customized, streamlined **V2bX-Patched-By-Haman v1.0** automated management script and compiled source code built on the **Sing-Box 1.14+** core for lightweight **Xboard** backend Machine Mode integration.

**V2bX** is a high-performance node backend designed for **Xboard** (modified from XrayR and powered by the Sing-Box core), supporting single-instance multi-node deployment via Machine Mode.

---

## ⚡ Supported Panels & Protocols

### 1. Panel Support
- **Xboard** (Recommended, fully supports Machines mode)

### 2. Supported Protocols
This streamlined build focuses on modern, high-performance protocols:
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

## 🚀 One-Click Install & Quick Start

On any Linux distribution (Ubuntu / Debian / CentOS / AlmaLinux, etc.), run the following command as root:

```bash
bash <(curl -fsSL "https://raw.githubusercontent.com/zwgundam/V2bX-Patched-By-Haman/main/v2bx.sh")
```

Once installed, type `v2bx` anywhere in your shell to open the interactive management menu.

---

## 🛠️ Management Menu (`v2bx`)

Typing `v2bx` launches the interactive CLI menu:

```text
=================================================
 V2bX-Patched-By-Haman v1.0 (Sing-Box 1.14+ Custom)
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

## 🌟 Key Features in v1.0

1. **Port 80 Intelligent Borrowing & Auto-Recovery**
   - Automatically detects Port 80 occupancy during ACME HTTP-01 certificate challenges (supports Nginx, Caddy, Apache, Lighttpd, Docker containers, etc.).
   - Temporarily suspends the occupying process, performs the ACME validation, and guarantees 100% automatic recovery via Go `defer` mechanisms.

2. **Smart TLS Certificate Lifecycle Management**
   - Scans certificate validity on startup. Auto-cleans expired self-signed or mismatched certificates and requests official ACME certificates from Let's Encrypt.
   - Seamlessly falls back to self-signed certificates upon unexpected ACME failures to ensure node uptime.

3. **Atomic Self-Update Mechanism via Bash `exec`**
   - Menu Option `2` updates the binary core, management script, and Sing-Box 1.14+ rule assets simultaneously.
   - Atomic replacement via `exec` prevents script pointer offset corruption during self-overwrites.

---

## ☕ Support & Sponsorship

If you find this project helpful, feel free to support its ongoing development!

- **Buy Me A Coffee**: [![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-Donate-orange.svg?style=flat&logo=buy-me-a-coffee)](https://www.buymeacoffee.com/zwgundam)
- **USDT (Solana)**: `D5t943MPXEhPhBoWqRaLDxPAHS6mbSXm9j5AGmFfxuM2`

---

## 🙏 Acknowledgements

Special thanks to the open-source projects and authors that made this project possible:

- [wyx2685/V2bX](https://github.com/wyx2685/V2bX) & [MoeclubM/V2bX](https://github.com/MoeclubM/V2bX) - Original V2bX framework
- [SagerNet/sing-box](https://github.com/SagerNet/sing-box) - Powerful proxy core
- [XrayR-project/XrayR](https://github.com/XrayR-project/XrayR) - Backend architecture reference
- [cedar2025/Xboard](https://github.com/cedar2025/Xboard) - Modern panel support

---

## ⚠️ Disclaimer

- This is an open-source project provided strictly for technical research and educational purposes.
- The author assumes no responsibility for any direct or indirect consequences arising from the use of this project.
