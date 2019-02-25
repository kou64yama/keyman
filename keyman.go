package keyman

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

var (
	ErrNoBlob = errors.New("fatal: no blob")
	ErrNoRefs = errors.New("fatal: no refs")
)

type Metadata struct {
	CTime  time.Time      `json:"ctime"`
	Length int            `json:"length"`
	Md5    [md5.Size]byte `json:"md5"`
}

type Keyman interface {
	Close() error
	Get(name string) ([]byte, error)
	Set(name string, data []byte) (uint64, error)
	Revert(name string, rev uint64) error
	Delete(name string) error
	Metadata(name string) (uint64, *Metadata, error)
	ForEach(fn func(name string, rev uint64) error) error
	History(name string, limit uint64, fn func(rev uint64, md *Metadata) error) error
}

func Open(dir string) (Keyman, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	db, err := bolt.Open(path.Join(dir, "keyman.db"), 0600, nil)
	if err != nil {
		return nil, err
	}

	k := &keyman{
		db: db,
	}
	return k, nil
}

type keyman struct {
	Keyman
	db *bolt.DB
}

func (k *keyman) Close() error {
	return k.db.Close()
}

func (k *keyman) Get(name string) ([]byte, error) {
	var data []byte
	err := k.db.View(func(tx *bolt.Tx) error {
		refs := tx.Bucket([]byte("refs"))
		if refs == nil {
			return fmt.Errorf("no password: %s", name)
		}

		blob := tx.Bucket([]byte("blob"))
		if blob == nil {
			return ErrNoBlob
		}

		data = blob.Get(ref(name, refs.Get([]byte(name))))
		return nil
	})
	return data, err
}

func (k *keyman) Set(name string, data []byte) (uint64, error) {
	if data == nil {
		return 0, errors.New("password is empty")
	}

	var rev uint64
	err := k.db.Update(func(tx *bolt.Tx) error {
		refs, err := tx.CreateBucketIfNotExists([]byte("refs"))
		if err != nil {
			return err
		}
		hist, err := tx.CreateBucketIfNotExists([]byte("history:" + name))
		if err != nil {
			return err
		}
		blob, err := tx.CreateBucketIfNotExists([]byte("blob"))
		if err != nil {
			return err
		}

		r, err := hist.NextSequence()
		if err != nil {
			return err
		}
		rev = r

		md := &Metadata{
			CTime:  time.Now().UTC(),
			Length: len(data),
			Md5:    md5.Sum(data),
		}
		bytes, err := json.Marshal(md)
		if err != nil {
			return err
		}

		if err := refs.Put([]byte(name), itob(rev)); err != nil {
			return err
		}
		if err := hist.Put(itob(rev), bytes); err != nil {
			return err
		}
		if err := blob.Put(ref(name, itob(rev)), data); err != nil {
			return err
		}

		return nil
	})
	return rev, err
}

func (k *keyman) Revert(name string, rev uint64) error {
	return k.db.Update(func(tx *bolt.Tx) error {
		blob := tx.Bucket([]byte("blob"))
		if blob == nil {
			return fmt.Errorf("no password: %s (rev %d)", name, rev)
		}

		data := blob.Get(ref(name, itob(rev)))
		if data == nil {
			return fmt.Errorf("no password: %s (rev %d)", name, rev)
		}

		refs := tx.Bucket([]byte("refs"))
		if refs == nil {
			return ErrNoRefs
		}

		return refs.Put([]byte(name), itob(rev))
	})
}

func (k *keyman) Delete(name string) error {
	return k.db.Update(func(tx *bolt.Tx) error {
		refs := tx.Bucket([]byte("refs"))
		if refs != nil {
			refs.Delete([]byte(name))
		}

		return nil
	})
}

func (k *keyman) Metadata(name string) (uint64, *Metadata, error) {
	var rev uint64
	var md Metadata
	return rev, &md, k.db.View(func(tx *bolt.Tx) error {
		hist := tx.Bucket([]byte("history:" + name))
		if hist == nil {
			return fmt.Errorf("no password: %s", name)
		}
		refs := tx.Bucket([]byte("refs"))
		if refs == nil {
			return ErrNoRefs
		}

		r := refs.Get([]byte(name))
		rev = btoi(r)

		return json.Unmarshal(hist.Get(r), &md)
	})
}

func (k *keyman) ForEach(fn func(name string, rev uint64) error) error {
	return k.db.View(func(tx *bolt.Tx) error {
		refs := tx.Bucket([]byte("refs"))
		if refs == nil {
			return nil
		}

		return refs.ForEach(func(k, v []byte) error {
			return fn(string(k), btoi(v))
		})
	})
}

func (k *keyman) History(name string, limit uint64, fn func(rev uint64, md *Metadata) error) error {
	return k.db.View(func(tx *bolt.Tx) error {
		hist := tx.Bucket([]byte("history:" + name))
		if hist == nil {
			return fmt.Errorf("no password: %s", name)
		}

		c := hist.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var md Metadata
			if err := json.Unmarshal(v, &md); err != nil {
				return err
			}

			if err := fn(btoi(k), &md); err != nil {
				return err
			}

			if limit > 0 {
				limit--
				if limit == 0 {
					break
				}
			}
		}

		return nil
	})
}
