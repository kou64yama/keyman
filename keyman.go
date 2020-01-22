package keyman

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"keyman/pb"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/golang/protobuf/ptypes"
)

const (
	ChunkSize = 512 * 1024
)

var (
	prefixBlob = "blob"
	prefixHead = "head"
	prefixHash = "hash"
	prefixTime = "time"
)

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

type Logger struct {
	ErrorLog   *log.Logger
	WarningLog *log.Logger
	InfoLog    *log.Logger
	DebugLog   *log.Logger
}

func (l *Logger) Errorf(fmt string, args ...interface{}) {
	l.ErrorLog.Printf(fmt, args...)
}

func (l *Logger) Warningf(fmt string, args ...interface{}) {
	l.WarningLog.Printf(fmt, args...)
}

func (l *Logger) Infof(fmt string, args ...interface{}) {
	l.InfoLog.Printf(fmt, args...)
}

func (l *Logger) Debugf(fmt string, args ...interface{}) {
	l.DebugLog.Printf(fmt, args...)
}

type metadata struct {
	Name string     `json:"name"`
	Hash string     `json:"hash"`
	Time *time.Time `json:"time"`
	Size uint64     `json:"size"`
}

type keymanServer struct {
	logger *Logger
	db     *badger.DB
}

func NewServer(logger *Logger, db *badger.DB) pb.KeymanServer {
	return &keymanServer{logger: logger, db: db}
}

