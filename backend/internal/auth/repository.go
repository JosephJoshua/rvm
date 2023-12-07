package auth

type Repository interface {
	DoesUserExist(uid string) (bool, error)
	CreateUser(uid string, fullName string, email string) error
}
