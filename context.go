package keyman

import (
	"fmt"

	"github.com/boltdb/bolt"
)

type bucketFactory interface {
	Writable() bool
	Bucket(name []byte) *bolt.Bucket
	CreateBucketIfNotExists(name []byte) (*bolt.Bucket, error)
}

type Context interface {
	OpenRevisionBucket() (RevisionBucket, error)
	OpenSecretBucket(name string) (SecretBucket, error)
	OpenHistoryBucket(name string) (HistoryBucket, error)
}

type context struct {
	tx *bolt.Tx
}

func openBucket(f bucketFactory, name string) (*bolt.Bucket, error) {
	if f.Writable() {
		b, err := f.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return nil, err
		}
		return b, nil
	}

	b := f.Bucket([]byte(name))
	if b == nil {
		return nil, fmt.Errorf("keyman: no bucket: %s", name)
	}
	return b, nil
}

func (c *context) OpenRevisionBucket() (RevisionBucket, error) {
	b, err := openBucket(c.tx, "rev")
	if err != nil {
		return nil, err
	}
	return &revision{bucket: b}, nil
}

func (c *context) OpenSecretBucket(name string) (SecretBucket, error) {
	p, err := openBucket(c.tx, "sec")
	if err != nil {
		return nil, err
	}
	b, err := openBucket(p, name)
	if err != nil {
		return nil, err
	}
	return &secretBucket{bucket: b}, nil
}

func (c *context) OpenHistoryBucket(name string) (HistoryBucket, error) {
	p, err := openBucket(c.tx, "hist")
	if err != nil {
		return nil, err
	}
	b, err := openBucket(p, name)
	if err != nil {
		return nil, err
	}
	return &historyBucket{bucket: b}, nil
}