func (s *keymanServer) List(req *pb.ListRequest, res pb.Keyman_ListServer) error {
	s.logger.Debugf("Keyman.List {All=%v}", req.GetAll())
	return s.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = []byte(prefixHead + ":")
		itr := txn.NewIterator(opt)
		defer itr.Close()
		for itr.Rewind(); itr.Valid(); itr.Next() {
			if err := itr.Item().Value(func(val []byte) error {
				var meta metadata
				if err := json.Unmarshal(val, &meta); err != nil {
					return err
				}
				if meta.Size == 0 && !req.GetAll() {
					return nil
				}
				timestamp, err := ptypes.TimestampProto(*meta.Time)
				if err != nil {
					return err
				}
				s.logger.Debugf("Read: %s %s", meta.Name, meta.Hash[:7])
				return res.Send(&pb.Metadata{
					Name: meta.Name,
					Hash: meta.Hash,
					Time: timestamp,
					Size: meta.Size,
				})
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *keymanServer) Get(req *pb.GetRequest, stream pb.Keyman_GetServer) error {
	s.logger.Debugf("Keyman.Get {Name=%s, Hash=%s}", req.GetName(), req.GetHash())
	return s.db.View(func(txn *badger.Txn) error {
		name := req.GetName()
		hash := req.GetHash()
		nameBase64 := base64.RawStdEncoding.EncodeToString([]byte(name))

		var metaItem *badger.Item
		if len(hash) == 0 {
			var err error
			metaItem, err = txn.Get([]byte(prefixHead + ":" + nameBase64))
			if err != nil {
				return err
			}
		} else {
			opt := badger.DefaultIteratorOptions
			opt.Prefix = []byte(prefixHash + ":" + nameBase64 + ":" + hash)
			itr := txn.NewIterator(opt)
			defer itr.Close()
			itr.Rewind()
			if !itr.Valid() {
				return errors.New("")
			}
			metaItem = itr.Item()
			if itr.Next(); itr.Valid() {
				return errors.New("")
			}
		}

		var meta metadata
		if err := metaItem.Value(func(val []byte) error {
			return json.Unmarshal(val, &meta)
		}); err != nil {
			return err
		}

		opt := badger.DefaultIteratorOptions
		opt.Prefix = []byte(prefixBlob + ":" + meta.Hash + ":")
		itr := txn.NewIterator(opt)
		defer itr.Close()
		for itr.Rewind(); itr.Valid(); itr.Next() {
			s.logger.Debugf("Chunk: %s", itr.Item().Key())
			if err := itr.Item().Value(func(val []byte) error {
				return stream.Send(&pb.GetResponse{Chunk: val})
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *keymanServer) Set(stream pb.Keyman_SetServer) error {
	s.logger.Debugf("Keyman.Set")

	temp, err := ioutil.TempFile("", "keyman-*")
	if err != nil {
		return err
	}
	s.logger.Debugf("Temp: %s", temp.Name())
	defer os.Remove(temp.Name())

	now := time.Now().UTC()
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	name := req.GetName()
	nameBase64 := base64.RawStdEncoding.EncodeToString([]byte(name))

	timeBytes := itob(uint64(now.UnixNano()))
	timeHex := hex.EncodeToString(timeBytes)
	hash := sha256.New()
	hash.Write(timeBytes)
	hash.Write([]byte{'\a'})
	hash.Write([]byte(name))
	hash.Write([]byte{'\a'})

	w := io.MultiWriter(temp, hash)
	var size uint64 = 0
	for true {
		req, err = stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		chunk := req.GetChunk()
		size += uint64(len(chunk))
		if _, err := w.Write(chunk); err != nil {
			return err
		}
	}
	if _, err := temp.Seek(0, 0); err != nil {
		return err
	}

	meta := &metadata{
		Name: name,
		Hash: hex.EncodeToString(hash.Sum(nil)),
		Time: &now,
		Size: size,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		var entries []*badger.Entry
		entries = append(entries, badger.NewEntry([]byte(prefixHead+":"+nameBase64), metaBytes))
		entries = append(entries, badger.NewEntry([]byte(prefixHash+":"+nameBase64+":"+meta.Hash), metaBytes))
		entries = append(entries, badger.NewEntry([]byte(prefixTime+":"+nameBase64+":"+timeHex), metaBytes))
		for _, e := range entries {
			if err := txn.SetEntry(e); err != nil {
				return err
			}
		}

		buf := make([]byte, ChunkSize)
		var n uint32 = 0
		for true {
			size, err := temp.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			n++
			b := make([]byte, 4)
			binary.BigEndian.PutUint32(b, n)
			chunk := prefixBlob + ":" + meta.Hash + ":" + hex.EncodeToString(b)
			s.logger.Debugf("Chunk: %s", chunk)
			if err := txn.Set([]byte(chunk), buf[:size]); err != nil {
				return err
			}
		}

		timestamp, err := ptypes.TimestampProto(*meta.Time)
		if err != nil {
			return err
		}
		return stream.SendAndClose(&pb.Metadata{
			Name: meta.Name,
			Hash: meta.Hash,
			Time: timestamp,
			Size: meta.Size,
		})
	})
}

func (s *keymanServer) Log(req *pb.LogRequest, res pb.Keyman_LogServer) error {
	s.logger.Debugf("Keyman.Log {Name=%s, Limit=%d}", req.GetName(), req.GetLimit())
	return s.db.View(func(txn *badger.Txn) error {
		name := req.GetName()
		nameBase64 := base64.RawStdEncoding.EncodeToString([]byte(name))
		opt := badger.DefaultIteratorOptions
		opt.Prefix = []byte(prefixTime + ":" + nameBase64 + ":")
		opt.Reverse = true
		itr := txn.NewIterator(opt)
		defer itr.Close()

		var count uint64 = 0
		for itr.Seek(append(opt.Prefix, '\xff')); itr.Valid(); itr.Next() {
			if req.GetLimit() > 0 && count >= req.GetLimit() {
				break
			}
			count++
			if err := itr.Item().Value(func(val []byte) error {
				var meta metadata
				if err := json.Unmarshal(val, &meta); err != nil {
					return err
				}
				timestamp, err := ptypes.TimestampProto(*meta.Time)
				if err != nil {
					return err
				}
				s.logger.Debugf("Read: %s %s", meta.Name, meta.Hash[:7])
				return res.Send(&pb.Metadata{
					Name: meta.Name,
					Hash: meta.Hash,
					Time: timestamp,
					Size: meta.Size,
				})
			}); err != nil {
				return err
			}
		}
		return nil
	})
}
