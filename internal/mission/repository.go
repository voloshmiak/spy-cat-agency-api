package mission

import (
	"database/sql"
	"errors"
)

type Repository struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) GetAllMissions() ([]Mission, error) {
	query := `SELECT * FROM missions`

	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	missions := make([]Mission, 0)

	for rows.Next() {
		var mission Mission
		if err = rows.Scan(&mission.ID, &mission.CatID, &mission.Complete, &mission.CreatedAt); err != nil {
			return nil, err
		}
		missions = append(missions, mission)
	}

	return missions, nil
}

func (r *Repository) CreateMission(mission *Mission) (int, error) {
	query := `INSERT INTO missions (complete) VALUES ($1) RETURNING id`

	err := r.conn.QueryRow(query, mission.Complete).Scan(&mission.ID)
	if err != nil {
		return 0, err
	}

	query = `INSERT INTO targets (mission_id, name, country, complete) VALUES ($1, $2, $3, $4)`
	for _, target := range mission.Targets {
		_, err = r.conn.Exec(query, mission.ID, target.Name, target.Country, target.Complete)
		if err != nil {
			return 0, err
		}
	}

	return mission.ID, nil
}

func (r *Repository) GetMissionByID(id int) (*Mission, error) {
	query := `SELECT * FROM missions WHERE id = $1`

	var mission Mission
	err := r.conn.QueryRow(query, id).Scan(&mission.ID, &mission.CatID, &mission.Complete, &mission.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	query = `SELECT * FROM targets WHERE mission_id = $1`
	rows, err := r.conn.Query(query, mission.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := make([]Target, 0)
	for rows.Next() {
		var target Target
		if err = rows.Scan(&target.ID, &target.MissionID, &target.Name, &target.Country, &target.Notes, &target.Complete); err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	mission.Targets = targets

	return &mission, nil
}

func (r *Repository) UpdateMission(mission *Mission) error {
	query := `UPDATE missions SET cat_id = $1, complete = $2 WHERE id = $3`

	_, err := r.conn.Exec(query, mission.CatID, mission.Complete, mission.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteMission(id int) error {
	query := `DELETE FROM missions WHERE id = $1`

	_, err := r.conn.Exec(query, id)
	return err
}

func (r *Repository) AddTarget(target *Target) (int, error) {
	query := `INSERT INTO targets (mission_id, name, country, complete) VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.conn.QueryRow(query, target.MissionID, target.Name, target.Country, target.Complete).Scan(&target.ID)
	if err != nil {
		return 0, err
	}

	return target.ID, nil
}

func (r *Repository) UpdateTarget(target *Target) error {
	query := `UPDATE targets SET notes = $1, complete = $2 WHERE id = $3`

	_, err := r.conn.Exec(query, target.Notes, target.Complete, target.ID)
	return err
}

func (r *Repository) DeleteTarget(id int) error {
	query := `DELETE FROM targets WHERE id = $1`

	_, err := r.conn.Exec(query, id)
	return err
}
