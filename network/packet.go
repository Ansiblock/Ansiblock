package network

import (
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/Ansiblock/Ansiblock/log"
)

const (
	packetDataSize = 256
	// NumPackets represents number of packets in Packets struct
	NumPackets  = 1024 * 8
	readTimeout = 120 * time.Millisecond //TODO to be determined
)

// Packet stores transaction data
type Packet struct {
	Data [packetDataSize]byte
	Size uint8
	Addr net.Addr //UDP Address
	_    [16]byte //padding
}

// Packets stores slice of packets
type Packets struct {
	Ps []Packet
}

// NewNumPackets returns Packets object with num packets
func NewNumPackets(num uint64) *Packets {
	res := make([]Packet, num)
	return &Packets{Ps: res}
}

// NewPackets returns Packets object
func NewPackets() *Packets {
	res := make([]Packet, NumPackets)
	return &Packets{Ps: res}
}

// ReadFrom reads from the UDP socket and return number of bytes read.
// func (ps *Packets) ReadFrom(reader net.PacketConn) int {
// 	deadlineCount := 0
// 	count := 0
// 	for i := range ps.Ps {
// 		// readerErr := reader.SetReadDeadline(time.Now().Add(readTimeout))
// 		n, addr, err := reader.ReadFrom(ps.Ps[i].Data[:])
// 		ps.Ps[i].Size = uint8(n)
// 		ps.Ps[i].Addr = addr
// 		count++

// 		if err != nil || n == 0 {
// 			if err, ok := err.(net.Error); ok && err.Timeout() {
// 				if deadlineCount == 0 {
// 					fmt.Printf("\n\ncount: %v\n\n", count)
// 					count = 0
// 					deadlineCount++
// 					reader.SetReadDeadline(time.Now().Add(readTimeout))
// 				} else {
// 					ps.Ps = ps.Ps[:i]
// 					reader.SetReadDeadline(time.Now().Add(1 * time.Hour))
// 					return 1
// 				}
// 			} else {
// 				ps.Ps = ps.Ps[:i]
// 				reader.SetReadDeadline(time.Now().Add(1 * time.Hour))
// 				return 1
// 			}
// 		} else {
// 			deadlineCount = 0
// 		}
// 		// fmt.Printf("timeout: %v\n", time.Now().Add(readTimeout))
// 		if deadlineCount == 0 {
// 			readerErr := reader.SetReadDeadline(time.Now().Add(readTimeout))
// 			if readerErr != nil {
// 				ps.Ps = ps.Ps[:i+1]
// 				fmt.Printf("error %v. time: %v\n", err, time.Now())
// 				reader.SetReadDeadline(time.Now().Add(1 * time.Hour))
// 				return 1
// 			}
// 		}
// 	}
// 	reader.SetReadDeadline(time.Now().Add(1 * time.Hour))
// 	return 1
// }

// ReadFrom reads from the UDP socket and return number of bytes read.
func (ps *Packets) ReadFrom(reader net.PacketConn) int {
	for i := range ps.Ps {
		n, addr, err := reader.ReadFrom(ps.Ps[i].Data[:])
		ps.Ps[i].Size = uint8(n)
		ps.Ps[i].Addr = addr
		if err != nil || n == 0 {
			ps.Ps = ps.Ps[:i]
			reader.SetReadDeadline(time.Now().Add(1 * time.Hour))
			return i
		}
		if i == 0 {
			reader.SetReadDeadline(time.Now().Add(readTimeout))
		}
	}
	reader.SetReadDeadline(time.Now().Add(1 * time.Hour))
	return len(ps.Ps)
}

// WriteTo writes all packets into the socket
func (ps *Packets) WriteTo(writer net.PacketConn) int {
	res := 0
	for i := range ps.Ps {
		n, err := writer.WriteTo(ps.Ps[i].Data[:ps.Ps[i].Size], ps.Ps[i].Addr)
		if err != nil {
			log.Error("Packets WriteTo error", zap.Error(err))
		}
		res += n
	}
	return res
}

// Write writes all packets into the socket
// func (ps *Packets) Write(writer net.Conn) int {
// 	res := 0
// 	for i := range ps.Ps {
// 		n, err := writer.Write(ps.Ps[i].Data[:ps.Ps[i].Size])
// 		if err != nil {
// 			log.Error("Packets Write error", zap.Error(err))
// 			fmt.Printf("Error: %v\n", err)
// 		}
// 		res += n
// 	}
// 	return res
// }
