#!/usr/bin/env bash
set -euo pipefail

if [[ $EUID -ne 0 ]]; then
    echo "запусти от root: sudo bash $0"
    exit 1
fi

command -v go >/dev/null 2>&1  || { echo "установи Go (https://go.dev/doc/install)"; exit 1; }
command -v git >/dev/null 2>&1 || { echo "установи git"; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "установи curl"; exit 1; }

REPO="https://github.com/nkinash/socks5-proxy.git"
BIN_PATH="/usr/local/bin/sock5"
SVC_FILE="/etc/systemd/system/sock5.service"

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

echo "==> клонирование..."
git clone -q "$REPO" "$TMPDIR"

echo "==> сборка..."
cd "$TMPDIR"
go build -o "$BIN_PATH" ./cmd/sock5/

# --- генерация логина и пароля ---
USER="proxy_$(head -c4 /dev/urandom | base64 | tr -dc 'a-z0-9' | head -c6)"
PASS="$(head -c12 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c16)"

# --- определение IP ---
detect_ip() {
    local ip
    ip=$(curl -s --max-time 3 ifconfig.me 2>/dev/null) && echo "$ip" && return
    ip=$(curl -s --max-time 3 icanhazip.com 2>/dev/null) && echo "$ip" && return
    ip=$(curl -s --max-time 3 ipinfo.io/ip 2>/dev/null) && echo "$ip" && return
    ip=$(hostname -I 2>/dev/null | awk '{print $1}') && echo "$ip" && return
    echo "127.0.0.1"
}

IP=$(detect_ip)
PORT=1080

# --- systemd unit ---
cat > "$SVC_FILE" <<EOF
[Unit]
Description=SOCKS5 proxy
After=network.target

[Service]
Type=simple
ExecStart=$BIN_PATH -addr :$PORT -user $USER -pass $PASS
Restart=always
RestartSec=5
User=nobody
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# --- запуск ---
systemctl daemon-reload
systemctl enable --now sock5

echo ""
echo "============================================"
echo "  SOCKS5 прокси запущен"
echo "============================================"
echo "  адрес  : $IP"
echo "  порт   : $PORT"
echo "  логин  : $USER"
echo "  пароль : $PASS"
echo ""
echo "  socks5://$USER:$PASS@$IP:$PORT"
echo "============================================"
echo ""
echo "управление:"
echo "  systemctl status sock5"
echo "  systemctl stop   sock5"
echo "  systemctl restart sock5"
echo "  journalctl -u sock5 -f"
