// Package block implements generation and maintenance of the Verifiable Delayed Function Values
// Ansiblock will be using VDFs for computational timestamping. Incremental VDFs can provide
// computational evidence that a given version of the state’s system is older
// (and therefore genuine) by proving that a long-running VDF computation has been performed
// on the genuine history just after the point of divergence with the fraudulent history.
// This potentially enables detecting long-range forks without relying on external
// timestamping mechanisms.
package block

import (
	"crypto/sha256"
)

// VDFValue represents the value of Verifiable Delayed Function.
type VDFValue = []byte

// VDFSize represents size of vdf value in bytes
const VDFSize int = sha256.Size

// VDF returns the SHA256 checksum of the data
func VDF(data []byte) VDFValue {
	res := sha256.Sum256(data)
	return res[:]
}

// ExtendedVDF returns the SHA256 of the data + previous VDFValue
func ExtendedVDF(data []byte, val VDFValue) VDFValue {
	appended := append(val, data...)
	return VDF(appended)
}
