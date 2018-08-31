package keyman

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

type Metadata struct {
	CreatedAt time.Time
}

type HistoryCursor interface {
	First() ([]byte, *Metadata)
	Last() ([]byte, *Metadata)
	Prev() ([]byte, *Metadata)
	Next() ([]byte, *Metadata)
}

type historyCursor struct {
	HistoryCursor
	cursor *bolt.Cursor
}

func (c *historyCursor) unmarshal(k, v []byte) ([]byte, *Metadata) {
	var d Metadata
	if err := json.Unmarshal(v, &d); err != nil {
		return nil, nil
	}
	return k, &d
}

func (c *historyCursor) First() ([]byte, *Metadata) {
	return c.unmarshal(c.cursor.First())
}

func (c *historyCursor) Last() ([]byte, *Metadata) {
	return c.unmarshal(c.cursor.Last())
}

func (c *historyCursor) Prev() ([]byte, *Metadata) {
	return c.unmarshal(c.cursor.Prev())
}

func (c *historyCursor) Next() ([]byte, *Metadata) {
	return c.unmarshal(c.cursor.Next())
}

type HistoryBucket interface {
	Get(rev []byte) (*Metadata, error)
	Push(m *Metadata) ([]byte, error)
	Cursor() HistoryCursor
}

type historyBucket struct {
	HistoryBucket
	bucket *bolt.Bucket
}

func (b *historyBucket) Get(rev []byte) (*Metadata, error) {
	v := b.bucket.Get(rev)
	if v == nil {
		return nil, errors.New("keyman: no history")
	}

	var m Metadata
	if err := json.Unmarshal(v, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (b *historyBucket) Push(h *Metadata) ([]byte, error) {
	s, err := b.bucket.NextSequence()
	if err != nil {
		return nil, err
	}

	v, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}

	rev := make([]byte, 8)
	binary.BigEndian.PutUint64(rev, s)
	return rev, b.bucket.Put(rev, v)
}

func (b *historyBucket) Cursor() HistoryCursor {
	c := b.bucket.Cursor()
	return &historyCursor{cursor: c}
}
