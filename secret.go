package keyman

import (
	"github.com/boltdb/bolt"
)

// SecretBucket stores secret values.
type SecretBucket interface {
	// Get returns the secret value.
	Get(rev []byte) []byte

	// Push save the secret value and returns the revision number.
	Put(rev, v []byte) error
}

type secretBucket struct {
	SecretBucket
	bucket *bolt.Bucket
}

func (s *secretBucket) Get(rev []byte) []byte {
	return s.bucket.Get(rev)
}

func (s *secretBucket) Put(rev, v []byte) error {
	return s.bucket.Put(rev, v)
}
