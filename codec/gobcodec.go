package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(conn),
	}
}

func (gc *GobCodec) Close() error {
	// log.Println("cc Close")
	return gc.conn.Close()
}

func (gc *GobCodec) ReadHeader(h *Header) error {
	return gc.dec.Decode(h)
}

func (gc *GobCodec) ReadBody(b interface{}) error {
	return gc.dec.Decode(b)
}

func (gc *GobCodec) Write(h *Header, b interface{}) (err error) {
	defer func() {
		_ = gc.buf.Flush()
		if err != nil {
			_ = gc.Close()
		}
	}()

	if err := gc.enc.Encode(h); err != nil {
		log.Println("ezrpc:gobcodec encode header err:", err.Error())
		return err
	}

	if err := gc.enc.Encode(b); err != nil {
		log.Println("ezrpc:gobcodec encode body err:", err.Error())
		return err
	}
	return nil
}
