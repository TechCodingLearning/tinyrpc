package codec

import (
	"encoding/binary"
	"io"
	"net"
)

func sendFrame(w io.Writer, data []byte) (err error) {
	var size [binary.MaxVarintLen64]byte

	if len(data) == 0 {
		n := binary.PutUvarint(size[:], uint64(0))
		if err = write(w, size[:n]); err != nil {
			return err
		}
	}

	n := binary.PutUvarint(size[:], uint64(len(data)))
	if err = write(w, size[:n]); err != nil {
		return err
	}

	if err = write(w, data); err != nil {
		return err
	}
	return nil
}

func recvFrame(r io.Reader) (data []byte, err error) {
	size, err := binary.ReadUvarint(r.(io.ByteReader))
	if err != nil {
		return nil, err
	}

	if size != 0 {
		data = make([]byte, size)
		if err = read(r, data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func write(w io.Writer, data []byte) error {
	for idx := 0; idx < len(data); {
		n, err := w.Write(data[idx:])
		if err != nil {
			if _, ok := err.(net.Error); !ok {
				return err
			}
		}
		idx += n
	}
	return nil
}

func read(r io.Reader, data []byte) error {
	for idx := 0; idx < len(data); {
		n, err := r.Read(data[idx:])
		if err != nil {
			if _, ok := err.(net.Error); !ok {
				return err
			}
		}
		idx += n
	}
	return nil
}
