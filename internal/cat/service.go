package cat

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListCats() ([]Cat, error) {
	cats, err := s.repo.GetAllCats()
	if err != nil {
		return nil, err
	}

	return cats, nil
}

func (s *Service) CreateCat(name, breed string, yearsOfExperience int, salary float64) (int, error) {
	cat := &Cat{
		Name:              name,
		Breed:             breed,
		YearsOfExperience: yearsOfExperience,
		Salary:            salary,
	}

	return s.repo.CreateCat(cat)
}

func (s *Service) GetCat(id int) (*Cat, error) {
	return s.repo.GetCatByID(id)
}

func (s *Service) UpdateCatSalary(id int, salary float64) error {
	cat, err := s.repo.GetCatByID(id)
	if err != nil {
		return err
	}
	if cat == nil {
		return nil
	}

	cat.Salary = salary

	return s.repo.UpdateCat(cat)
}

func (s *Service) DeleteCat(id int) error {
	return s.repo.DeleteCat(id)
}
