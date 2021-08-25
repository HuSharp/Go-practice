package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn	io.ReadWriteCloser
	buf		*bufio.Writer		// 带缓冲的 writer，防止阻塞
	dec		*gob.Decoder
	enc		*gob.Encoder
}

// 检查结构体是否实现了这个接口
var _ Codec = (*GobCodec)(nil)

func NewGobCode(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

func (g GobCodec) Close() error {
	return g.conn.Close()
}

func (g GobCodec) ReadHeader(header *Header) error {
	return g.dec.Decode(header)
}

func (g GobCodec) ReadBody(body interface{}) error {
	return g.dec.Decode(body)
}

func (g GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		// 说有数据都写入后，调用者有义务调用Flush方法以保证所有的数据都交给了下层的io.Writer
		err = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err = g.enc.Encode(header); err != nil {
		log.Println("[Write] gob error encoding header! err: ", err)
		return
	}
	if err = g.enc.Encode(body); err != nil {
		log.Println("[Write] gob error encoding body! err: ", err)
		return
	}
	return nil
}

