package transaction

type CodeGenerator interface {
	Generate() (string, error)
}
