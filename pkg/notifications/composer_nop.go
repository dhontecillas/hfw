package notifications

type NopComposer struct {
}

func NewNopComposer() *NopComposer {
	return &NopComposer{}
}

func (n *NopComposer) Render(notification string, data map[string]interface{},
	carrier string) (*ContentSet, error) {

	return &ContentSet{
		HTMLs: map[string]string{},
		Texts: map[string]string{},
	}, nil
}
