package keyman

import "testing"

func TestItobLength(t *testing.T) {
	bytes := itob(0)
	l := len(bytes)
	if l != 8 {
		t.Fatalf("expect 8, but actual %d", l)
	}
}

func TestItobEndian(t *testing.T) {
	for i := 0; i < 8; i++ {
		bytes := itob(255 << uint(56-8*i))
		if bytes[i] != byte(255) {
			t.Fatalf("%v", bytes)
		}
	}
}

func TestBtoi(t *testing.T) {
	for i := 0; i < 8; i++ {
		n := uint64(255) << uint(56-8*i)
		bytes := itob(n)
		if btoi(bytes) != n {
			t.Fatalf("%v", bytes)
		}
	}

}
