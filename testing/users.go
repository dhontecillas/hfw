package testing

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"
)

func RandomEmailAndPassword() (string, string) {
	randints := []uint32{
		uint32(time.Now().UTC().UnixMicro() & 0xffffffff),
		rand.Uint32(), rand.Uint32(), rand.Uint32(),
	}

	b := make([]byte, len(randints)*4)
	for idx, i := range randints {
		b[idx*4] = (byte)(i & 0xff)
		b[idx*4+1] = (byte)((i >> 8) & 0xff)
		b[idx*4+2] = (byte)((i >> 16) & 0xff)
		b[idx*4+3] = (byte)((i >> 24) & 0xff)
	}

	str := base64.StdEncoding.EncodeToString(b)
	user := fmt.Sprintf("u%s@example.com", str[:8])
	pass := str[8:12]
	return user, pass
}
