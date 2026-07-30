#!/usr/bin/env bash
set -euo pipefail

if [[ $EUID -ne 0 ]]; then
    echo "запусти от root: sudo bash $0"
    exit 1
fi

ensure_pkg() {
    local pkg=$1
    command -v "$pkg" >/dev/null 2>&1 && return
    echo "==> установка $pkg..."
    apt-get update -qq && apt-get install -y -qq "$pkg"
}

ensure_pkg git
ensure_pkg curl

install_go() {
    local ver arch
    ver=$(curl -s https://go.dev/VERSION?m=text 2>/dev/null | head -1)
    ver="${ver#go}"
    [[ -n "$ver" ]] || ver="1.24.0"
    arch="linux-amd64"

    echo "==> установка Go $ver..."
    curl -sSL "https://go.dev/dl/go${ver}.${arch}.tar.gz" | tar -C /usr/local -xz
    export PATH="/usr/local/go/bin:$PATH"
}

if ! command -v go >/dev/null 2>&1; then
    install_go
fi

REPO="https://github.com/nkinash/socks5-proxy.git"
BIN_PATH="/usr/local/bin/sock5"

detect_systemd_dir() {
    local dir
    for dir in /etc/systemd/system /usr/lib/systemd/system /lib/systemd/system; do
        [[ -d "$dir" ]] && echo "$dir" && return
    done
    echo "/etc/systemd/system"
}
SVC_DIR=$(detect_systemd_dir)
SVC_FILE="$SVC_DIR/sock5.service"

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

pick_port() {
    local port=$1
    while ss -tlnp 2>/dev/null | grep -q ":$port "; do
        port=$((port + 1))
    done
    echo "$port"
}
PORT=$(pick_port 1080)

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

# --- проверка статуса ---
echo "==> проверка статуса сервиса..."
for i in $(seq 1 10); do
    if systemctl is-active --quiet sock5 2>/dev/null; then
        echo "==> сервис активен"
        break
    fi
    sleep 0.5
done

if ! systemctl is-active --quiet sock5 2>/dev/null; then
    echo "==> ОШИБКА: сервис не запустился"
    echo "==> проверь логи: journalctl -u sock5 -n 20"
    exit 1
fi

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
