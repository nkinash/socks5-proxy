package protocol

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/nkinash/socks5-proxy/internal/domain"
)

func ReadHandshake(r io.Reader) ([]byte, error) {
	head := make([]byte, 2)
	if _, err := io.ReadFull(r, head); err != nil {
		return nil, fmt.Errorf("read handshake head: %w", err)
	}
	if head[0] != domain.Version {
		return nil, fmt.Errorf("bad version: %d", head[0])
	}
	n := int(head[1])
	if n == 0 {
		return nil, nil
	}
	methods := make([]byte, n)
	if _, err := io.ReadFull(r, methods); err != nil {
		return nil, fmt.Errorf("read methods: %w", err)
	}
	return methods, nil
}

func WriteHandshake(w io.Writer, method byte) error {
	_, err := w.Write([]byte{domain.Version, method})
	return err
}

func ReadUserPass(r io.Reader) (string, string, error) {
	ver := make([]byte, 1)
	if _, err := io.ReadFull(r, ver); err != nil {
		return "", "", fmt.Errorf("read userpass ver: %w", err)
	}
	if ver[0] != domain.UserPassVer {
		return "", "", fmt.Errorf("bad userpass ver: %d", ver[0])
	}

	ulen := make([]byte, 1)
	if _, err := io.ReadFull(r, ulen); err != nil {
		return "", "", fmt.Errorf("read ulen: %w", err)
	}
	userBytes := make([]byte, ulen[0])
	if _, err := io.ReadFull(r, userBytes); err != nil {
		return "", "", fmt.Errorf("read username: %w", err)
	}

	plen := make([]byte, 1)
	if _, err := io.ReadFull(r, plen); err != nil {
		return "", "", fmt.Errorf("read plen: %w", err)
	}
	passBytes := make([]byte, plen[0])
	if _, err := io.ReadFull(r, passBytes); err != nil {
		return "", "", fmt.Errorf("read password: %w", err)
	}

	return string(userBytes), string(passBytes), nil
}

func WriteUserPassReply(w io.Writer, ok bool) error {
	status := byte(domain.UserPassFail)
	if ok {
		status = domain.UserPassOK
	}
	_, err := w.Write([]byte{domain.UserPassVer, status})
	return err
}

func ReadRequest(r io.Reader) (*domain.Request, error) {
	head := make([]byte, 4)
	if _, err := io.ReadFull(r, head); err != nil {
		return nil, fmt.Errorf("read request head: %w", err)
	}
	if head[0] != domain.Version {
		return nil, fmt.Errorf("bad version: %d", head[0])
	}

	cmd := head[1]
	atyp := head[3]

	host, err := readHost(r, atyp)
	if err != nil {
		return nil, err
	}

	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(r, portBytes); err != nil {
		return nil, fmt.Errorf("read port: %w", err)
	}
	port := binary.BigEndian.Uint16(portBytes)

	return &domain.Request{
		Cmd: cmd,
		Addr: domain.Addr{
			Type: atyp,
			Host: host,
			Port: port,
		},
	}, nil
}

func WriteReply(w io.Writer, rep byte) error {
	buf := []byte{domain.Version, rep, domain.RSV, domain.AtypIPv4, 0, 0, 0, 0, 0, 0}
	_, err := w.Write(buf)
	return err
}

func readHost(r io.Reader, atyp byte) (string, error) {
	switch atyp {
	case domain.AtypIPv4:
		buf := make([]byte, domain.IPv4Len)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", fmt.Errorf("read ipv4: %w", err)
		}
		return fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3]), nil
	case domain.AtypDomain:
		lenb := make([]byte, 1)
		if _, err := io.ReadFull(r, lenb); err != nil {
			return "", fmt.Errorf("read domain len: %w", err)
		}
		buf := make([]byte, lenb[0])
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", fmt.Errorf("read domain: %w", err)
		}
		return string(buf), nil
	case domain.AtypIPv6:
		buf := make([]byte, domain.IPv6Len)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", fmt.Errorf("read ipv6: %w", err)
		}
		return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x:%02x",
			buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7],
			buf[8], buf[9], buf[10], buf[11], buf[12], buf[13], buf[14], buf[15]), nil
	default:
		return "", fmt.Errorf("unknown address type: %d", atyp)
	}
}
