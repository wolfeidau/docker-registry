package uuid

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

func NewUUID() string {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		panic("/dev/urandom is broken, process is untrustworthy, uuid gen failed")
	}
	dashedId := make([]byte, 36)
	hex.Encode(dashedId[0:8], id[0:4])
	dashedId[8] = '-'
	hex.Encode(dashedId[9:13], id[4:6])
	dashedId[13] = '-'
	hex.Encode(dashedId[14:18], id[6:8])
	dashedId[18] = '-'
	hex.Encode(dashedId[19:23], id[8:10])
	dashedId[23] = '-'
	hex.Encode(dashedId[24:36], id[10:16])
	return string(dashedId)
}
