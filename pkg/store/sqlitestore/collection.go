package sqlitestore

import (
	"database/sql"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
)

type collectionStatements struct {
	create       *sql.Stmt
	get          *sql.Stmt
	getByUserID  *sql.Stmt
	update       *sql.Stmt
	delete       *sql.Stmt
	list         *sql.Stmt
	listByUserID *sql.Stmt
}

func (s *sqliteStore) initCollectionStatements() error {
	var err error

	if s.collectionStatements.create, err = s.db.Prepare(`
	INSERT INTO collections (
		team_id,
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

	if s.collectionStatements.get, err = s.db.Prepare(`
	SELECT
		id,
		team_id,
		name,
		description
	FROM collections
	WHERE id=$1
	`); err != nil {
		return err
	}

	if s.collectionStatements.getByUserID, err = s.db.Prepare(`
	SELECT
		collections.id,
		collections.team_id,
		collections.name,
		collections.description
	FROM
		collections,
		teams,
		team_members
	WHERE
		teams.id = collections.team_id
		AND
		teams.id = team_members.team_id
		AND
		collections.id=$1
		AND
		team_members.user_id = $2
	`); err != nil {
		return err
	}

	if s.collectionStatements.update, err = s.db.Prepare(`
	UPDATE collections
	SET
		team_id = $1,
		name = $2,
		description = $3
	WHERE id=$4
	`); err != nil {
		return err
	}

	if s.collectionStatements.delete, err = s.db.Prepare(`
	DELETE FROM collections
	WHERE id=$1
	`); err != nil {
		return err
	}

	if s.collectionStatements.list, err = s.db.Prepare(`
	SELECT
		id,
		team_id,
		name,
		description
	FROM collections
	ORDER BY
		id ASC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.collectionStatements.listByUserID, err = s.db.Prepare(`
	SELECT
		collections.id,
		collections.team_id,
		collections.name,
		collections.description
	FROM
		collections,
		team_members
	WHERE
		collections.team_id = team_members.team_id
		AND
		team_members.user_id = $1
	ORDER BY
		id ASC
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	return err
}

func (s *sqliteStore) CreateCollection(collection *model.Collection, userID int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return -1, errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfTeam(tx, userID, collection.TeamID)
	if err != nil {
		return -1, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	res, err := tx.Stmt(s.collectionStatements.create).Exec(
		collection.TeamID,
		collection.Name,
		collection.Description,
	)

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.NewStorageErrorFromError(err)
	}

	newCollectionID, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return -1, err
	}

	return newCollectionID, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) GetCollection(collectionID int64) (*model.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.collectionStatements.get.QueryRow(
		collectionID,
	)

	collection, err := scanCollectionRow(row)
	if err != nil {
		return &collection, errors.NewStorageErrorFromError(err)
	}

	return &collection, nil
}

func (s *sqliteStore) GetCollectionByUserID(collectionID int64, userID int64) (*model.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.collectionStatements.getByUserID.QueryRow(
		collectionID,
		userID,
	)

	collection, err := scanCollectionRow(row)
	if err == sql.ErrNoRows {
		return &model.Collection{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	return &collection, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) UpdateCollection(collection *model.Collection, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	_, err = s.ensureAdminOfCollection(tx, userID, collection.ID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	// We do this as the collection can change teams
	err = s.ensureAdminOfTeam(tx, userID, collection.TeamID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.collectionStatements.update).Exec(
		collection.TeamID,
		collection.Name,
		collection.Description,
		collection.ID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) DeleteCollection(collectionID int64, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	_, err = s.ensureAdminOfCollection(tx, userID, collectionID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.collectionStatements.delete).Exec(
		collectionID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) ListCollections(offset int64, limit int64) ([]model.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var collections []model.Collection
	rows, err := s.collectionStatements.list.Query(
		limit,
		offset,
	)

	if err != nil {
		return collections, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		collection, err := scanCollectionRow(rows)

		if err != nil {
			return collections, errors.NewStorageErrorFromError(err)
		}

		collections = append(collections, collection)
	}

	return collections, nil
}

func (s *sqliteStore) ListCollectionsByUserID(userID int64, offset int64, limit int64) ([]model.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var collections []model.Collection
	rows, err := s.collectionStatements.listByUserID.Query(
		userID,
		limit,
		offset,
	)

	if err != nil {
		return collections, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		collection, err := scanCollectionRow(rows)

		if err != nil {
			return collections, errors.NewStorageErrorFromError(err)
		}

		collections = append(collections, collection)
	}

	return collections, nil
}

func scanCollectionRow(row rowScanner) (model.Collection, error) {
	collection := model.Collection{}

	err := row.Scan(
		&collection.ID,
		&collection.TeamID,
		&collection.Name,
		&collection.Description,
	)

	return collection, err
}
