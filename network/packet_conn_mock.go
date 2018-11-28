package network

import (
	"errors"
	"net"
	"time"
)

// PCon mock
type PCon struct {
	B []byte
}

//ReadFrom implements the PacketConn ReadFrom mock method.
func (p *PCon) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	var ad *net.UDPAddr
	ad, err = net.ResolveUDPAddr("udp", "127.0.0.1:8000")
	if b[0] == 0 {
		b[0] = 1
		return 1, ad, nil
	} else if b[0] == 1 {
		return packetDataSize + 1, nil, nil
	}
	return 0, nil, errors.New("bla")
}

//WriteTo implements the PacketConn WriteTo mock method.
func (p *PCon) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	if p.B[0] == 0 {
		return 7, nil
	}
	return 0, errors.New("bla")
}

//Close implements the PacketConn Close mock method.
func (p *PCon) Close() error {
	return nil
}

//LocalAddr implements the PacketConn LocalAddr mock method.
func (p *PCon) LocalAddr() net.Addr {
	return nil
}

//SetDeadline implements the PacketConn SetDeadline mock method.
func (p *PCon) SetDeadline(t time.Time) error {
	return nil
}

//SetReadDeadline implements the PacketConn SetReadDeadline mock method.
func (p *PCon) SetReadDeadline(t time.Time) error {
	if p.B[9] == 1 {
		return errors.New("bla")
	}
	return nil
}

//SetWriteDeadline implements the PacketConn SetWriteDeadline mock method.
func (p *PCon) SetWriteDeadline(t time.Time) error {
	return nil
}
