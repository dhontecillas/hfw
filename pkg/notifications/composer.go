package notifications

// ContentSet contains the set of rendered pieces to
// be used for a given Carrier (to create a Dispatch
// for each target user)
type ContentSet struct {
	HTMLs map[string]string
	Texts map[string]string
}

// Composer defines the interface to create a ContentSet
type Composer interface {
	Render(notification string, data map[string]interface{}, carrier string) (*ContentSet, error)
}
