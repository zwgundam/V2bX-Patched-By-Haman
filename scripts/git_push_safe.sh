#!/usr/bin/env bash
# =========================================================
# V2bX-Patched-By-Haman 标准原子化 Git 提交流水线
# 职责:
# 1. 自动全架构编译 (amd64 + arm64) 嵌入最新 build tags
# 2. 同步更新本地镜像目录 haman-pub/v2bx
# 3. 语法与状态自检
# 4. 提交本地 Git 并安全推送至 GitHub main
# 5. 调用 GitHub API 自动同步更新 Releases 二进制资产 (彻底防止开倒车)
# =========================================================
set -euo pipefail

REPO_DIR="/root/.openclaw/workspace/v2bx-custom"
PUB_DIR="/root/.openclaw/workspace/haman-pub/v2bx"
TOKEN_FILE="/root/.openclaw/workspace/keys/github_token"
REMOTE_REPO="zwgundam/V2bX-Patched-By-Haman"

COMMIT_MSG="${1:-"fix: update v2bx repository and release assets"}"

echo "========================================================="
echo ">>> [1/5] 编译最新全架构二进制 (with_utls, with_quic 等完整 tags)..."
echo "========================================================="
mkdir -p "${REPO_DIR}/build_assets"
mkdir -p "${PUB_DIR}"

cd "${REPO_DIR}"
echo "--> 正在编译 arm64 二进制..."
GOEXPERIMENT=jsonv2 CGO_ENABLED=0 go build -v -o build_assets/V2bX-linux-arm64 -tags "sing with_quic with_grpc with_utls with_wireguard with_acme with_gvisor" -trimpath -ldflags "-X 'github.com/MoeclubM/V2bX/cmd.version=v1.0.0-haman' -s -w -buildid="

echo "--> 正在交叉编译 amd64 二进制..."
GOEXPERIMENT=jsonv2 GOARCH=amd64 CGO_ENABLED=0 go build -v -o build_assets/V2bX-linux-amd64 -tags "sing with_quic with_grpc with_utls with_wireguard with_acme with_gvisor" -trimpath -ldflags "-X 'github.com/MoeclubM/V2bX/cmd.version=v1.0.0-haman' -s -w -buildid="

echo "========================================================="
echo ">>> [2/5] 同步分发镜像目录..."
echo "========================================================="
if [[ -f "${REPO_DIR}/v2bx.sh" ]]; then
    cp -f "${REPO_DIR}/v2bx.sh" "${PUB_DIR}/v2bx.sh"
fi
cp -f "${REPO_DIR}/build_assets/V2bX-linux-arm64" "${PUB_DIR}/V2bX-linux-arm64"
cp -f "${REPO_DIR}/build_assets/V2bX-linux-amd64" "${PUB_DIR}/V2bX-linux-amd64"

echo "========================================================="
echo ">>> [3/5] 脚本语法自检..."
echo "========================================================="
bash -n "${REPO_DIR}/v2bx.sh"

echo "========================================================="
echo ">>> [4/5] 提交本地 Git 并推送至 GitHub main..."
echo "========================================================="
if [[ ! -f "${TOKEN_FILE}" ]]; then
    echo "错误：未找到 token 文件 ${TOKEN_FILE}" >&2
    exit 1
fi
GH_TOKEN=$(tr -d '\n\r ' < "${TOKEN_FILE}")

git add -A
if ! git diff --cached --quiet; then
    git commit -m "${COMMIT_MSG}"
else
    echo "工作区无变动，跳过 commit。"
fi

git push "https://${GH_TOKEN}@github.com/${REMOTE_REPO}.git" main

echo "========================================================="
echo ">>> [5/5] 同步更新 GitHub Releases 官方发布资产..."
echo "========================================================="
python3 - << 'PYEOF'
import urllib.request
import json
import os
import sys

TOKEN_FILE = "/root/.openclaw/workspace/keys/github_token"
REPO = "zwgundam/V2bX-Patched-By-Haman"

with open(TOKEN_FILE, "r") as f:
    token = f.read().strip()

headers = {
    "Authorization": f"token {token}",
    "Accept": "application/vnd.github.v3+json",
    "User-Agent": "OpenClaw-Release-Updater"
}

req = urllib.request.Request(f"https://api.github.com/repos/{REPO}/releases/latest", headers=headers)
try:
    with urllib.request.urlopen(req) as resp:
        release = json.loads(resp.read().decode())
except Exception as e:
    print(f"获取最新 Release 失败: {e}", file=sys.stderr)
    sys.exit(1)

print(f"目标 Release: {release['name']} (ID: {release['id']})")

for asset in release.get("assets", []):
    if asset["name"] in ["v2bx-linux-amd64", "v2bx-linux-arm64", "v2bx.sh"]:
        print(f"正在删除旧资产: {asset['name']} (ID: {asset['id']})...")
        del_req = urllib.request.Request(f"https://api.github.com/repos/{REPO}/releases/assets/{asset['id']}", headers=headers, method="DELETE")
        try:
            with urllib.request.urlopen(del_req) as del_resp:
                pass
        except Exception as e:
            print(f"删除资产 {asset['name']} 失败: {e}")

upload_url_base = release["upload_url"].split("{")[0]
files_to_upload = [
    ("v2bx-linux-amd64", "/root/.openclaw/workspace/v2bx-custom/build_assets/V2bX-linux-amd64", "application/octet-stream"),
    ("v2bx-linux-arm64", "/root/.openclaw/workspace/v2bx-custom/build_assets/V2bX-linux-arm64", "application/octet-stream"),
    ("v2bx.sh", "/root/.openclaw/workspace/v2bx-custom/v2bx.sh", "text/x-shellscript")
]

for name, path, mime in files_to_upload:
    print(f"正在上传新资产 {name} ({os.path.getsize(path)} bytes)...")
    with open(path, "rb") as f:
        data = f.read()
    
    url = f"{upload_url_base}?name={name}"
    up_headers = dict(headers)
    up_headers["Content-Type"] = mime
    up_headers["Content-Length"] = str(len(data))
    
    req = urllib.request.Request(url, data=data, headers=up_headers, method="POST")
    with urllib.request.urlopen(req) as up_resp:
        res = json.loads(up_resp.read().decode())
        print(f"--> 上传成功: {name} (ID: {res['id']})")

print("Release 资产全量对齐完成！")
PYEOF

echo "========================================================="
echo "✅ [SUCCESS] V2bX 标准流水线发布全部成功！"
echo "========================================================="
