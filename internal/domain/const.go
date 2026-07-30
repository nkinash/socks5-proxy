package domain

// Версия протокола
const Version = 0x05

// Методы аутентификации
const (
	AuthNone     = 0x00
	AuthUserPass = 0x02
	AuthNoAccep  = 0xFF
)

// Команды
const (
	CmdConnect = 0x01
	CmdBind    = 0x02
	CmdUDP     = 0x03
)

// Типы адресов
const (
	AtypIPv4   = 0x01
	AtypDomain = 0x03
	AtypIPv6   = 0x04
)

// Коды ответов
const (
	RepSuccess              = 0x00
	RepGeneralFailure       = 0x01
	RepNotAllowed           = 0x02
	RepNetworkUnreachable   = 0x03
	RepHostUnreachable      = 0x04
	RepConnRefused          = 0x05
	RepTTLExpired           = 0x06
	RepCmdNotSupported      = 0x07
	RepAddrTypeNotSupported = 0x08
)

// User/pass-подсогласование (RFC 1929)
const (
	UserPassVer  = 0x01
	UserPassOK   = 0x00
	UserPassFail = 0x01
)

// Размеры полей
const (
	IPv4Len = 4
	IPv6Len = 16
)

// RSV — зарезервированный байт в запросах и ответах
const RSV = 0x00
