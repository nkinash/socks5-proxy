# sock5

## Команды

```bash
go run ./cmd/sock5/      # запуск
go build -o sock5 ./cmd/sock5/ # сборка
go vet ./...              # статический анализ
```

## Структура

Одномодульный Go-проект, DDD-слои:

```
cmd/sock5/        точка входа, композиция зависимостей
internal/
  domain/         типы, константы протокола SOCKS5, интерфейс Dialer
  protocol/       wire-формат: разбор/запись handshake, request, reply
  service/        бизнес-логика прокси: handshake → connect → relay
  infra/          TCP-сервер, приём соединений
```

Порядок зависимостей: `domain` ← `protocol` ← `service` ← `infra` ← `cmd`.
