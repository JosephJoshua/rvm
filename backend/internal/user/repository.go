package user

type Repository interface {
	GetPoints(uid string) (int, error)
}
