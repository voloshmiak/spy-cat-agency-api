package cat

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type CreateCatRequest struct {
	Name              string  `json:"name" binding:"required"`
	Breed             string  `json:"breed" binding:"required"`
	YearsOfExperience int     `json:"years_of_experience" binding:"required"`
	Salary            float64 `json:"salary" binding:"required"`
}
type Breed struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UpdateCatSalaryRequest struct {
	Salary float64 `json:"salary" binding:"required"`
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
	c.JSON(200, cats)
}

func (h *Handler) CreateCat(c *gin.Context) {
	var catRequest CreateCatRequest
	err := c.ShouldBindJSON(&catRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": "The request body is invalid or missing required fields"})
		return
	}

	isValid, err := validateBreed(catRequest.Breed)
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	if !isValid {
		c.JSON(400, gin.H{"error": "The specified breed is not recognized."})
		return
	}

	id, err := h.Service.CreateCat(catRequest.Name, catRequest.Breed, catRequest.YearsOfExperience, catRequest.Salary)
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server."})
		return
	}
	c.JSON(201, gin.H{"id": id})
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

	c.JSON(200, cat)
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

	err = h.Service.UpdateCatSalary(id, catRequest.Salary)
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server."})
		return
	}

	c.JSON(200, gin.H{"id": id, "new_salary": catRequest.Salary})
}

func (h *Handler) DeleteCat(c *gin.Context) {
	stringID := c.Param("id")
	id, err := strconv.Atoi(stringID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	err = h.Service.DeleteCat(id)
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server."})
		return
	}

	c.Status(204)
}

func validateBreed(breedName string) (bool, error) {
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
		if strings.EqualFold(b.Name, breedName) {
			return true, nil
		}
	}

	return false, nil
}
