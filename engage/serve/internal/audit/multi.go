package audit

// MultiStore writes to multiple audit backends.
type MultiStore struct {
	stores []Appender
}

type Appender interface {
	Append(Event) error
}

func NewMultiStore(stores ...Appender) *MultiStore {
	out := make([]Appender, 0, len(stores))
	for _, s := range stores {
		if s != nil {
			out = append(out, s)
		}
	}
	return &MultiStore{stores: out}
}

func (m *MultiStore) Append(e Event) error {
	for _, s := range m.stores {
		_ = s.Append(e)
	}
	return nil
}
