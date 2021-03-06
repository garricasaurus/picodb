package picodb

import (
	"sync"

	"github.com/google/uuid"
)

// PicoDb is a simplistic directory based key-value storage.
// Keys are always of type string, and PicoDb will store the data
// associated with the given key under a file with the same name.
// PicoDb is always initialized with a root path, which will
// contain the data.
type PicoDb struct {
	id  uuid.UUID      // the unique id of this picodb instance
	opt *PicoDbOptions // picodb options
	kvs kvs            // the key-value store backend
}

// New returns a new PicoDb instance.
func New(options *PicoDbOptions) *PicoDb {
	return &PicoDb{
		id:  uuid.New(),
		kvs: newKvs(options),
		opt: options,
	}
}

func newKvs(options *PicoDbOptions) kvs {
	var dirfs = &dirfs{
		root:    options.RootDir,
		s:       newStorage(options),
		locking: options.Locking,
	}
	if !options.Caching {
		return dirfs
	}
	return &chain{
		list: []kvs{
			&cache{m: &sync.Map{}},
			dirfs,
		},
	}
}

func newStorage(opt *PicoDbOptions) storage {
	fs := &fs{
		fmode: opt.FileMode,
		dmode: opt.DirMode,
	}
	if opt.Compression {
		return &fsc{fs}
	} else {
		return fs
	}
}

// Store a key.
func (p *PicoDb) Store(key string, val []byte) error {
	return p.kvs.store(key, val)
}

// Store a key with a string value.
func (p *PicoDb) StoreString(key, val string) error {
	return p.Store(key, []byte(val))
}

// Load a key.
// If the key is missing, an error is returned.
func (p *PicoDb) Load(key string) ([]byte, error) {
	return p.kvs.load(key)
}

// Load a key with a string value.
// If the key is missing, and error is returned.
func (p *PicoDb) LoadString(key string) (string, error) {
	b, err := p.Load(key)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Delete a key.
// If the key is missing, an error is returned.
func (p *PicoDb) Delete(key string) error {
	return p.kvs.delete(key)
}
