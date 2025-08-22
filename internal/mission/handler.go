package mission

import (
	"errors"
	"github.com/gin-gonic/gin"
	"spy-cat-agency/internal/cat"
	"strconv"
)

type MissionsResponse struct {
	ID       int  `json:"id"`
	CatID    *int `json:"cat_id"`
	Complete bool `json:"complete"`
}

type TargetRequest struct {
	Name    string `json:"name" binding:"required"`
	Country string `json:"country" binding:"required"`
}

type CreateMissionRequest struct {
	Targets []TargetRequest `json:"targets" binding:"required,min=1,max=3"`
}

type UpdateMissionRequest struct {
	CatID    int  `json:"cat_id"`
	Complete bool `json:"complete"`
}

type UpdateTargetRequest struct {
	Notes    string `json:"notes"`
	Complete bool   `json:"complete"`
}

type Handler struct {
	MissionService *Service
	CatService     *cat.Service
}

func NewHandler(missionService *Service, catService *cat.Service) *Handler {
	return &Handler{MissionService: missionService, CatService: catService}
}

func (h *Handler) ListMissions(c *gin.Context) {
	missions, err := h.MissionService.ListMissions()
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	missionsResponse := make([]MissionsResponse, 0)
	for _, m := range missions {
		missionsResponse = append(missionsResponse, MissionsResponse{
			ID:       m.ID,
			CatID:    m.CatID,
			Complete: m.Complete,
		})
	}

	c.JSON(200, missionsResponse)
}

func (h *Handler) CreateMission(c *gin.Context) {
	var missionRequest CreateMissionRequest
	err := c.ShouldBindJSON(&missionRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": "The request body is invalid or missing required fields"})
		return
	}

	id, err := h.MissionService.CreateMission(missionRequest.Targets)
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	c.JSON(201, gin.H{"id": id})
}

func (h *Handler) GetMission(c *gin.Context) {
	stringID := c.Param("id")
	id, err := strconv.Atoi(stringID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	mission, err := h.MissionService.GetMission(id)
	if err != nil {
		switch {
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist."})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	c.JSON(200, mission)
}

func (h *Handler) UpdateMission(c *gin.Context) {
	var missionRequest UpdateMissionRequest
	err := c.ShouldBindJSON(&missionRequest)
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

	_, err = h.CatService.GetCat(missionRequest.CatID)
	if err != nil {
		switch {
		case errors.Is(err, cat.NotFoundErr):
			c.JSON(400, gin.H{"error": "Provided cat_id does not exist"})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	err = h.MissionService.UpdateMission(id, missionRequest.CatID, missionRequest.Complete)
	if err != nil {
		switch {
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist."})
			return
		case errors.Is(err, ConflictErr):
			c.JSON(409, gin.H{"error": "All targets must be complete before a mission can be marked as complete."})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	c.JSON(200, gin.H{"id": id, "cat_id": missionRequest.CatID, "complete": missionRequest.Complete})
}

func (h *Handler) DeleteMission(c *gin.Context) {
	stringID := c.Param("id")
	id, err := strconv.Atoi(stringID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	err = h.MissionService.DeleteMission(id)
	if err != nil {
		switch {
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist."})
			return
		case errors.Is(err, AssignedErr):
			c.JSON(409, gin.H{"error": "The request could not be completed due to a conflict with the current state of the resource."})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	c.Status(204)
}

func (h *Handler) AddTarget(c *gin.Context) {
	var targetRequest TargetRequest
	err := c.ShouldBindJSON(&targetRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": "The request body is invalid or missing required fields"})
		return
	}

	stringMissionID := c.Param("id")
	missionID, err := strconv.Atoi(stringMissionID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	targetID, err := h.MissionService.AddTarget(missionID, targetRequest.Name, targetRequest.Country)
	if err != nil {
		switch {
		case errors.Is(err, MaxTargetsErr):
			c.JSON(400, gin.H{"error": "A mission cannot have more than 3 targets"})
			return
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist."})
			return
		case errors.Is(err, ConflictErr):
			c.JSON(409, gin.H{"error": "The request could not be completed due to a conflict with the current state of the resource."})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	c.JSON(201, gin.H{"id": targetID})
}

func (h *Handler) UpdateTarget(c *gin.Context) {
	var targetRequest UpdateTargetRequest
	err := c.ShouldBindJSON(&targetRequest)
	if err != nil {
		c.JSON(400, gin.H{"error": "The request body is invalid or missing required fields"})
		return
	}

	stringMissionID := c.Param("id")
	missionID, err := strconv.Atoi(stringMissionID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	stringTargetID := c.Param("target_id")
	targetID, err := strconv.Atoi(stringTargetID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	err = h.MissionService.UpdateTarget(missionID, targetID, targetRequest.Notes, targetRequest.Complete)
	if err != nil {
		switch {
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist."})
			return
		case errors.Is(err, ConflictErr):
			c.JSON(409, gin.H{"error": "The request could not be completed due to a conflict with the current state of the resource."})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	c.JSON(200, gin.H{"id": targetID, "notes": targetRequest.Notes, "complete": targetRequest.Complete})
}

func (h *Handler) DeleteTarget(c *gin.Context) {
	stringMissionID := c.Param("id")
	missionID, err := strconv.Atoi(stringMissionID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	stringTargetID := c.Param("target_id")
	targetID, err := strconv.Atoi(stringTargetID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	err = h.MissionService.DeleteTarget(missionID, targetID)
	if err != nil {
		switch {
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist"})
			return
		case errors.Is(err, ConflictErr):
			c.JSON(409, gin.H{"error": "The request could not be completed due to a conflict with the current state of the resource."})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	c.Status(204)
}
