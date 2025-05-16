package trojan

import (
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/Dreamacro/protobytes"
)

const (
	// 最大数据包长度
	maxLength = 8192
	// Trojan 命令：TCP
	CommandTCP = 1
)

func hexSha224(data []byte) []byte {
	buf := make([]byte, 56)
	hash := sha256.New224()
	hash.Write(data)
	hex.Encode(buf, hash.Sum(nil))
	return buf
}

func writeHeader(w io.Writer, hexPassword []byte, command byte, socks5Addr []byte) error {
	buf := protobytes.BytesWriter{}
	buf.PutSlice(hexPassword)
	buf.PutSlice([]byte{'\r', '\n'})
	buf.PutUint8(command)
	buf.PutSlice(socks5Addr)
	buf.PutSlice([]byte{'\r', '\n'})

	_, err := w.Write(buf.Bytes())
	return err
}
