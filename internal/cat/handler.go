package cat

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strconv"
)

type ListCatsResponse struct {
	Cats []CatResponse `json:"cats"`
}

type CatResponse struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	YearsOfExperience int     `json:"years_of_experience"`
	Breed             string  `json:"breed"`
	Salary            float64 `json:"salary"`
}

type CreateCatRequest struct {
	Name              string  `json:"name" binding:"required"`
	Breed             string  `json:"breed" binding:"required"`
	YearsOfExperience int     `json:"years_of_experience" binding:"required"`
	Salary            float64 `json:"salary" binding:"required"`
}

type CreateCatResponse struct {
	ID int `json:"id"`
}

type UpdateCatSalaryRequest struct {
	Salary float64 `json:"salary" binding:"required"`
}

type UpdateCatSalaryResponse struct {
	ID        int     `json:"id"`
	NewSalary float64 `json:"new_salary"`
}

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) ListCats(c *gin.Context) {
	cats, err := h.Service.ListCats()
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	response := ListCatsResponse{}
	var catsResponse []CatResponse
	for _, cat := range cats {
		catResponse := CatResponse{
			ID:                cat.ID,
			Name:              cat.Name,
			YearsOfExperience: cat.YearsOfExperience,
			Breed:             cat.Breed,
			Salary:            cat.Salary,
		}
		catsResponse = append(catsResponse, catResponse)
	}

	response.Cats = catsResponse

	c.JSON(200, response)
}

func (h *Handler) CreateCat(c *gin.Context) {
	var catRequest CreateCatRequest
	err := c.ShouldBindJSON(&catRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": "The request body is invalid or missing required fields"})
		return
	}

	id, err := h.Service.CreateCat(catRequest.Name, catRequest.Breed, catRequest.YearsOfExperience, catRequest.Salary)
	if err != nil {
		switch {
		case errors.Is(err, WrongBreedErr):
			c.JSON(400, gin.H{"error": "The specified breed is not recognized."})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server."})
		return
	}

	response := CreateCatResponse{
		ID: id,
	}

	c.JSON(201, response)
}

func (h *Handler) GetCat(c *gin.Context) {
	stringID := c.Param("id")
	id, err := strconv.Atoi(stringID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	cat, err := h.Service.GetCat(id)
	if err != nil {
		switch {
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist."})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server."})
		return
	}

	response := CatResponse{
		ID:                cat.ID,
		Name:              cat.Name,
		YearsOfExperience: cat.YearsOfExperience,
		Breed:             cat.Breed,
		Salary:            cat.Salary,
	}

	c.JSON(200, response)
}

func (h *Handler) UpdateCat(c *gin.Context) {
	var catRequest UpdateCatSalaryRequest
	err := c.ShouldBindJSON(&catRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": "The request body is invalid or missing required fields"})
		return
	}

	stringID := c.Param("id")
	id, err := strconv.Atoi(stringID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	ctx := c.Request.Context()

	err = h.Service.UpdateCatSalary(ctx, id, catRequest.Salary)
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server."})
		return
	}

	response := UpdateCatSalaryResponse{
		ID:        id,
		NewSalary: catRequest.Salary,
	}

	c.JSON(200, response)
}

func (h *Handler) DeleteCat(c *gin.Context) {
	stringID := c.Param("id")
	id, err := strconv.Atoi(stringID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	ctx := c.Request.Context()

	err = h.Service.DeleteCat(ctx, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server."})
		return
	}

	c.Status(204)
}
