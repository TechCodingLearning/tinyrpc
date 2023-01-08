package tinyrpc

import (
	"tinyrpc/compressor"
	"tinyrpc/serializer"
)

type Option func(*options)

type options struct {
	compressType compressor.CompressType
	serializer   serializer.Serializer
}

func WithCompress(c compressor.CompressType) Option {
	return func(o *options) {
		o.compressType = c
	}
}

func WithSerializer(serializer serializer.Serializer) Option {
	return func(o *options) {
		o.serializer = serializer
	}
}
