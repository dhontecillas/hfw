package main

import (
	"fmt"

	"github.com/dhontecillas/hfw/pkg/ids"
)

func main() {
	gen := ids.NewIDGenerator()
	i := gen.MustNew()

	fmt.Printf("ULID:      %s\n", i.ToULID().String())
	fmt.Printf("UUID:      %s\n", i.ToUUID())
	fmt.Printf("Shuffled:  %s\n", i.ToShuffled())
}
