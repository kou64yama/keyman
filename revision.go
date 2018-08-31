package keyman

import (
	"github.com/boltdb/bolt"
)

type RevisionBucket interface {
	Get(name string) []byte
	Put(name string, rev []byte) error
	ForEach(fn func(name string, rev []byte) error) error
}

type revision struct {
	RevisionBucket
	bucket *bolt.Bucket
}

func (b *revision) Get(name string) []byte {
	return b.bucket.Get([]byte(name))
}

func (b *revision) Put(name string, rev []byte) error {
	return b.bucket.Put([]byte(name), rev)
}

func (b *revision) ForEach(fn func(name string, rev []byte) error) error {
	return b.bucket.ForEach(func(k, v []byte) error {
		return fn(string(k), v)
	})
}
