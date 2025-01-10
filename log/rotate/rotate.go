package rotate

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

const (
	MaxFileSize1M   = 1024 * 1024
	MaxFileSize512M = 512 * MaxFileSize1M
	MaxFileSize1G   = 1024 * MaxFileSize1M
)

var RootPerm = os.FileMode(0755)

var FilePerm = os.FileMode(0666)

type Writer struct {
	root    string
	prefix  string
	current *os.File
	size    int
	max     int
	sync.Mutex
}

func New(rootpath, name string) (*Writer, error) {

	l := &Writer{root: rootpath, prefix: name, max: MaxFileSize512M}
	if err := l.setup(); err != nil {
		return nil, err
	}
	return l, nil
}

// SetMax sets the maximum size for a file in bytes.
func (r *Writer) SetMax(size int) {
	r.max = size
}

// Write writes p to the current file, then checks to see if
// rotation is necessary.
func (r *Writer) Write(p []byte) (n int, err error) {
	r.Lock()
	defer r.Unlock()
	n, err = r.current.Write(p)
	if err != nil {
		return n, err
	}
	r.size += n
	if r.size >= r.max {
		if err := r.rotate(); err != nil {
			return n, err
		}
	}
	return n, nil
}

// Close closes the current file.  Writer is unusable after this
// is called.
func (r *Writer) Close() error {
	r.Lock()
	defer r.Unlock()
	if err := r.current.Close(); err != nil {
		return err
	}
	r.current = nil
	return nil
}

func (r *Writer) setup() error {
	fi, err := os.Stat(r.root)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(r.root, RootPerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !fi.IsDir() {
		return errors.New("root must be a directory")
	}

	return r.openCurrent()
}

func (r *Writer) openCurrent() error {
	cp := path.Join(r.root, r.prefix+"_current.log")
	var err error
	r.current, err = os.OpenFile(cp, os.O_RDWR|os.O_CREATE|os.O_APPEND, FilePerm)
	if err != nil {
		return err
	}
	r.size = 0
	info, err := r.current.Stat()
	if err != nil {
		return err
	}
	if info.Size() != 0 {
		r.size = int(info.Size())
	}
	return nil
}

func (r *Writer) rotate() error {
	if err := r.current.Close(); err != nil {
		return err
	}
	filename := fmt.Sprintf("%s_%s.log", r.prefix, time.Now().Format("2006_0102_150405.000"))
	if err := os.Rename(path.Join(r.root, r.prefix+"_current.log"), path.Join(r.root, filename)); err != nil {
		return err
	}
	return r.openCurrent()
}

