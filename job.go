package omniq

type Job[T any] interface {
	Run(container T)
	Type() string
	GetIDContainer() *WithID
}

type WithID struct {
	ID string
}

func (w *WithID) GetID() string {
	return w.ID
}

func (w *WithID) SetID(id string) {
	w.ID = id
}
