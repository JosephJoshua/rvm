package transaction

import (
	"fmt"
)

type Service struct {
	r  Repository
	cg CodeGenerator
}

func NewService(r Repository, cg CodeGenerator) *Service {
	return &Service{r: r, cg: cg}
}

func (s *Service) StartTransaction() (string, error) {
	code, err := s.cg.Generate()
	if err != nil {
		return "", fmt.Errorf("StartTransaction(): failed to generate code: %w", err)
	}

	if err = s.r.StartTransaction(code); err != nil {
		return "", fmt.Errorf("StartTransaction(): failed to create transaction: %w", err)
	}

	return code, nil
}
