package mission

import (
	"context"
	"database/sql"
	"errors"
	"spy-cat-agency/internal/cat"
)

var (
	MaxTargetsErr = errors.New("maximum targets exceeded")
	NotFoundErr   = errors.New("not found")
	ConflictErr   = errors.New("conflict with current state")
	AssignedErr   = errors.New("cannot delete an assigned mission")
	CatBusyErr    = errors.New("cat is already assigned to an active mission")
)

type Service struct {
	repo       *Repository
	catService *cat.Service
}

func NewService(repo *Repository, catService *cat.Service) *Service {
	return &Service{
		repo:       repo,
		catService: catService,
	}
}

func (s *Service) ListMissions() ([]Mission, error) {
	missions, err := s.repo.GetAllMissions()
	if err != nil {
		return nil, err
	}

	return missions, nil
}

func (s *Service) CreateMission(targetsRequest []TargetRequest) (int, error) {
	var targets []Target
	for _, t := range targetsRequest {
		targets = append(targets, Target{
			Name:     t.Name,
			Country:  t.Country,
			Notes:    "",
			Complete: false,
		})
	}

	mission := &Mission{
		Complete: false,
		Targets:  targets,
	}

	return s.repo.CreateMission(mission)
}

func (s *Service) GetMission(id int) (*Mission, error) {
	mission, err := s.repo.GetMissionByID(id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, NotFoundErr
		}
		return nil, err
	}

	return mission, nil
}

func (s *Service) UpdateMission(ctx context.Context, id int, r UpdateMissionRequest) (*Mission, error) {
	if r.CatID == nil && r.Complete == nil {
		return nil, nil
	}

	if r.CatID != nil {
		_, err := s.catService.GetCat(*r.CatID)
		if err != nil {
			return nil, err
		}

		activeMission, err := s.repo.FindActiveMissionByCatID(ctx, *r.CatID)
		if err != nil {
			return nil, err
		}

		if activeMission != nil && activeMission.ID != id {
			return nil, CatBusyErr
		}
	}

	if r.Complete != nil && *r.Complete {
		mission, err := s.GetMission(id)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, NotFoundErr
			}
			return nil, err
		}
		for _, t := range mission.Targets {
			if !t.Complete {
				return nil, ConflictErr
			}
		}
	}

	mission, err := s.repo.UpdateMission(ctx, id, r)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFoundErr
		}
		return nil, err
	}

	return mission, nil
}

func (s *Service) DeleteMission(ctx context.Context, id int) error {
	mission, err := s.GetMission(id)
	if err != nil {
		return err
	}

	if mission.CatID != nil {
		return AssignedErr
	}

	err = s.repo.DeleteMission(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return NotFoundErr
		}
		return err
	}

	return nil
}

func (s *Service) GetTarget(missionID, targetID int) (*Target, error) {
	mission, err := s.repo.GetMissionByID(missionID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, NotFoundErr
		}
		return nil, err
	}

	for i := range mission.Targets {
		if mission.Targets[i].ID == targetID {
			return &mission.Targets[i], nil
		}
	}

	return nil, NotFoundErr
}

func (s *Service) AddTarget(missionID int, name, country string) (int, error) {
	mission, err := s.GetMission(missionID)
	if err != nil {
		return 0, err
	}
	if len(mission.Targets) >= 3 {
		return 0, MaxTargetsErr
	}
	if mission.Complete {
		return 0, ConflictErr
	}

	target := &Target{
		MissionID: missionID,
		Name:      name,
		Country:   country,
		Notes:     "",
		Complete:  false,
	}

	return s.repo.AddTarget(target)
}

func (s *Service) UpdateTarget(ctx context.Context, missionID, targetID int, notes string, complete bool) error {
	mission, err := s.GetMission(missionID)
	if err != nil {
		return err
	}

	if mission.Complete {
		return ConflictErr
	}

	target, err := s.GetTarget(missionID, targetID)
	if err != nil {
		return err
	}

	if target.Complete {
		return ConflictErr
	}

	target.Notes = notes
	target.Complete = complete

	return s.repo.UpdateTarget(ctx, target)
}

func (s *Service) DeleteTarget(ctx context.Context, missionID, targetID int) error {
	target, err := s.GetTarget(missionID, targetID)
	if err != nil {
		return err
	}

	if target.Complete {
		return ConflictErr
	}

	err = s.repo.DeleteTarget(ctx, targetID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return NotFoundErr
		}
		return err
	}

	return nil
}
