package omniq

type Job[T any] interface {
	Run(container T)
	Type() string
	GetIDContainer() *WithID
}

type JobID string

type WithID struct {
	ID JobID
}

func (w *WithID) GetID() JobID {
	return w.ID
}

func (w *WithID) SetID(id JobID) {
	w.ID = id
}
