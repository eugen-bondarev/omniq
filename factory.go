package omniq

type JobFactory[T any] interface {
	Instantiate(t string, id string, data map[string]any) Job[T]
}
