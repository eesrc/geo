package sqlitestore

import (
	"database/sql"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
)

type trackerStatements struct {
	create             *sql.Stmt
	get                *sql.Stmt
	getByUserID        *sql.Stmt
	update             *sql.Stmt
	delete             *sql.Stmt
	list               *sql.Stmt
	listByCollectionID *sql.Stmt
}

func (s *sqliteStore) initTrackerStatements() error {
	var err error

	if s.trackerStatements.create, err = s.db.Prepare(`
	INSERT INTO trackers (
		collection_id,
		name,
		description
	) VALUES (
		$1,
		$2,
		$3
	)
	`); err != nil {
		return err
	}

	if s.trackerStatements.get, err = s.db.Prepare(`
	SELECT
		id,
		collection_id,
		name,
		description
	FROM trackers
	WHERE id = $1
	`); err != nil {
		return err
	}

	if s.trackerStatements.getByUserID, err = s.db.Prepare(`
	SELECT
		trackers.id,
		trackers.collection_id,
		trackers.name,
		trackers.description
	FROM
		trackers,
		collections,
		team_members
	WHERE
		collections.id = trackers.collection_id
		AND
		collections.team_id = team_members.team_id
		AND
		trackers.id = $1
		AND
		team_members.user_id = $2
	`); err != nil {
		return err
	}

	if s.trackerStatements.update, err = s.db.Prepare(`
	UPDATE trackers
	SET
		collection_id = $1,
		name = $2,
		description = $3
	WHERE id = $4
	`); err != nil {
		return err
	}

	if s.trackerStatements.delete, err = s.db.Prepare(`
	DELETE FROM trackers
	WHERE id = $1
	`); err != nil {
		return err
	}

	if s.trackerStatements.list, err = s.db.Prepare(`
	SELECT
		id,
		collection_id,
		name,
		description
	FROM trackers
	ORDER BY
		id ASC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.trackerStatements.listByCollectionID, err = s.db.Prepare(`
	SELECT
		trackers.id,
		trackers.collection_id,
		trackers.name,
		trackers.description
	FROM
		trackers,
		collections,
		team_members
	WHERE
		trackers.collection_id = collections.id
		AND
		collections.team_id = team_members.team_id
		AND
		collections.id = $1
		AND
		team_members.user_id = $2
	ORDER BY
		trackers.id ASC
	LIMIT $3
	OFFSET $4`); err != nil {
		return err
	}

	return err
}

func (s *sqliteStore) CreateTracker(tracker *model.Tracker, userID int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return -1, errors.NewStorageErrorFromError(err)
	}

	_, err = s.ensureAdminOfCollection(tx, userID, tracker.CollectionID)
	if err != nil {
		return -1, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	row, err := tx.Stmt(s.trackerStatements.create).Exec(
		tracker.CollectionID,
		tracker.Name,
		tracker.Description,
	)

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.NewStorageErrorFromError(err)
	}

	lastInsertID, err := row.LastInsertId()

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.NewStorageErrorFromError(err)
	}

	return lastInsertID, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) GetTracker(trackerID int64) (*model.Tracker, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.trackerStatements.get.QueryRow(
		trackerID,
	)

	tracker, err := scanTrackerRow(row)

	return &tracker, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) GetTrackerByUserID(trackerID int64, userID int64) (*model.Tracker, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.trackerStatements.getByUserID.QueryRow(
		trackerID,
		userID,
	)

	tracker, err := scanTrackerRow(row)

	if err == sql.ErrNoRows {
		return &model.Tracker{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	return &tracker, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) UpdateTracker(tracker *model.Tracker, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfTracker(tx, userID, tracker.ID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	// We do this as the tracker can change collections
	_, err = s.ensureAdminOfCollection(tx, userID, tracker.CollectionID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.trackerStatements.update).Exec(
		tracker.CollectionID,
		tracker.Name,
		tracker.Description,
		tracker.ID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) DeleteTracker(trackerID int64, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfTracker(tx, userID, trackerID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.trackerStatements.delete).Exec(
		trackerID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) ListTrackers(offset int64, limit int64) ([]model.Tracker, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var trackers []model.Tracker
	rows, err := s.trackerStatements.list.Query(
		limit,
		offset,
	)

	if err != nil {
		return trackers, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		tracker, err := scanTrackerRow(rows)

		if err != nil {
			return trackers, errors.NewStorageErrorFromError(err)
		}

		trackers = append(trackers, tracker)
	}

	return trackers, nil
}

func (s *sqliteStore) ListTrackersByCollectionID(collectionID int64, userID int64, offset int64, limit int64) ([]model.Tracker, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return []model.Tracker{}, errors.NewStorageErrorFromError(err)
	}

	_, err = s.ensureAdminOfCollection(tx, userID, collectionID)
	if err != nil {
		return []model.Tracker{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	var trackers []model.Tracker
	rows, err := tx.Stmt(s.trackerStatements.listByCollectionID).Query(
		collectionID,
		userID,
		limit,
		offset,
	)

	if err != nil {
		_ = tx.Rollback()
		return trackers, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		tracker, err := scanTrackerRow(rows)

		if err != nil {
			_ = tx.Rollback()
			return trackers, errors.NewStorageErrorFromError(err)
		}

		trackers = append(trackers, tracker)
	}

	return trackers, errors.NewStorageErrorFromError(tx.Commit())
}

func scanTrackerRow(row rowScanner) (model.Tracker, error) {
	tracker := model.Tracker{}

	err := row.Scan(
		&tracker.ID,
		&tracker.CollectionID,
		&tracker.Name,
		&tracker.Description,
	)

	return tracker, err
}
