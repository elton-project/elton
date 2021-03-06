package localStorage

import (
	"encoding/hex"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/idgen"
	"math/rand"
	"sync"
)

// KeyGenerator generates the object key.
type KeyGenerator interface {
	Generate() Key
}

// TimeKeyGen generates unique ID by the SonyFlake algorithm.
type UniqueKeyGen struct {
	init sync.Once
	gen  idgen.Generator
}

func (u *UniqueKeyGen) Generate() Key {
	u.init.Do(func() {
		u.gen = idgen.NewSonyFlake()
	})
	id, err := u.gen.NextStringID()
	if err != nil {
		panic(err)
	}
	return Key{
		ID: id,
	}
}

// RandomKeyGen generates the object key by random.
type RandomKeyGen struct {
	Seed int64

	init   sync.Once
	random *rand.Rand
}

func (r *RandomKeyGen) Generate() Key {
	r.init.Do(func() {
		src := rand.NewSource(r.Seed)
		r.random = rand.New(src)
	})

	b := make([]byte, 24)
	for i := 0; i < 24; i += 8 {
		u := r.random.Uint64()
		b[i+0] = byte(u & 7)
		u >>= 8
		b[i+1] = byte(u & 7)
		u >>= 8
		b[i+2] = byte(u & 7)
		u >>= 8
		b[i+3] = byte(u & 7)
		u >>= 8
		b[i+4] = byte(u & 7)
		u >>= 8
		b[i+5] = byte(u & 7)
		u >>= 8
		b[i+6] = byte(u & 7)
		u >>= 8
		b[i+7] = byte(u & 7)
	}

	return Key{
		ID: hex.EncodeToString(b),
	}
}
