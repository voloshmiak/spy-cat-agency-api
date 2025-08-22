package cat

import "time"

type Cat struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	YearsOfExperience int       `json:"years_of_experience"`
	Breed             string    `json:"breed"`
	Salary            float64   `json:"salary"`
	CreatedAt         time.Time `json:"created_at"`
}
