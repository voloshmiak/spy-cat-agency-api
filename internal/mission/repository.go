package mission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
	tx, err := r.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	missionQuery := `INSERT INTO missions (complete) VALUES ($1) RETURNING id`
	err = tx.QueryRow(missionQuery, mission.Complete).Scan(&mission.ID)
	if err != nil {
		return 0, err
	}

	if len(mission.Targets) == 0 {
		if err = tx.Commit(); err != nil {
			return 0, err
		}
		return mission.ID, nil
	}

	valueStrings := make([]string, 0, len(mission.Targets))
	valueArgs := make([]interface{}, 0, len(mission.Targets)*4)
	i := 1
	for _, target := range mission.Targets {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i, i+1, i+2, i+3))
		valueArgs = append(valueArgs, mission.ID, target.Name, target.Country, target.Complete)
		i += 4
	}

	targetsQuery := fmt.Sprintf(
		"INSERT INTO targets (mission_id, name, country, complete) VALUES %s",
		strings.Join(valueStrings, ","),
	)

	_, err = tx.Exec(targetsQuery, valueArgs...)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return mission.ID, nil
}

func (r *Repository) GetMissionByID(id int) (*Mission, error) {
	query := `
		SELECT
			m.id, m.cat_id, m.complete, m.created_at,
			t.id, t.mission_id, t.name, t.country, t.notes, t.complete
		FROM
			missions m
		LEFT JOIN
			targets t ON m.id = t.mission_id
		WHERE
			m.id = $1`

	rows, err := r.conn.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mission *Mission
	targets := make([]Target, 0)

	for rows.Next() {
		var target Target
		var targetID sql.NullInt64
		var targetMissionID sql.NullInt64
		var targetName sql.NullString
		var targetCountry sql.NullString
		var targetNotes sql.NullString
		var targetComplete sql.NullBool

		if mission == nil {
			mission = &Mission{}
		}

		err := rows.Scan(
			&mission.ID, &mission.CatID, &mission.Complete, &mission.CreatedAt,
			&targetID, &targetMissionID, &targetName, &targetCountry, &targetNotes, &targetComplete,
		)
		if err != nil {
			return nil, err
		}

		if targetID.Valid {
			target.ID = int(targetID.Int64)
			target.MissionID = int(targetMissionID.Int64)
			target.Name = targetName.String
			target.Country = targetCountry.String
			target.Notes = targetNotes.String
			target.Complete = targetComplete.Bool
			targets = append(targets, target)
		}
	}

	if mission == nil {
		return nil, sql.ErrNoRows
	}

	mission.Targets = targets

	return mission, nil
}

func (r *Repository) UpdateMission(ctx context.Context, id int, req UpdateMissionRequest) (*Mission, error) {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argID := 1

	if req.CatID != nil {
		setValues = append(setValues, fmt.Sprintf("cat_id = $%d", argID))
		args = append(args, *req.CatID)
		argID++
	}

	if req.Complete != nil {
		setValues = append(setValues, fmt.Sprintf("complete = $%d", argID))
		args = append(args, *req.Complete)
		argID++
	}

	if len(setValues) == 0 {
		return nil, nil
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE missions SET %s WHERE id = $%d RETURNING id, cat_id, complete, created_at",
		strings.Join(setValues, ", "), argID)

	var mission Mission
	err := r.conn.QueryRowContext(ctx, query, args...).Scan(
		&mission.ID,
		&mission.CatID,
		&mission.Complete,
		&mission.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	fullMission, err := r.GetMissionByID(mission.ID)
	if err != nil {
		return nil, err
	}

	return fullMission, nil
}

func (r *Repository) DeleteMission(ctx context.Context, id int) error {
	query := `DELETE FROM missions WHERE id = $1`

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

func (r *Repository) AddTarget(target *Target) (int, error) {
	query := `INSERT INTO targets (mission_id, name, country, complete) VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.conn.QueryRow(query, target.MissionID, target.Name, target.Country, target.Complete).Scan(&target.ID)
	if err != nil {
		return 0, err
	}

	return target.ID, nil
}

func (r *Repository) UpdateTarget(ctx context.Context, target *Target) error {
	query := `UPDATE targets SET notes = $1, complete = $2 WHERE id = $3`

	_, err := r.conn.ExecContext(ctx, query, target.Notes, target.Complete, target.ID)
	return err
}

func (r *Repository) DeleteTarget(ctx context.Context, id int) error {
	query := `DELETE FROM targets WHERE id = $1`

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

func (r *Repository) FindActiveMissionByCatID(ctx context.Context, catID int) (*Mission, error) {
	query := `SELECT id, cat_id, complete FROM missions WHERE cat_id = $1 AND complete = false`

	var mission Mission
	err := r.conn.QueryRowContext(ctx, query, catID).Scan(&mission.ID, &mission.CatID, &mission.Complete)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &mission, nil
}
