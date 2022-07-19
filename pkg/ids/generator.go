package ids

import (
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	seed    int64
	seedMtx = &sync.Mutex{}
)

// IDGenerator returns an ulid generator
type IDGenerator struct {
	entropy io.Reader
}

// NewIDGenerator createts new IDGenerator instance with a unique
// seed for the current running applications. (Different instances
// of the application, that are initialized at the same time could
// create a collision)
func NewIDGenerator() *IDGenerator {
	seedMtx.Lock()
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	curSeed := seed
	seed++
	seedMtx.Unlock()
	return &IDGenerator{
		entropy: rand.New(rand.NewSource(curSeed)),
	}
}

// New returns a new ulid
func (g *IDGenerator) New() (ID, error) {
	u, e := ulid.New(ulid.Now(), g.entropy)
	return ID(u), e
}

// MustNew returns a new ulid, but if an error happens it
// will panic
func (g *IDGenerator) MustNew() ID {
	u, e := ulid.New(ulid.Now(), g.entropy)
	if e != nil {
		panic(e.Error())
	}
	return ID(u)
}
