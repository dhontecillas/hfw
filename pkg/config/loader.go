package config

type ConfLoader interface {
	Section(name []string) (ConfLoader, error)
	Parse(target any) error
}
