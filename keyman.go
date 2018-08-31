package keyman

import (
	"os/exec"
	"regexp"
	"time"

	"github.com/boltdb/bolt"
)

var p = regexp.MustCompile("%[^%]+%")

type Keyman interface {
	Context() Context
	Get(name string) ([]byte, error)
	Put(name string, value []byte) ([]byte, error)
	Command(name string, args ...string) (*exec.Cmd, error)
	ForEach(fn func(name string, rev []byte, ctx Context) error) error
	History(name string) (HistoryBucket, error)
}

type keyman struct {
	Keyman
	ctx Context
}

func New(tx *bolt.Tx) Keyman {
	ctx := &context{tx: tx}
	k := &keyman{ctx: ctx}
	return k
}

func (k *keyman) Context() Context {
	return k.ctx
}

func (k *keyman) Get(name string) ([]byte, error) {
	r, err := k.ctx.OpenRevisionBucket()
	if err != nil {
		return nil, err
	}

	s, err := k.ctx.OpenSecretBucket(name)
	if err != nil {
		return nil, err
	}

	return s.Get(r.Get(name)), nil
}

func (k *keyman) Put(name string, value []byte) ([]byte, error) {
	h, err := k.ctx.OpenHistoryBucket(name)
	if err != nil {
		return nil, err
	}

	rev, err := h.Push(&Metadata{
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}

	s, err := k.ctx.OpenSecretBucket(name)
	if err != nil {
		return nil, err
	}
	if err := s.Put(rev, value); err != nil {
		return nil, err
	}

	r, err := k.ctx.OpenRevisionBucket()
	if err != nil {
		return nil, err
	}

	return rev, r.Put(name, rev)
}

func (k *keyman) Command(name string, args ...string) (*exec.Cmd, error) {
	dict := map[string][]byte{}
	for _, arg := range args {
		for _, m := range p.FindAllString(arg, -1) {
			dict[m] = nil
		}
	}

	for name := range dict {
		last := len(name) - 1
		v, err := k.Get(name[1:last])
		if err != nil {
			return nil, err
		}

		dict[name] = v
	}

	for i, arg := range args {
		args[i] = p.ReplaceAllStringFunc(arg, func(m string) string {
			return string(dict[m])
		})
	}

	return exec.Command(name, args...), nil
}

func (k *keyman) ForEach(fn func(name string, rev []byte, ctx Context) error) error {
	r, err := k.ctx.OpenRevisionBucket()
	if err != nil {
		return err
	}

	return r.ForEach(func(name string, rev []byte) error {
		return fn(name, rev, k.ctx)
	})
}

func (k *keyman) History(name string) (HistoryBucket, error) {
	return k.ctx.OpenHistoryBucket(name)
}
