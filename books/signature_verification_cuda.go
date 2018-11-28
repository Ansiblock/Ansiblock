// +build cuda

package books

/*
#cgo CFLAGS: -I/usr/local/cuda/include/
#cgo LDFLAGS: ${SRCDIR}/libcuda_verify_ed25519.a
#cgo LDFLAGS: /usr/local/cuda/lib64/libcudart.so
#cgo LDFLAGS: /usr/local/cuda/lib64/libcudadevrt.a
#include <cuda.h>
#include <cuda_runtime.h>
#include <cuda_device_runtime_api.h>
void ed25519_set_verbose(_Bool b);
_Bool ed25519_init();
typedef struct{
	char *elems;
	unsigned int num;
}Elems;
unsigned int ed25519_verify_many(
	Elems* elems,
	unsigned int num,
	unsigned int message_size,
	unsigned int public_key_offset,
	unsigned int signature_offset,
	unsigned int signed_message_offset,
	unsigned int signed_message_len_offset,
	unsigned char *out
);
*/
import "C"

import (
	"unsafe"

	"github.com/Ansiblock/Ansiblock/log"
	"github.com/Ansiblock/Ansiblock/network"
	"go.uber.org/zap"
)

const txOffset = 0
const publicKeyOffset = 64
const signatureOffset = 0
const signedMessageOffset = 64
const signedMessageLenOffset = 256
const packetDataSize = 256

func init() {
	_, err := C.ed25519_set_verbose(true)
	if err != nil {
		log.Error("can not initialize ed25519 cuda library")
	}

	if !C.ed25519_init() {
		log.Panic("ed25519_init() failed")
	}
	C.ed25519_set_verbose(false)
}

// verifies packets on nVidia CUDA
// verifyPackets recives array of pointers to network packets.
// each packet contains several transaction. The signature of each transation is verified.
// If signature is not valied, packets size is set to zero.
func verifyPackets(packets []*network.Packets) ([]*network.Packets, int) {
	if len(packets) == 0 {
		return packets, 0
	}

	elems := C.malloc(C.size_t(len(packets)) * C.size_t(C.sizeof_Elems))
	defer C.free(elems)
	elemsArray := (*[1<<30 - 1]C.Elems)(elems)
	num := 0
	length := 0
	for i, p := range packets {
		numOfPackets := len(p.Ps)
		elem := C.Elems{
			elems: (*C.char)(unsafe.Pointer(&p.Ps[0])),
			num:   C.uint(numOfPackets),
		}
		elemsArray[i] = elem
		num += numOfPackets
		length++
	}
	packetSize := int(unsafe.Sizeof(network.Packet{}))
	// fmt.Println(elemsArray[:length])
	// fmt.Println("Starting verify num packets: ", num)
	// fmt.Println("elem len: ", length)
	// fmt.Println("packet sizeof: ", packetSize)
	// fmt.Println("pub key: ", txOffset+publicKeyOffset)
	// fmt.Println("sig offset: ", txOffset+signatureOffset)
	// fmt.Println("sign data: ", txOffset+signedMessageOffset)
	// fmt.Println("len offset: ", packetDataSize)
	log.Debug("Starting verify num packets: ", zap.Int("", num))
	log.Debug("elem len: ", zap.Int("", length))
	log.Debug("packet sizeof: ", zap.Int("", packetSize))
	log.Debug("pub key: ", zap.Int("", txOffset+publicKeyOffset))
	log.Debug("sig offset: ", zap.Int("", txOffset+signatureOffset))
	log.Debug("sign data: ", zap.Int("", txOffset+signedMessageOffset))
	log.Debug("len offset: ", zap.Int("", packetDataSize))

	out := make([]byte, num)
	res := C.ed25519_verify_many(
		(*C.Elems)(unsafe.Pointer(elems)),
		C.uint(length),
		C.uint(packetSize),
		C.uint(txOffset+publicKeyOffset),
		C.uint(txOffset+signatureOffset),
		C.uint(txOffset+signedMessageOffset),
		C.uint(signedMessageLenOffset),
		(*C.uchar)(&out[0]))

	if res != 0 {
		log.Info("GPU returned error code: ", zap.Int("", int(res)))
	}

	ans := 0
	index := 0
	for i := 0; i < len(packets); i++ {
		for j := 0; j < len(packets[i].Ps); j++ {
			if out[index] == 0 {
				packets[i].Ps[j].Size = 0
			} else {
				ans += 1
			}
			index += 1
		}
	}
	return packets, ans
}
