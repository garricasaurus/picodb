package picodb

import (
	"os"
	"path"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// PicoDb is a simplistic directory based key-value storage.
// Keys are always of type string, and PicoDb will store the data
// associated with the given key under a file with the same name.
// PicoDb is always initialized with a root path, which will
// contain the data.
type PicoDb struct {
	opt   *PicoDbOptions
	cache *sync.Map
	id    uuid.UUID
}

// PicoDbOptions contains options which are passed on to the
// New function to create a PicoDb instace.
type PicoDbOptions struct {
	RootDir     string      // root directory
	Compression bool        // enable compression at rest
	Caching     bool        // enable in-memory cache
	FileMode    os.FileMode // file mode used to create files
	DirMode     os.FileMode // file mode used to create directories
}

// Defaults returns a PicoDbOptions with reasonable defaults.
func Defaults() *PicoDbOptions {
	return &PicoDbOptions{
		RootDir:     "./picodb",
		Compression: false,
		Caching:     false,
		FileMode:    0644,
		DirMode:     0744,
	}
}

// New returns a new PicoDb instance.
func New(options *PicoDbOptions) *PicoDb {
	return &PicoDb{
		opt:   options,
		cache: &sync.Map{},
		id:    uuid.New(),
	}
}

// Store a key.
func (p *PicoDb) Store(key string, val []byte) error {
	if err := p.check(key); err != nil {
		return err
	}
	name := p.path(key)
	dir := path.Dir(name)
	if err := os.MkdirAll(dir, p.opt.DirMode); err != nil {
		return err
	}
	return os.WriteFile(name, val, p.opt.FileMode)
}

// Load a key.
// If the key is missing, an error is returned.
func (p *PicoDb) Load(key string) ([]byte, error) {
	if err := p.check(key); err != nil {
		return nil, err
	}
	name := p.path(key)
	fi, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewKeyNotFound(key)
		}
		return nil, err
	}
	if fi.IsDir() {
		return nil, NewKeyNotFound(key)
	}
	return os.ReadFile(name)
}

// Delete a key.
// If the key is missing, it doesn't do anything.
func (p *PicoDb) Delete(key string) error {
	if err := p.check(key); err != nil {
		return err
	}
	name := p.path(key)
	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Remove(name)
}

func (p *PicoDb) check(key string) error {
	if strings.ContainsRune(key, os.PathSeparator) {
		return NewInvalidKey(key)
	}
	return nil
}

func (p *PicoDb) path(key string) string {
	return path.Join(p.opt.RootDir, key)
}
