package repository

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type EventLog struct {
	mu   sync.Mutex
	file *os.File
	path string
}

func OpenEventLog(dir string) (*EventLog, error) {
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, err
	}
	p := filepath.Join(dir, "events.jsonl")
	f, e := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	return &EventLog{file: f, path: p}, nil
}
func (l *EventLog) Append(e cdc_domain.ChangeEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return errors.New("event log closed")
	}
	b, er := json.Marshal(e)
	if er != nil {
		return er
	}
	b = append(b, '\n')
	if _, er = l.file.Write(b); er != nil {
		return er
	}
	return l.file.Sync()
}
func (l *EventLog) Replay(fn func(cdc_domain.ChangeEvent) error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, e := l.file.Seek(0, 0); e != nil {
		return e
	}
	scan := bufio.NewScanner(l.file)
	for scan.Scan() {
		var e cdc_domain.ChangeEvent
		if er := json.Unmarshal(scan.Bytes(), &e); er != nil {
			return er
		}
		if er := fn(e); er != nil {
			return er
		}
	}
	return scan.Err()
}
func (l *EventLog) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	e := l.file.Close()
	l.file = nil
	return e
}
func copyTo(w io.Writer, r io.Reader) error { _, e := io.Copy(w, r); return e }
