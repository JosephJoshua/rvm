package user

import "fmt"

type Service struct {
	r Repository
}

func NewService(r Repository) *Service {
	return &Service{r}
}

func (s *Service) GetPoints(uid string) (int, error) {
	p, err := s.r.GetPoints(uid)
	if err != nil {
		return 0, fmt.Errorf("GetPoints(): failed to get points: %w", err)
	}

	return p, nil
}
