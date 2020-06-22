package sqlitestore

import (
	"database/sql"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
)

type teamStatements struct {
	create       *sql.Stmt
	get          *sql.Stmt
	getByUserID  *sql.Stmt
	update       *sql.Stmt
	delete       *sql.Stmt
	list         *sql.Stmt
	listByUserID *sql.Stmt
	addMember    *sql.Stmt
	removeMember *sql.Stmt
}

func (s *sqliteStore) initTeamStatements() error {
	var err error

	if s.teamStatements.create, err = s.db.Prepare(`
	INSERT INTO teams (
		name,
		description
	) VALUES (
		$1,
		$2
	)
	`); err != nil {
		return err
	}

	if s.teamStatements.get, err = s.db.Prepare(`
	SELECT
		id,
		name,
		description
	FROM
		teams
	WHERE id = $1
	`); err != nil {
		return err
	}

	if s.teamStatements.getByUserID, err = s.db.Prepare(`
	SELECT
		teams.id,
		teams.name,
		teams.description
	FROM
		teams, team_members
	WHERE
		team_members.team_id = teams.id
		AND
		team_members.team_id = $1
		AND
		team_members.user_id = $2
	`); err != nil {
		return err
	}

	if s.teamStatements.update, err = s.db.Prepare(`
	UPDATE teams
	SET
		name = $1,
		description = $2
	WHERE id = $3
	`); err != nil {
		return err
	}

	if s.teamStatements.delete, err = s.db.Prepare(`
	DELETE FROM teams
	WHERE id=$1
	`); err != nil {
		return err
	}

	if s.teamStatements.list, err = s.db.Prepare(`
	SELECT
		id,
		name,
		description
	FROM
		teams
	ORDER BY
		id ASC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.teamStatements.listByUserID, err = s.db.Prepare(`
	SELECT
		teams.id,
		teams.name,
		teams.description
	FROM
		teams
	LEFT JOIN
		team_members
	ON
		team_members.team_id = teams.id
	WHERE
		team_members.user_id = $1
	ORDER BY
		teams.id ASC
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	if s.teamStatements.addMember, err = s.db.Prepare(`
	INSERT INTO team_members (
		user_id,
		team_id,
		admin
	) VALUES(
		$1,
		$2,
		$3
	)`); err != nil {
		return err
	}

	if s.teamStatements.removeMember, err = s.db.Prepare(`
	DELETE FROM team_members
	WHERE
		user_id = $1
		AND
		team_id = $2
	`); err != nil {
		return err
	}

	return err
}

func (s *sqliteStore) CreateTeam(team *model.Team) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	r, err := s.teamStatements.create.Exec(
		team.Name,
		team.Description,
	)

	if err != nil {
		return -1, errors.NewStorageErrorFromError(err)
	}
	return r.LastInsertId()
}

func (s *sqliteStore) GetTeam(teamID int64) (*model.Team, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.teamStatements.get.QueryRow(
		teamID,
	)

	team, err := scanTeamRow(row)

	return &team, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) GetTeamByUserID(teamID int64, userID int64) (*model.Team, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.teamStatements.getByUserID.QueryRow(
		teamID,
		userID,
	)

	team, err := scanTeamRow(row)

	if err == sql.ErrNoRows {
		return &model.Team{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	return &team, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) UpdateTeam(team *model.Team, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfTeam(tx, userID, team.ID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.teamStatements.update).Exec(
		team.Name,
		team.Description,
		team.ID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) DeleteTeam(teamID int64, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfTeam(tx, userID, teamID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.teamStatements.delete).Exec(
		teamID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) ListTeams(offset int64, limit int64) ([]model.Team, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var teams []model.Team
	rows, err := s.teamStatements.list.Query(
		limit,
		offset,
	)

	if err != nil {
		return teams, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		team, err := scanTeamRow(rows)

		if err != nil {
			return teams, errors.NewStorageErrorFromError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (s *sqliteStore) ListTeamsByUserID(userID int64, offset int64, limit int64) ([]model.Team, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var teams []model.Team
	rows, err := s.teamStatements.listByUserID.Query(
		userID,
		limit,
		offset,
	)

	if err != nil {
		return teams, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		team, err := scanTeamRow(rows)

		if err != nil {
			return teams, errors.NewStorageErrorFromError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (s *sqliteStore) SetTeamMember(userID int64, teamID int64, admin bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.teamStatements.addMember.Exec(
		userID,
		teamID,
		admin,
	)

	return errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) RemoveTeamMember(user int64, team int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.teamStatements.removeMember.Exec(
		user,
		team,
	)

	return errors.NewStorageErrorFromError(err)
}

func scanTeamRow(row rowScanner) (model.Team, error) {
	team := model.Team{}

	err := row.Scan(
		&team.ID,
		&team.Name,
		&team.Description,
	)

	return team, err
}
