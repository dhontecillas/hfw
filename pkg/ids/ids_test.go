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
