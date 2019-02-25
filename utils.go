package keyman

import (
	"encoding/binary"
)

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func ref(name string, rev []byte) []byte {
	return append([]byte(name+":"), rev...)
}
