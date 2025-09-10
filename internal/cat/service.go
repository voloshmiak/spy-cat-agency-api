package cat

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

var (
	NotFoundErr   = errors.New("not found")
	WrongBreedErr = errors.New("the specified breed is not recognized")
)

type Breed struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

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
	isValid, err := validateBreed(breed)
	if err != nil {
		return 0, WrongBreedErr
	}

	if !isValid {
		return 0, WrongBreedErr
	}

	cat := &Cat{
		Name:              name,
		Breed:             breed,
		YearsOfExperience: yearsOfExperience,
		Salary:            salary,
	}

	return s.repo.CreateCat(cat)
}

func (s *Service) GetCat(id int) (*Cat, error) {
	cat, err := s.repo.GetCatByID(id)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, NotFoundErr
	}
	return cat, nil
}

func (s *Service) UpdateCatSalary(ctx context.Context, id int, salary float64) error {
	cat, err := s.repo.GetCatByID(id)
	if err != nil {
		return err
	}

	cat.Salary = salary

	return s.repo.UpdateCat(ctx, cat)
}

func (s *Service) DeleteCat(ctx context.Context, id int) error {
	err := s.repo.DeleteCat(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return NotFoundErr
		}
		return err
	}

	return nil
}

func validateBreed(breed string) (bool, error) {
	resp, err := http.Get("https://api.thecatapi.com/v1/breeds")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var breeds []Breed
	err = json.Unmarshal(body, &breeds)
	if err != nil {
		return false, err
	}

	for _, b := range breeds {
		if strings.EqualFold(b.Name, breed) {
			return true, nil
		}
	}

	return false, nil
}
