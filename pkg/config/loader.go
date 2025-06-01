package config

type ConfValidator interface {
	Validate() error
}

type ConfLoader interface {
	Section(name []string) (ConfLoader, error)
	Parse(target any) error
}
