package omniq

type JobFactory[T any] interface {
	Instantiate(t string, id JobID, data string) Job[T]
}
