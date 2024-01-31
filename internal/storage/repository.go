package storage

type Repository interface {
	Save(string) (string, error)
	Get(string) (string, error)
}
