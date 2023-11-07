package transaction

type Repository interface {
	StartTransaction(code string) error
}
