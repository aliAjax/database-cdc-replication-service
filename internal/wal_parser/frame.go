package wal_parser

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"hash/crc32"
	"io"
	"time"
)

const MaxFrame = 16 << 20

type FrameType byte

const (
	FrameBegin    FrameType = 1
	FrameRow      FrameType = 2
	FrameCommit   FrameType = 3
	FrameDDL      FrameType = 4
	FrameRollback FrameType = 5
)

type Frame struct {
	Type          FrameType
	TxID          string
	Position      cdc_domain.Position
	Table         string
	Before        map[string]any
	After         map[string]any
	SchemaVersion int
}
type Reader struct {
	r        *bufio.Reader
	protocol string
	last     cdc_domain.Position
}

func NewReader(rd io.Reader, protocol string) *Reader {
	return &Reader{r: bufio.NewReader(rd), protocol: protocol}
}
func (rd *Reader) ReadFrame() (Frame, error) {
	var header [9]byte
	if _, err := io.ReadFull(rd.r, header[:]); err != nil {
		return Frame{}, err
	}
	n := binary.BigEndian.Uint32(header[:4])
	if n < 5 || n > MaxFrame {
		return Frame{}, fmt.Errorf("invalid frame length %d", n)
	}
	typ := FrameType(header[4])
	if typ < FrameBegin || typ > FrameRollback {
		return Frame{}, fmt.Errorf("unknown frame type %d", typ)
	}
	payload := make([]byte, n-5)
	if _, err := io.ReadFull(rd.r, payload); err != nil {
		return Frame{}, fmt.Errorf("short frame: %w", err)
	}
	want := binary.BigEndian.Uint32(header[5:])
	if crc32.ChecksumIEEE(payload) != want {
		return Frame{}, errors.New("frame checksum mismatch")
	}
	f, err := decode(typ, payload, rd.protocol)
	if err != nil {
		return Frame{}, err
	}
	rd.last = f.Position
	if f.Position.Compare(rd.last) < 0 {
		return Frame{}, errors.New("position moved backwards")
	}
	return f, nil
}
func Encode(f Frame) []byte {
	payload := encodePayload(f)
	n := uint32(5 + len(payload))
	out := bytes.NewBuffer(nil)
	binary.Write(out, binary.BigEndian, n)
	out.WriteByte(byte(f.Type))
	binary.Write(out, binary.BigEndian, crc32.ChecksumIEEE(payload))
	out.Write(payload)
	return out.Bytes()
}
func encodePayload(f Frame) []byte {
	b := bytes.NewBuffer(nil)
	put := func(s string) { binary.Write(b, binary.BigEndian, uint16(len(s))); b.WriteString(s) }
	put(f.TxID)
	put(f.Position.File)
	binary.Write(b, binary.BigEndian, f.Position.Offset)
	binary.Write(b, binary.BigEndian, f.Position.Epoch)
	put(f.Table)
	binary.Write(b, binary.BigEndian, uint16(f.SchemaVersion))
	if f.Before == nil {
		f.Before = map[string]any{}
	}
	if f.After == nil {
		f.After = map[string]any{}
	}
	putMap(b, f.Before)
	putMap(b, f.After)
	return b.Bytes()
}
func putMap(b *bytes.Buffer, m map[string]any) {
	binary.Write(b, binary.BigEndian, uint16(len(m)))
	for k, v := range m {
		binary.Write(b, binary.BigEndian, uint16(len(k)))
		b.WriteString(k)
		s := fmt.Sprint(v)
		binary.Write(b, binary.BigEndian, uint16(len(s)))
		b.WriteString(s)
	}
}
func readString(r *bytes.Reader) (string, error) {
	var n uint16
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return "", err
	}
	if int(n) > MaxFrame || int(n) > r.Len() {
		return "", errors.New("invalid string length")
	}
	b := make([]byte, n)
	r.Read(b)
	return string(b), nil
}
func readMap(r *bytes.Reader) (map[string]any, error) {
	var n uint16
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return nil, err
	}
	m := map[string]any{}
	for i := 0; i < int(n); i++ {
		k, e := readString(r)
		if e != nil {
			return nil, e
		}
		v, e := readString(r)
		if e != nil {
			return nil, e
		}
		m[k] = v
	}
	return m, nil
}
func decode(t FrameType, p []byte, protocol string) (Frame, error) {
	r := bytes.NewReader(p)
	tx, e := readString(r)
	if e != nil {
		return Frame{}, e
	}
	file, e := readString(r)
	if e != nil {
		return Frame{}, e
	}
	var off, epoch uint64
	binary.Read(r, binary.BigEndian, &off)
	binary.Read(r, binary.BigEndian, &epoch)
	table, e := readString(r)
	if e != nil {
		return Frame{}, e
	}
	var schema uint16
	binary.Read(r, binary.BigEndian, &schema)
	before, e := readMap(r)
	if e != nil {
		return Frame{}, e
	}
	after, e := readMap(r)
	if e != nil {
		return Frame{}, e
	}
	return Frame{Type: t, TxID: tx, Position: cdc_domain.Position{Protocol: protocol, File: file, Offset: off, Epoch: epoch}, Table: table, Before: before, After: after, SchemaVersion: int(schema)}, nil
}

type Transaction struct {
	ID        string
	Events    []cdc_domain.ChangeEvent
	BeganAt   time.Time
	Committed bool
}

func Assemble(frames []Frame) ([]cdc_domain.ChangeEvent, error) {
	var tx *Transaction
	var out []cdc_domain.ChangeEvent
	for _, f := range frames {
		switch f.Type {
		case FrameBegin:
			if tx != nil {
				return nil, errors.New("nested transaction")
			}
			tx = &Transaction{ID: f.TxID, BeganAt: time.Now()}
		case FrameRow:
			if tx == nil {
				return nil, errors.New("row outside transaction")
			}
			op := cdc_domain.OpUpdate
			if f.Before == nil || len(f.Before) == 0 {
				op = cdc_domain.OpInsert
			}
			tx.Events = append(tx.Events, cdc_domain.NewEvent(tx.ID, f.Position, f.Table, op, f.Before, f.After))
		case FrameDDL:
			if tx == nil {
				return nil, errors.New("ddl outside transaction")
			}
			tx.Events = append(tx.Events, cdc_domain.NewEvent(tx.ID, f.Position, f.Table, cdc_domain.OpDDL, nil, f.After))
		case FrameCommit:
			if tx == nil {
				return nil, errors.New("commit without begin")
			}
			out = append(out, tx.Events...)
			tx.Committed = true
			tx = nil
		case FrameRollback:
			tx = nil
		}
	}
	if tx != nil {
		return nil, errors.New("unterminated transaction")
	}
	return out, nil
}
