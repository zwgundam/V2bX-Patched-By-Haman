#!/usr/bin/env bash
# =========================================================
# V2bX-Patched-By-Haman 标准原子化 Git 提交流水线
# =========================================================
set -euo pipefail

REPO_DIR="/root/.openclaw/workspace/v2bx-custom"
PUB_DIR="/root/.openclaw/workspace/haman-pub/v2bx"
TOKEN_FILE="/root/.openclaw/workspace/keys/github_token"
REMOTE_REPO="zwgundam/V2bX-Patched-By-Haman.git"

COMMIT_MSG="${1:-"fix: update v2bx repository files"}"

echo ">>> [1/4] 同步分发镜像目录..."
if [[ -f "${REPO_DIR}/v2bx.sh" ]]; then
    cp -f "${REPO_DIR}/v2bx.sh" "${PUB_DIR}/v2bx.sh"
fi

echo ">>> [2/4] 语法自检..."
bash -n "${REPO_DIR}/v2bx.sh"

echo ">>> [3/4] 提交本地 Git..."
cd "${REPO_DIR}"
git add -A
if ! git diff --cached --quiet; then
    git commit -m "${COMMIT_MSG}"
else
    echo "工作区无变动，跳过 commit。"
fi

echo ">>> [4/4] 安全推送至 GitHub main..."
if [[ ! -f "${TOKEN_FILE}" ]]; then
    echo "错误：未找到 token 文件 ${TOKEN_FILE}" >&2
    exit 1
fi

GH_TOKEN=$(tr -d '\n\r ' < "${TOKEN_FILE}")
git push "https://${GH_TOKEN}@github.com/${REMOTE_REPO}" main

echo "✅ [SUCCESS] V2bX 流水线发布完成！"
