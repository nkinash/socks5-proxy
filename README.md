# sock5

SOCKS5-прокси на Go, один бинарник.

## Быстрая установка

```bash
wget -qO- https://raw.githubusercontent.com/nkinash/socks5-proxy/master/install.sh | sudo bash
```

Требования: Go, git, curl, systemd (Ubuntu/Debian).

## Установка из репозитория

```bash
git clone https://github.com/nkinash/socks5-proxy.git
cd socks5-proxy
sudo bash install.sh
```

Скрипт соберёт бинарник, сгенерирует логин/пароль, определит IP и запустит systemd-сервис.

## Ручная установка

### 1. Сборка

На машине с Go 1.26+:

```bash
git clone https://github.com/<user>/sock5.git
cd sock5
go build -o sock5 ./cmd/sock5/
```

Перенеси бинарник на сервер:

```bash
scp sock5 user@server:/tmp/
```

### 2. Установка бинарника

```bash
sudo cp /tmp/sock5 /usr/local/bin/sock5
sudo chmod +x /usr/local/bin/sock5
```

### 3. Создание systemd-сервиса

```bash
sudo tee /etc/systemd/system/sock5.service <<'EOF'
[Unit]
Description=SOCKS5 proxy
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/sock5 -addr :1080 -user ЛОГИН -pass ПАРОЛЬ
Restart=always
RestartSec=5
User=nobody
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
```

Замени `ЛОГИН` и `ПАРОЛЬ` на свои.

### 4. Запуск

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now sock5
```

### 5. Проверка

```bash
systemctl status sock5
journalctl -u sock5 -f
```

## Подключение

Без аутентификации:

```bash
./sock5 -addr :1080
```

С аутентификацией:

```bash
./sock5 -addr :1080 -user admin -pass secret
```

Строка подключения:

```
socks5://логин:пароль@IP:1080
```

IP сервера можно узнать командой:

```bash
curl -s ifconfig.me
```

## Флаги

| Флаг | По умолчанию | Описание |
|------|-------------|----------|
| `-addr` | `:1080` | адрес для прослушивания |
| `-user` | — | логин (если задан — включает аутентификацию) |
| `-pass` | — | пароль (обязателен вместе с `-user`) |

## Тесты

```bash
go test -v -count=1 -run TestProxyIntegration ./internal/infra/
```
