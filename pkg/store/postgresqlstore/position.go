package postgresqlstore

import (
	"database/sql"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
)

type positionStatements struct {
	create          *sql.Stmt
	get             *sql.Stmt
	getByUserID     *sql.Stmt
	delete          *sql.Stmt
	list            *sql.Stmt
	listByTrackerID *sql.Stmt
}

func (s *sqlStore) initPositionStatements() error {
	var err error

	if s.positionStatements.create, err = s.db.Prepare(`
	INSERT INTO positions (
		tracker_id,
		ts,
		lat,
		lon,
		alt,
		heading,
		speed,
		payload,
		precision
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9
	) RETURNING id
	`); err != nil {
		return err
	}

	if s.positionStatements.get, err = s.db.Prepare(`
	SELECT
		id,
		tracker_id,
		ts,
		lat,
		lon,
		alt,
		heading,
		speed,
		payload,
		precision
	FROM
		positions
	WHERE
		id=$1
	`); err != nil {
		return err
	}

	if s.positionStatements.getByUserID, err = s.db.Prepare(`
	SELECT
		positions.id,
		positions.tracker_id,
		positions.ts,
		positions.lat,
		positions.lon,
		positions.alt,
		positions.heading,
		positions.speed,
		positions.payload,
		positions.precision
	FROM
		positions,
		trackers,
		collections,
		team_members
	WHERE
		positions.tracker_id = trackers.id
		AND
		trackers.collection_id = collections.id
		AND
		collections.team_id = team_members.team_id
		AND
		positions.id=$1
		AND
		team_members.user_id=$2
	`); err != nil {
		return err
	}

	if s.positionStatements.delete, err = s.db.Prepare(`
	DELETE FROM positions
	WHERE id = $1
	`); err != nil {
		return err
	}

	if s.positionStatements.list, err = s.db.Prepare(`
	SELECT
		id,
		tracker_id,
		ts,
		lat,
		lon,
		alt,
		heading,
		speed,
		payload,
		precision
	FROM
		positions
	ORDER BY
		ts DESC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.positionStatements.listByTrackerID, err = s.db.Prepare(`
	SELECT
		positions.id,
		positions.tracker_id,
		positions.ts,
		positions.lat,
		positions.lon,
		positions.alt,
		positions.heading,
		positions.speed,
		positions.payload,
		positions.precision
	FROM
		positions,
		trackers,
		collections,
		team_members
	WHERE
		positions.tracker_id = trackers.id
		AND
		trackers.collection_id = collections.id
		AND
		collections.team_id = team_members.team_id
		AND
		tracker_id = $1
		AND
		team_members.user_id = $2
	ORDER BY
		ts DESC
	LIMIT $3
	OFFSET $4
	`); err != nil {
		return err
	}

	return err
}

func (s *sqlStore) CreatePosition(position *model.Position, userID int64) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return -1, errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfTracker(tx, userID, position.TrackerID)
	if err != nil {
		return -1, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	row := tx.Stmt(s.positionStatements.create).QueryRow(
		position.TrackerID,
		position.Timestamp,
		position.Lat,
		position.Lon,
		position.Alt,
		position.Heading,
		position.Speed,
		position.Payload,
		position.Precision,
	)

	lastInsertID, err := scanIDRow(row)

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.NewStorageErrorFromError(err)
	}

	return lastInsertID, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) GetPosition(positionID int64) (*model.Position, error) {
	row := s.positionStatements.get.QueryRow(
		positionID,
	)

	position, err := scanPositionRow(row)
	return &position, errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) GetPositionByUserID(positionID int64, userID int64) (*model.Position, error) {
	row := s.positionStatements.getByUserID.QueryRow(
		positionID,
		userID,
	)

	position, err := scanPositionRow(row)

	if err == sql.ErrNoRows {
		return &model.Position{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	return &position, errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) DeletePosition(positionID int64, userID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfPosition(tx, userID, positionID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.positionStatements.delete).Exec(positionID)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) ListPositions(offset int64, limit int64) ([]model.Position, error) {
	var positions []model.Position
	rows, err := s.positionStatements.list.Query(
		limit,
		offset,
	)

	if err != nil {
		return positions, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		position, err := scanPositionRow(rows)

		if err != nil {
			return positions, errors.NewStorageErrorFromError(err)
		}

		positions = append(positions, position)
	}

	return positions, nil
}

func (s *sqlStore) ListPositionsByTrackerID(trackerID int64, userID int64, offset int64, limit int64) ([]model.Position, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return []model.Position{}, errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfTracker(tx, userID, trackerID)
	if err != nil {
		return []model.Position{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	var positions []model.Position
	rows, err := tx.Stmt(s.positionStatements.listByTrackerID).Query(
		trackerID,
		userID,
		limit,
		offset,
	)

	if err != nil {
		_ = tx.Rollback()
		return positions, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		position, err := scanPositionRow(rows)

		if err != nil {
			_ = tx.Rollback()
			return positions, errors.NewStorageErrorFromError(err)
		}

		positions = append(positions, position)
	}

	return positions, errors.NewStorageErrorFromError(tx.Commit())
}

func scanPositionRow(row rowScanner) (model.Position, error) {
	position := model.Position{}

	err := row.Scan(
		&position.ID,
		&position.TrackerID,
		&position.Timestamp,
		&position.Lat,
		&position.Lon,
		&position.Alt,
		&position.Heading,
		&position.Speed,
		&position.Payload,
		&position.Precision,
	)

	return position, err
}
