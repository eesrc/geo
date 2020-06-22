package postgresqlstore

import (
	"database/sql"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
)

type shapeCollectionStatements struct {
	create       *sql.Stmt
	get          *sql.Stmt
	getByUserID  *sql.Stmt
	update       *sql.Stmt
	delete       *sql.Stmt
	list         *sql.Stmt
	listByTeamID *sql.Stmt
	listByUserID *sql.Stmt
}

func (s *sqlStore) initShapeCollectionStatements() error {
	var err error

	if s.shapeCollectionStatements.create, err = s.db.Prepare(`
	INSERT INTO shape_collections (
		team_id,
		name,
		description
	) VALUES (
		$1,
		$2,
		$3
	) RETURNING id
	`); err != nil {
		return err
	}

	if s.shapeCollectionStatements.get, err = s.db.Prepare(`
	SELECT
		id,
		team_id,
		name,
		description
	FROM
		shape_collections
	WHERE id = $1
	`); err != nil {
		return err
	}

	if s.shapeCollectionStatements.getByUserID, err = s.db.Prepare(`
	SELECT
		shape_collections.id,
		shape_collections.team_id,
		shape_collections.name,
		shape_collections.description
	FROM
		shape_collections,
		team_members
	WHERE
		shape_collections.team_id = team_members.team_id
		AND
		shape_collections.id = $1
		AND
		team_members.user_id = $2
	`); err != nil {
		return err
	}

	if s.shapeCollectionStatements.update, err = s.db.Prepare(`
	UPDATE shape_collections
	SET
		team_id = $1,
		name = $2,
		description = $3
	WHERE id=$4
	`); err != nil {
		return err
	}

	if s.shapeCollectionStatements.delete, err = s.db.Prepare(`
	DELETE FROM shape_collections
	WHERE id=$1
	`); err != nil {
		return err
	}

	if s.shapeCollectionStatements.list, err = s.db.Prepare(`
	SELECT
		id,
		team_id,
		name,
		description
	FROM shape_collections
	ORDER BY
		id ASC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.shapeCollectionStatements.listByTeamID, err = s.db.Prepare(`
	SELECT
		id,
		team_id,
		name,
		description
	FROM shape_collections
	WHERE team_id = $1
	ORDER BY
		id ASC
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	if s.shapeCollectionStatements.listByUserID, err = s.db.Prepare(`
	SELECT
		shape_collections.id,
		shape_collections.team_id,
		shape_collections.name,
		shape_collections.description
	FROM
		shape_collections,
		team_members
	WHERE
		shape_collections.team_id = team_members.team_id
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

func (s *sqlStore) CreateShapeCollection(shapeCollection *model.ShapeCollection, userID int64) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return -1, errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfTeam(tx, userID, shapeCollection.TeamID)
	if err != nil {
		return -1, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	row := tx.Stmt(s.shapeCollectionStatements.create).QueryRow(
		shapeCollection.TeamID,
		shapeCollection.Name,
		shapeCollection.Description,
	)

	lastInsertID, err := scanIDRow(row)

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.NewStorageErrorFromError(err)
	}

	return lastInsertID, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) GetShapeCollection(shapeCollectionID int64) (*model.ShapeCollection, error) {
	var shapeCollection model.ShapeCollection
	row := s.shapeCollectionStatements.get.QueryRow(
		shapeCollectionID,
	)

	shapeCollection, err := scanShapeCollectionRow(row)

	return &shapeCollection, errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) GetShapeCollectionByUserID(shapeCollectionID int64, userID int64) (*model.ShapeCollection, error) {
	var shapeCollection model.ShapeCollection
	row := s.shapeCollectionStatements.getByUserID.QueryRow(
		shapeCollectionID,
		userID,
	)

	shapeCollection, err := scanShapeCollectionRow(row)
	if err == sql.ErrNoRows {
		return &model.ShapeCollection{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	return &shapeCollection, errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) UpdateShapeCollection(shapeCollection *model.ShapeCollection, userID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfShapeCollection(tx, userID, shapeCollection.ID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	err = s.ensureAdminOfTeam(tx, userID, shapeCollection.TeamID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.shapeCollectionStatements.update).Exec(
		shapeCollection.TeamID,
		shapeCollection.Name,
		shapeCollection.Description,
		shapeCollection.ID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) DeleteShapeCollection(shapeCollectionID int64, userID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfShapeCollection(tx, userID, shapeCollectionID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.shapeCollectionStatements.delete).Exec(
		shapeCollectionID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) ListShapeCollections(offset int64, limit int64) ([]model.ShapeCollection, error) {
	var shapeCollections []model.ShapeCollection
	rows, err := s.shapeCollectionStatements.list.Query(
		limit,
		offset,
	)

	if err != nil {
		return shapeCollections, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		shapeCollection, err := scanShapeCollectionRow(rows)

		if err != nil {
			return shapeCollections, errors.NewStorageErrorFromError(err)
		}

		shapeCollections = append(shapeCollections, shapeCollection)
	}

	return shapeCollections, nil
}

func (s *sqlStore) ListShapeCollectionsByTeamID(teamID int64, offset int64, limit int64) ([]model.ShapeCollection, error) {
	var shapeCollections []model.ShapeCollection
	rows, err := s.shapeCollectionStatements.listByTeamID.Query(
		teamID,
		limit,
		offset,
	)

	if err != nil {
		return shapeCollections, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		shapeCollection, err := scanShapeCollectionRow(rows)

		if err != nil {
			return shapeCollections, errors.NewStorageErrorFromError(err)
		}

		shapeCollections = append(shapeCollections, shapeCollection)
	}

	return shapeCollections, nil
}

func (s *sqlStore) ListShapeCollectionsByUserID(userID int64, offset int64, limit int64) ([]model.ShapeCollection, error) {
	var shapeCollections []model.ShapeCollection
	rows, err := s.shapeCollectionStatements.listByUserID.Query(
		userID,
		limit,
		offset,
	)

	if err != nil {
		return shapeCollections, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		shapeCollection, err := scanShapeCollectionRow(rows)

		if err != nil {
			return shapeCollections, errors.NewStorageErrorFromError(err)
		}

		shapeCollections = append(shapeCollections, shapeCollection)
	}

	return shapeCollections, nil
}

func scanShapeCollectionRow(row rowScanner) (model.ShapeCollection, error) {
	shapeCollection := model.ShapeCollection{}

	err := row.Scan(
		&shapeCollection.ID,
		&shapeCollection.TeamID,
		&shapeCollection.Name,
		&shapeCollection.Description,
	)

	return shapeCollection, err
}
