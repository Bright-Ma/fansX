package server

import (
	"encoding/binary"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
)

type Codec struct {
}

func Serve(Addr string) error {
	// 设置编解码器，使用4字节的报文头标识消息边界
	codec := gnet.NewLengthFieldBasedFrameCodec(
		gnet.EncoderConfig{
			ByteOrder:                       binary.BigEndian,
			LengthFieldLength:               4,
			LengthIncludesLengthFieldLength: false,
		},
		gnet.DecoderConfig{
			ByteOrder:           binary.BigEndian,
			LengthFieldLength:   4,
			InitialBytesToStrip: 4,
		})

	handler := &Handler{pool: goroutine.Default()}
	return gnet.Serve(handler, Addr, gnet.WithMulticore(true), gnet.WithCodec(codec))
}
