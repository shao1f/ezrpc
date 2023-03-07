package codec

import "io"

type Header struct {
	ServiceMethod string
	Seq           uint64
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type CodecType string

const (
	GobType  CodecType = "application/gob"
	JsonType CodecType = "application/json"
)

var NewCodecFuncMap map[CodecType]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[CodecType]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
