package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	flagBinary     byte = 0x80
	flagEncrypted  byte = 0x40
	flagCompressed byte = 0x20
	flagBlueBox    byte = 0x10
	flagBigSize    byte = 0x08
)

// maxSmallSize is the largest body that fits a 2-byte length field
const maxSmallSize = 0xFFFF

// writes body as a single framed packet: a header byte, a 2- or 4-byte big-endian length, then body
func WriteFrame(w io.Writer, body []byte) error {
	size := len(body)
	var header []byte
	if size > maxSmallSize {
		header = make([]byte, 5)
		header[0] = flagBinary | flagBigSize
		binary.BigEndian.PutUint32(header[1:], uint32(size))
	} else {
		header = make([]byte, 3)
		header[0] = flagBinary
		binary.BigEndian.PutUint16(header[1:], uint16(size))
	}
	if _, err := w.Write(header); err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	return nil
}

func ReadFrame(r io.Reader) ([]byte, error) {
	var head [1]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return nil, err
	}
	flags := head[0]

	if flags&flagBinary == 0 {
		return nil, fmt.Errorf("protocol: not a binary packet (header 0x%02x)", flags)
	}
	if flags&flagEncrypted != 0 {
		return nil, fmt.Errorf("protocol: encrypted packets are not supported")
	}
	if flags&flagCompressed != 0 {
		return nil, fmt.Errorf("protocol: compressed packets are not supported")
	}
	if flags&flagBlueBox != 0 {
		return nil, fmt.Errorf("protocol: BlueBox (HTTP-tunneled) packets are not supported")
	}

	var size int
	if flags&flagBigSize != 0 {
		var b [4]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, err
		}
		size = int(binary.BigEndian.Uint32(b[:]))
	} else {
		var b [2]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, err
		}
		size = int(binary.BigEndian.Uint16(b[:]))
	}

	body := make([]byte, size)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}
