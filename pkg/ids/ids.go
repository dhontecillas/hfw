package ids

import (
	"fmt"

	"github.com/oklog/ulid/v2"
)

// ID is an alias for an ULID
type ID ulid.ULID

const uidFmt string = "%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x"

// ToUUID converts an ULID to a UUID format string.
func (i *ID) ToUUID() string {
	return fmt.Sprintf(uidFmt,
		i[0], i[1], i[2], i[3],
		i[4], i[5],
		i[6], i[7],
		i[8], i[9],
		i[10], i[11], i[12], i[13], i[14], i[15])
}

func (i *ID) IsZero() bool {
	var zero ID
	return *i == zero
}

// FromUUID reads the bytes from a UUID formatted ulid.
func (i *ID) FromUUID(uuid string) error {
	_, err := fmt.Sscanf(uuid, uidFmt,
		&i[0], &i[1], &i[2], &i[3],
		&i[4], &i[5],
		&i[6], &i[7],
		&i[8], &i[9],
		&i[10], &i[11], &i[12], &i[13], &i[14], &i[15])
	return err
}

func (i *ID) ToULID() ulid.ULID {
	return ulid.ULID(*i)
}

// ToShuffled just scrambles the string to make it less obvious
// that IDs are sequential, and that starts with the timestamp
func (i *ID) ToShuffled() string {
	var shuffled ID

	shuffled[8] = i[0]
	shuffled[10] = i[1]
	shuffled[12] = i[2]
	shuffled[14] = i[3]
	shuffled[0] = i[4]
	shuffled[2] = i[5]
	shuffled[4] = i[6]
	shuffled[6] = i[7]

	shuffled[1] = i[8]
	shuffled[3] = i[9]
	shuffled[5] = i[10]
	shuffled[7] = i[11]
	shuffled[9] = i[12]
	shuffled[11] = i[13]
	shuffled[13] = i[14]
	shuffled[15] = i[15]

	return ulid.ULID(shuffled).String()
}

// FromShuffled reverse the mangling applied with ToMangle
func (i *ID) FromShuffled(shuffledStr string) error {
	shuffled, err := ulid.ParseStrict(shuffledStr)
	if err != nil {
		return err
	}

	i[0] = shuffled[8]
	i[1] = shuffled[10]
	i[2] = shuffled[12]
	i[3] = shuffled[14]
	i[4] = shuffled[0]
	i[5] = shuffled[2]
	i[6] = shuffled[4]
	i[7] = shuffled[6]

	i[8] = shuffled[1]
	i[9] = shuffled[3]
	i[10] = shuffled[5]
	i[11] = shuffled[7]
	i[12] = shuffled[9]
	i[13] = shuffled[11]
	i[14] = shuffled[13]
	i[15] = shuffled[15]

	return nil
}

// Scan is the interface to be able to read the uuid from the database
// and put them into an ID type directly.
func (i *ID) Scan(src interface{}) error {
	var s string
	var ok bool
	if b, ok := src.([]byte); ok {
		if len(b) == 16 {
			var ii = ((*[16]byte)(i))[:]
			copy(ii, b)
			return nil
		}
		s = string(b)
	}
	if len(s) == 0 {
		s, ok = src.(string)
		if !ok {
			return fmt.Errorf("cannot understand type")
		}
	}
	if len(s) == 36 && i.FromUUID(s) == nil {
		return nil
	}
	return i.FromShuffled(s)
}
