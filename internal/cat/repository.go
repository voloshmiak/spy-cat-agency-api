package cat

import (
	"context"
	"database/sql"
	"errors"
	"log"
)

type Repository struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) GetAllCats() ([]Cat, error) {
	query := `SELECT * FROM cats`

	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cats := make([]Cat, 0)

	for rows.Next() {
		var cat Cat
		if err = rows.Scan(&cat.ID, &cat.Name, &cat.Breed, &cat.YearsOfExperience, &cat.Salary, &cat.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, cat)
	}

	return cats, nil
}

func (r *Repository) CreateCat(cat *Cat) (int, error) {
	query := `INSERT INTO cats (name, years_of_experience, breed, salary) VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.conn.QueryRow(query, cat.Name, cat.YearsOfExperience, cat.Breed, cat.Salary).Scan(&cat.ID)
	if err != nil {
		return 0, err
	}

	return cat.ID, nil
}

func (r *Repository) GetCatByID(id int) (*Cat, error) {
	query := `SELECT * FROM cats WHERE id = $1`

	var cat Cat
	err := r.conn.QueryRow(query, id).Scan(&cat.ID, &cat.Name, &cat.Breed, &cat.YearsOfExperience, &cat.Salary, &cat.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Println(err)
		return nil, err
	}

	return &cat, nil
}

func (r *Repository) UpdateCat(ctx context.Context, cat *Cat) error {
	query := `UPDATE cats SET name = $1, years_of_experience = $2, breed = $3, salary = $4 WHERE id = $5`

	_, err := r.conn.ExecContext(ctx, query, cat.Name, cat.YearsOfExperience, cat.Breed, cat.Salary, cat.ID)
	return err
}

func (r *Repository) DeleteCat(ctx context.Context, id int) error {
	query := `DELETE FROM cats WHERE id = $1`

	res, err := r.conn.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
