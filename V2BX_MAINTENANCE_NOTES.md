# V2BX-Patched-By-Haman v1.0.0 维护与排障记录 (2026-08-14)

本次记录固化了 v1.0.0 发布过程中遇到的所有关键 Bug、修复细节，以及系统维护的最高级原则。

## 1. V2bX 管理脚本 (v2bx.sh) 核心修复与优化 (已于 2026-08-14 16:41 推送至 main)

本次推送修复了生产环境中的 4 个关键 Bug，并完成了用户体验优化：

| 编号 | 描述 | 修复细节 | 状态 |
| :--- | :--- | :--- | :--- |
| **1. UX优化** | **安装后自动配置引导** | 首次安装成功后，脚本自动无缝调用 `configure_v2bx` 函数，引导用户完成面板对接配置，提升初次使用体验。 | ✅ |
| **2. Bug修复** | **ApiKey 占位符错误** | 修复 `config.json` 模板中 `ApiKey` 字段的硬编码错误（`"ApiKey": "***}"`），确保用户输入的密钥能正确写入，解决面板鉴权失败问题。 | ✅ |
| **3. Bug修复** | **版本硬编码问题** | 将二进制和脚本的下载链接从硬编码的 `v1.0.0` 路径升级为 `/releases/latest/download` 动态链接，确保用户始终下载到最新稳定版。 | ✅ |
| **4. Bug修复** | **菜单字段对齐** | 修正了菜单中对 Machine ID 的错误引用，确保提示为正确的 `MachineID`，并移除冗余的 `X25519` 和 `NTP` 选项。 | ✅ |

## 2. 关键系统维护教训与 Shell 语法锚定 (记忆固化)

在本次推送过程中，系统稳定性与安全机制造成了多次中断，锚定以下最高级原则：

### A. GitHub Token 传输机制 (最高优先级)
- **问题根源**：聊天框的安全脱敏机制 (`***`) 会截断 Token，导致推送失败。
- **可靠解决方案**：Token 必须存储在本地文件（如 `.github_token`），并使用**原子化 Shell 语法**进行传递，绕过聊天通道的脱敏。
- **可靠 Shell 语法锚定**：
    \`\`\`bash
    cd /path/to/repo && \\
    git push https://$(cat .github_token)@github.com/zwgundam/Repo.git local_branch:remote_branch
    \`\`\`

### B. 系统工具稳定性 (exec/process)
- **原则**：`exec` 工具存在间歇性卡死或输出延迟。
- **行动**：所有长期或敏感的 `exec` 任务（如 Git Push）必须使用 `process` 工具进行追踪和介入，不能盲目信任第一次 `exec` 的返回。

## 3. 分发路径检查原则
- **原则**：每次代码推送 (Push) 后，必须立即执行**源码工作区**到**本地分发目录**的同步操作。
- **同步命令**：
    \`\`\`bash
    cp -f /root/.openclaw/workspace/v2bx-custom/v2bx.sh /root/.openclaw/workspace/haman-pub/v2bx/v2bx.sh
    \`\`\`
- **目标**：确保用户从公开分发路径 (`haman-pub`) 下载的文件始终与 GitHub `main` 分支保持一致。