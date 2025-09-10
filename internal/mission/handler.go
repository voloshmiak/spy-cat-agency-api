package mission

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"spy-cat-agency/internal/cat"
	"strconv"
)

type MissionResponse struct {
	ID       int  `json:"id"`
	CatID    *int `json:"cat_id"`
	Complete bool `json:"complete"`
}

type ListMissionsResponse struct {
	Missions []MissionResponse `json:"missions"`
}

type CreateMissionRequest struct {
	Targets []TargetRequest `json:"targets" binding:"required,min=1,max=3"`
}

type CreateMissionResponse struct {
	ID int `json:"id"`
}

type GetMissionResponse struct {
	ID       int      `json:"id"`
	CatID    *int     `json:"cat_id"`
	Complete bool     `json:"complete"`
	Targets  []Target `json:"targets"`
}

type UpdateMissionRequest struct {
	CatID    *int  `json:"cat_id"`
	Complete *bool `json:"complete"`
}

type UpdateMissionResponse struct {
	ID       int  `json:"id"`
	CatID    *int `json:"cat_id"`
	Complete bool `json:"complete"`
}

type TargetRequest struct {
	Name    string `json:"name" binding:"required"`
	Country string `json:"country" binding:"required"`
}

type AddTargetResponse struct {
	ID int `json:"id"`
}

type UpdateTargetRequest struct {
	Notes    string `json:"notes" binding:"required"`
	Complete bool   `json:"complete" binding:"required"`
}

type UpdateTargetResponse struct {
	ID       int    `json:"id"`
	Notes    string `json:"notes"`
	Complete bool   `json:"complete"`
}

type Handler struct {
	MissionService *Service
}

func NewHandler(missionService *Service) *Handler {
	return &Handler{MissionService: missionService}
}

func (h *Handler) ListMissions(c *gin.Context) {
	missions, err := h.MissionService.ListMissions()
	if err != nil {
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	response := ListMissionsResponse{}
	var missionsResp []MissionResponse
	for _, m := range missions {
		missionsResp = append(missionsResp, MissionResponse{
			ID:       m.ID,
			CatID:    m.CatID,
			Complete: m.Complete,
		})
	}

	response.Missions = missionsResp

	c.JSON(200, response)
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

	response := CreateMissionResponse{
		ID: id,
	}

	c.JSON(201, response)
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

	response := GetMissionResponse{
		ID:       mission.ID,
		CatID:    mission.CatID,
		Complete: mission.Complete,
		Targets:  mission.Targets,
	}

	c.JSON(200, response)
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

	ctx := c.Request.Context()

	updatedMission, err := h.MissionService.UpdateMission(ctx, id, missionRequest)
	if err != nil {
		switch {
		case errors.Is(err, CatBusyErr):
			c.JSON(400, gin.H{"error": "The cat is already assigned to another active mission"})
			return
		case errors.Is(err, cat.NotFoundErr):
			c.JSON(400, gin.H{"error": "Provided cat_id does not exist"})
			return
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist."})
			return
		case errors.Is(err, ConflictErr):
			c.JSON(409, gin.H{"error": "All targets must be complete before a mission can be marked as complete."})
			return
		}
		fmt.Println(err)
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	response := UpdateMissionResponse{
		ID:       updatedMission.ID,
		CatID:    updatedMission.CatID,
		Complete: updatedMission.Complete,
	}

	c.JSON(200, response)
}

func (h *Handler) DeleteMission(c *gin.Context) {
	stringID := c.Param("id")
	id, err := strconv.Atoi(stringID)
	if err != nil {
		c.JSON(404, gin.H{"error": "The requested resource does not exist."})
		return
	}

	ctx := c.Request.Context()

	err = h.MissionService.DeleteMission(ctx, id)
	if err != nil {
		fmt.Println(err)
		switch {
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist."})
			return
		case errors.Is(err, AssignedErr):
			c.JSON(409, gin.H{"error": "Missions assigned to a cat cannot be deleted."})
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
			c.JSON(409, gin.H{"error": "Mission is already complete"})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	response := AddTargetResponse{
		ID: targetID,
	}

	c.JSON(201, response)
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

	ctx := c.Request.Context()

	err = h.MissionService.UpdateTarget(ctx, missionID, targetID, targetRequest.Notes, targetRequest.Complete)
	if err != nil {
		switch {
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist."})
			return
		case errors.Is(err, ConflictErr):
			c.JSON(409, gin.H{"error": "Target's mission is already complete"})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	response := UpdateTargetResponse{
		ID:       targetID,
		Notes:    targetRequest.Notes,
		Complete: targetRequest.Complete,
	}

	c.JSON(200, response)
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

	ctx := c.Request.Context()

	err = h.MissionService.DeleteTarget(ctx, missionID, targetID)
	if err != nil {
		switch {
		case errors.Is(err, NotFoundErr):
			c.JSON(404, gin.H{"error": "The requested resource does not exist"})
			return
		case errors.Is(err, ConflictErr):
			c.JSON(409, gin.H{"error": "Target's mission is already complete"})
			return
		}
		c.JSON(500, gin.H{"error": "An unexpected error occurred on the server"})
		return
	}

	c.Status(204)
}
