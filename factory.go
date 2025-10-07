package omniq

type JobFactory[T any] interface {
	Instantiate(t string, id JobID, data map[string]any) Job[T]
}
