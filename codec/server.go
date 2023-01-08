package codec

import (
	"bufio"
	"hash/crc32"
	"io"
	"net/rpc"
	"sync"
	"tinyrpc/compressor"
	"tinyrpc/header"
	"tinyrpc/serializer"
)

type reqCtx struct {
	requestID   uint64
	compareType compressor.CompressType
}

type serverCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	request    header.RequestHeader
	serializer serializer.Serializer
	mutex      sync.Mutex // protect seq pending
	seq        uint64
	pending    map[uint64]*reqCtx
}

// NewServerCodec Create a new server codec
func NewServerCodec(conn io.ReadWriteCloser, serializer serializer.Serializer) rpc.ServerCodec {
	return &serverCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		serializer: serializer,
		pending:    make(map[uint64]*reqCtx),
	}
}

// ReadRequestHeader read the rpc request header from the io stream
func (s *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	s.request.ResetHeader()
	data, err := recvFrame(s.r)
	if err != nil {
		return err
	}

	err = s.request.Unmarshal(data)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	s.seq++ // sequence self-add
	s.pending[s.seq] = &reqCtx{s.request.ID, s.request.CompressType}
	r.Seq = s.seq
	r.ServiceMethod = s.request.Method
	s.mutex.Unlock()
	return nil
}

// ReadRequestBody read the rpc request from the io stream
func (s *serverCodec) ReadRequestBody(param interface{}) error {
	if param == nil {
		if s.request.RequestLen != 0 { // discard the rest
			if err := read(s.r, make([]byte, s.request.RequestLen)); err != nil {
				return err
			}
		}
		return nil
	}

	reqBody := make([]byte, s.request.RequestLen)

	err := read(s.r, reqBody)
	if err != nil {
		return err
	}

	if s.request.Checksum != 0 {
		if crc32.ChecksumIEEE(reqBody) != s.request.Checksum {
			return UnexpectedChecksumError
		}
	}

	if _, ok := compressor.Compressors[s.request.GetCompressType()]; !ok {
		return NotFoundCompressorError
	}

	req, err := compressor.Compressors[s.request.GetCompressType()].Unzip(reqBody)
	if err != nil {
		return err
	}
	return s.serializer.Unmarshal(req, param)
}

// WriteResponse write the rpc response header and body to the io stream
func (s *serverCodec) WriteResponse(r *rpc.Response, param interface{}) error {
	s.mutex.Lock()
	reqCtx, ok := s.pending[r.Seq]
	if !ok {
		s.mutex.Unlock()
		return InvalidSequenceError
	}
	s.mutex.Unlock()

	if r.Error != "" {
		param = nil
	}

	if _, ok := compressor.Compressors[reqCtx.compareType]; !ok {
		return NotFoundCompressorError
	}

	// marshal response body
	respBody, err := s.serializer.Marshal(param)
	if err != nil {
		return err
	}
	// zip response body
	compressedRespBody, err := compressor.Compressors[reqCtx.compareType].Zip(respBody)
	if err != nil {
		return err
	}

	// marshal response header
	h := header.ResponsePool.Get().(*header.ResponseHeader)
	defer func() {
		h.ResetHeader()
		header.ResponsePool.Put(h)
	}()

	h.ID = reqCtx.requestID
	h.ResponseLen = uint32(len(compressedRespBody))
	h.Checksum = crc32.ChecksumIEEE(compressedRespBody)
	h.CompressType = reqCtx.compareType
	h.Error = r.Error

	if err = sendFrame(s.w, h.Marshal()); err != nil {
		return err
	}

	if err = write(s.w, compressedRespBody); err != nil {
		return err
	}

	if err = s.w.(*bufio.Writer).Flush(); err != nil {
		return err
	}
	return nil
}

// Close .
func (s *serverCodec) Close() error {
	return s.c.Close()
}
