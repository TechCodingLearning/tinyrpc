package compressor

// RawCompressor implements the Compressor interface
type RawCompressor struct {
}

// Zip raw
func (_ RawCompressor) Zip(data []byte) ([]byte, error) {
	return data, nil
}

// Unzip raw
func (_ RawCompressor) Unzip(data []byte) ([]byte, error) {
	return data, nil
}