package server

import (
	"encoding/binary"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
)

type Codec struct {
}

func Serve(Addr string) error {
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
