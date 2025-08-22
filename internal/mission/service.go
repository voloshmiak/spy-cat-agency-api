package mission

import "errors"

var (
	MaxTargetsErr = errors.New("maximum targets exceeded")
	NotFoundErr   = errors.New("not found")
	ConflictErr   = errors.New("conflict with current state")
	AssignedErr   = errors.New("cannot delete an assigned mission")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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
		return nil, err
	}
	if mission == nil {
		return nil, NotFoundErr
	}

	return mission, nil
}

func (s *Service) UpdateMission(id int, catID int, complete bool) error {
	mission, err := s.GetMission(id)
	if err != nil {
		return err
	}

	if complete {
		for _, t := range mission.Targets {
			if !t.Complete {
				return ConflictErr
			}
		}
	}

	mission.CatID = &catID
	mission.Complete = complete

	return s.repo.UpdateMission(mission)
}

func (s *Service) DeleteMission(id int) error {
	mission, err := s.GetMission(id)
	if err != nil {
		return err
	}

	if mission.CatID != nil {
		return AssignedErr
	}

	return s.repo.DeleteMission(id)
}

func (s *Service) GetTarget(missionID, targetID int) (*Target, error) {
	mission, err := s.repo.GetMissionByID(missionID)
	if err != nil {
		return nil, err
	}
	if mission == nil {
		return nil, NotFoundErr
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

func (s *Service) UpdateTarget(missionID, targetID int, notes string, complete bool) error {
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

	return s.repo.UpdateTarget(target)
}

func (s *Service) DeleteTarget(missionID, targetID int) error {
	target, err := s.GetTarget(missionID, targetID)
	if err != nil {
		return err
	}

	if target.Complete {
		return ConflictErr
	}

	return s.repo.DeleteTarget(targetID)
}
