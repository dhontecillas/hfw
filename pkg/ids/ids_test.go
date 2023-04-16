package ids

import (
	"testing"

	"github.com/oklog/ulid/v2"
)

func Test_ToUUID(t *testing.T) {
	id := ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	want := "00010203-0405-0607-0809-0a0b0c0d0e0f"

	got := id.ToUUID()
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func Test_FromUUID(t *testing.T) {
	uuid := "00010203-0405-0607-0809-0a0b0c0d0e0f"
	want := ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	id := ID{}
	err := id.FromUUID(uuid)
	if err != nil {
		t.Errorf("cannot parse uuid %s", err)
	}

	if id != want {
		t.Errorf("want: %s, got: %s", want, id)
	}
}

func Test_Shuffling(t *testing.T) {
	want := ID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	shuffled := want.ToShuffled()

	if shuffled == ulid.ULID(want).String() {
		t.Errorf("shuffled should be different than the original")
	}

	var got ID
	err := got.FromShuffled(shuffled)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
		return
	}

	if want != got {
		t.Errorf("want: %s - got %s", ulid.ULID(want).String(), ulid.ULID(want).String())
	}
}

func Test_Scan(t *testing.T) {
	validBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	shortBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	longBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}

	var bytesID ID
	err := bytesID.Scan(shortBytes)
	if err == nil {
		t.Errorf("should fail on short bytes input")
		return
	}
	err = bytesID.Scan(longBytes)
	if err == nil {
		t.Error("should fail on long bytes inputs")
		return
	}
	err = bytesID.Scan(validBytes)
	if err != nil {
		t.Error("should succeed on good bytes input")
		return
	}

	validUUID := "3ab78503-2fae-46b2-8170-4350831467c7"
	invalidUUID := "3ab7850-22fae-46b2-8170-4350831467c7"
	shortUUID := "3ab7850-22fae-46b2-8170-4350831467c"
	var stringID ID
	err = stringID.Scan(invalidUUID)
	if err == nil {
		t.Errorf("should fail on invalid uuid")
		return
	}
	err = stringID.Scan(shortUUID)
	if err == nil {
		t.Error("should fail on long short uuid")
		return
	}
	err = stringID.Scan(validUUID)
	if err != nil {
		t.Error("should succeed on valid uuid")
		return
	}

	shuffled := stringID.ToShuffled()
	var shuffledID ID
	err = shuffledID.Scan(shuffled)
	if err != nil {
		t.Error("should succeed on scanning a shuffled string")
		return
	}
}

func Test_Generator(t *testing.T) {
	gen := NewIDGenerator()

	var zeroID ID
	id, err := gen.New()
	if err != nil {
		t.Errorf("non expected error")
		return
	}
	if id == zeroID {
		t.Errorf("id should not be empty")
		return
	}

	id = gen.MustNew()
	if id == zeroID {
		t.Errorf("must id should not be empty")
		return
	}
}
