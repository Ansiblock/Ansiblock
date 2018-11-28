package network

import (
	"net"
	"time"
)

// SocketMock is socket connection mock for testing
type SocketMock struct {
	readError     error
	writeError    error
	readBuff      chan []byte
	writeBuff     chan []byte
	readBuffSize  int
	writeBuffSize int
	Addr          net.Addr
}

func NewSocketMock(readError error, writeError error, addr net.Addr) *SocketMock {
	sm := new(SocketMock)
	sm.readError = readError
	sm.writeError = writeError
	sm.readBuff = make(chan []byte, 1000)
	sm.writeBuff = make(chan []byte, 1000)
	sm.readBuffSize = 0
	sm.writeBuffSize = 0
	sm.Addr = addr

	return sm
}

func (conn *SocketMock) ReadBuffSize() int {
	return conn.readBuffSize
}

func (conn *SocketMock) WriteBuffSize() int {
	return conn.writeBuffSize
}

func (conn *SocketMock) AddToReadBuff(b []byte) {
	conn.readBuff <- b
	conn.readBuffSize += len(b)
}

func (conn *SocketMock) EmptyWriteBuff() []byte {
	data := make([]byte, 0, conn.writeBuffSize)
	for b := range conn.writeBuff {
		data = append(data, b...)
		conn.writeBuffSize -= len(b)
		if conn.writeBuffSize == 0 {
			break
		}
	}
	return data
}

//ReadFrom implements the PacketConn ReadFrom mock method.
func (conn *SocketMock) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	if conn.readError != nil || conn.readBuffSize <= 0 {
		return 0, conn.Addr, conn.readError
	}
	data, ok := <-conn.readBuff
	if !ok {
		return 0, conn.Addr, conn.readError
	}

	copy(b, data)
	conn.readBuffSize -= len(data)
	return len(data), conn.Addr, conn.readError
}

//WriteTo implements the PacketConn WriteTo mock method.
func (conn *SocketMock) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	if conn.writeError != nil {
		return 0, conn.writeError
	}

	conn.writeBuff <- b
	conn.writeBuffSize += len(b)
	conn.Addr = addr
	return len(b), conn.writeError
}

//Close implements the PacketConn Close mock method.
func (conn *SocketMock) Close() error {
	return nil
}

//LocalAddr implements the PacketConn LocalAddr mock method.
func (conn *SocketMock) LocalAddr() net.Addr {
	return nil
}

//SetDeadline implements the PacketConn SetDeadline mock method.
func (conn *SocketMock) SetDeadline(t time.Time) error {
	return nil
}

//SetReadDeadline implements the PacketConn SetReadDeadline mock method.
func (conn *SocketMock) SetReadDeadline(t time.Time) error {
	return nil
}

//SetWriteDeadline implements the PacketConn SetWriteDeadline mock method.
func (conn *SocketMock) SetWriteDeadline(t time.Time) error {
	return nil
}
