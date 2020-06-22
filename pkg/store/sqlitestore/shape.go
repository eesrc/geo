package sqlitestore

import (
	"database/sql"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
	"github.com/eesrc/geo/pkg/tria/geometry"
)

type shapeStatements struct {
	create                                        *sql.Stmt
	get                                           *sql.Stmt
	getIncludeShape                               *sql.Stmt
	getByUserID                                   *sql.Stmt
	getIncludeShapeByUserID                       *sql.Stmt
	update                                        *sql.Stmt
	delete                                        *sql.Stmt
	deleteByShapeCollectionID                     *sql.Stmt
	list                                          *sql.Stmt
	listIncludeShapes                             *sql.Stmt
	listByShapeCollectionID                       *sql.Stmt
	listIncludeShapesByShapeCollectionID          *sql.Stmt
	listByShapeCollectionIDAndUserID              *sql.Stmt
	listIncludeShapesByShapeCollectionIDAndUserID *sql.Stmt
}

func (s *sqliteStore) initShapeStatements() error {
	var err error

	if s.shapeStatements.create, err = s.db.Prepare(`
	INSERT INTO shapes (
		shape_collection_id,
		name,
		properties,
		shape
	) VALUES (
		$1,
		$2,
		$3,
		$4
	)`); err != nil {
		return err
	}

	if s.shapeStatements.get, err = s.db.Prepare(`
	SELECT
		id,
		shape_collection_id,
		name,
		properties
	FROM
		shapes
	WHERE
		id=$1
	AND
		shape_collection_id = $2
	`); err != nil {
		return err
	}

	if s.shapeStatements.getIncludeShape, err = s.db.Prepare(`
	SELECT
		id,
		shape_collection_id,
		name,
		properties,
		shape
	FROM
		shapes
	WHERE
		id=$1
	AND
		shape_collection_id = $2
	`); err != nil {
		return err
	}

	if s.shapeStatements.getByUserID, err = s.db.Prepare(`
	SELECT
		shapes.id,
		shapes.shape_collection_id,
		shapes.name,
		shapes.properties
	FROM
		shapes,
		shape_collections,
		team_members
	WHERE
		shapes.shape_collection_id = shape_collections.id
		AND
		shape_collections.team_id = team_members.team_id
		AND
		shapes.id=$1
		AND
		shapes.shape_collection_id = $2
		AND
		team_members.user_id = $3
	`); err != nil {
		return err
	}

	if s.shapeStatements.getIncludeShapeByUserID, err = s.db.Prepare(`
	SELECT
		shapes.id,
		shapes.shape_collection_id,
		shapes.name,
		shapes.properties,
		shapes.shape
	FROM
		shapes,
		shape_collections,
		team_members
	WHERE
		shapes.shape_collection_id = shape_collections.id
		AND
		shape_collections.team_id = team_members.team_id
		AND
		shapes.id=$1
		AND
		shapes.shape_collection_id = $2
		AND
		team_members.user_id = $3
	`); err != nil {
		return err
	}

	if s.shapeStatements.update, err = s.db.Prepare(`
	UPDATE shapes
	SET
		shape_collection_id = $1,
		name = $2,
		properties = $3,
		shape = $4
	WHERE
		id = $5
	AND
		shape_collection_id = $6
	`); err != nil {
		return err
	}

	if s.shapeStatements.delete, err = s.db.Prepare(`
	DELETE FROM shapes
	WHERE
		id = $1
		AND
		shape_collection_id = $2
	`); err != nil {
		return err
	}

	if s.shapeStatements.list, err = s.db.Prepare(`
	SELECT
		id,
		shape_collection_id,
		name,
		properties
	FROM
		shapes
	ORDER BY
		id ASC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.shapeStatements.listIncludeShapes, err = s.db.Prepare(`
	SELECT
		id,
		shape_collection_id,
		name,
		properties,
		shape
	FROM
		shapes
	ORDER BY
		id ASC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.shapeStatements.listByShapeCollectionID, err = s.db.Prepare(`
	SELECT
		id,
		shape_collection_id,
		name,
		properties
	FROM
		shapes
	WHERE
		shape_collection_id = $1
	ORDER BY
		id ASC
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	if s.shapeStatements.listIncludeShapesByShapeCollectionID, err = s.db.Prepare(`
	SELECT
		id,
		shape_collection_id,
		name,
		properties,
		shape
	FROM
		shapes
	WHERE
		shape_collection_id = $1
	ORDER BY
		id ASC
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	if s.shapeStatements.listByShapeCollectionIDAndUserID, err = s.db.Prepare(`
	SELECT
		shapes.id,
		shapes.shape_collection_id,
		shapes.name,
		shapes.properties
	FROM
		shapes,
		shape_collections,
		team_members
	WHERE
		shapes.shape_collection_id = shape_collections.id
		AND
		shape_collections.team_id = team_members.team_id
		AND
		shapes.shape_collection_id = $1
		AND
		team_members.user_id = $2
	ORDER BY
		shapes.id ASC
	LIMIT $3
	OFFSET $4
	`); err != nil {
		return err
	}

	if s.shapeStatements.listIncludeShapesByShapeCollectionIDAndUserID, err = s.db.Prepare(`
	SELECT
		shapes.id,
		shapes.shape_collection_id,
		shapes.name,
		shapes.properties,
		shapes.shape
	FROM
		shapes,
		shape_collections,
		team_members
	WHERE
		shapes.shape_collection_id = shape_collections.id
		AND
		shape_collections.team_id = team_members.team_id
		AND
		shapes.shape_collection_id = $1
		AND
		team_members.user_id = $2
	ORDER BY
		shapes.id ASC
	LIMIT $3
	OFFSET $4
	`); err != nil {
		return err
	}

	if s.shapeStatements.deleteByShapeCollectionID, err = s.db.Prepare(`
	DELETE FROM shapes
	WHERE shape_collection_id = $1
	`); err != nil {
		return err
	}

	return err
}

func (s *sqliteStore) CreateShape(shape *model.Shape, userID int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return -1, errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfShapeCollection(tx, userID, shape.ShapeCollectionID)
	if err != nil {
		return -1, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	shapeStorage, err := shapeStoragefromShapeModel(shape)
	if err != nil {
		return -1, errors.NewStorageError(errors.InternalError, err)
	}

	r, err := tx.Stmt(s.shapeStatements.create).Exec(
		shape.ShapeCollectionID,
		shape.Name,
		shape.Properties,
		shapeStorage,
	)

	if err != nil {
		_ = tx.Rollback()
		return -1, errors.NewStorageErrorFromError(err)
	}

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return -1, errors.NewStorageErrorFromError(err)
	}

	return lastInsertID, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) CreateShapes(shapes []*model.Shape, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()

	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	shapeCollectionIDs := make(map[int64]bool)

	for _, shape := range shapes {
		if shapeCollectionIDs[shape.ShapeCollectionID] {
			continue
		}

		err = s.ensureAdminOfShapeCollection(tx, userID, shape.ShapeCollectionID)
		if err != nil {
			return errors.NewStorageError(errors.AccessDeniedError, err)
		}

		shapeCollectionIDs[shape.ShapeCollectionID] = true
	}

	for _, shape := range shapes {
		shapeStorage, err := shapeStoragefromShapeModel(shape)
		if err != nil {
			return errors.NewStorageErrorFromError(err)
		}
		_, err = tx.Stmt(s.shapeStatements.create).Exec(
			shape.ShapeCollectionID,
			shape.Name,
			shape.Properties,
			shapeStorage,
		)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error("Error on rollback for createShape", rbErr)
			}
			return errors.NewStorageErrorFromError(err)
		}
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) GetShape(shapeCollectionID, shapeID int64, includeGeoJSON bool) (*model.Shape, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var shape model.Shape
	var err error

	if includeGeoJSON {
		row := s.shapeStatements.getIncludeShape.QueryRow(
			shapeID,
			shapeCollectionID,
		)
		shape, err = scanShapeRowWithShape(row)
	} else {
		row := s.shapeStatements.get.QueryRow(
			shapeID,
			shapeCollectionID,
		)
		shape, err = scanShapeRow(row)
	}

	return &shape, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) GetShapeByUserID(shapeCollectionID int64, shapeID int64, userID int64, includeGeoJSON bool) (*model.Shape, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var shape model.Shape
	var err error

	if includeGeoJSON {
		row := s.shapeStatements.getIncludeShapeByUserID.QueryRow(
			shapeID,
			shapeCollectionID,
			userID,
		)
		shape, err = scanShapeRowWithShape(row)
	} else {
		row := s.shapeStatements.getByUserID.QueryRow(
			shapeID,
			shapeCollectionID,
			userID,
		)
		shape, err = scanShapeRow(row)
	}

	if err == sql.ErrNoRows {
		return &model.Shape{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	if err != nil {
		log.Error(err)
	}

	return &shape, errors.NewStorageErrorFromError(err)
}

func (s *sqliteStore) UpdateShape(shape *model.Shape, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfShapeCollection(tx, userID, shape.ShapeCollectionID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	shapeStorage, err := shapeStoragefromShapeModel(shape)
	if err != nil {
		return errors.NewStorageError(errors.InternalError, err)
	}

	_, err = tx.Stmt(s.shapeStatements.update).Exec(
		shape.ShapeCollectionID,
		shape.Name,
		shape.Properties,
		shapeStorage,
		shape.ID,
		shape.ShapeCollectionID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) DeleteShape(shapeCollectionID, shapeID int64, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfShapeCollection(tx, userID, shapeCollectionID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.shapeStatements.delete).Exec(
		shapeID,
		shapeCollectionID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) ListShapes(includeGeoJSON bool, offset, limit int64) ([]model.Shape, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var shapes []model.Shape
	var rows *sql.Rows
	var err error
	if includeGeoJSON {
		rows, err = s.shapeStatements.listIncludeShapes.Query(
			limit,
			offset,
		)
	} else {
		rows, err = s.shapeStatements.list.Query(
			limit,
			offset,
		)
	}

	if err != nil {
		return shapes, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		var shape model.Shape
		if includeGeoJSON {
			shape, err = scanShapeRowWithShape(rows)
		} else {
			shape, err = scanShapeRow(rows)
		}

		if err != nil {
			return shapes, errors.NewStorageErrorFromError(err)
		}

		shapes = append(shapes, shape)
	}

	return shapes, nil
}

func (s *sqliteStore) ListShapesByShapeCollectionID(shapeCollectionID int64, includeGeoJSON bool, offset int64, limit int64) ([]model.Shape, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var shapes []model.Shape
	var rows *sql.Rows
	var err error
	if includeGeoJSON {
		rows, err = s.shapeStatements.listIncludeShapesByShapeCollectionID.Query(shapeCollectionID, limit, offset)
	} else {
		rows, err = s.shapeStatements.listByShapeCollectionID.Query(shapeCollectionID, limit, offset)
	}

	if err != nil {
		return shapes, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		var shape model.Shape
		if includeGeoJSON {
			shape, err = scanShapeRowWithShape(rows)
		} else {
			shape, err = scanShapeRow(rows)
		}

		if err != nil {
			return shapes, errors.NewStorageErrorFromError(err)
		}

		shapes = append(shapes, shape)
	}

	return shapes, err
}

func (s *sqliteStore) ListShapesByShapeCollectionIDAndUserID(shapeCollectionID int64, userID int64, includeGeoJSON bool, offset int64, limit int64) ([]model.Shape, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return []model.Shape{}, errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfShapeCollection(tx, userID, shapeCollectionID)
	if err != nil {
		return []model.Shape{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	var shapes []model.Shape
	var rows *sql.Rows
	if includeGeoJSON {
		rows, err = tx.Stmt(s.shapeStatements.listIncludeShapesByShapeCollectionIDAndUserID).Query(shapeCollectionID, userID, limit, offset)
	} else {
		rows, err = tx.Stmt(s.shapeStatements.listByShapeCollectionIDAndUserID).Query(shapeCollectionID, userID, limit, offset)
	}

	if err != nil {
		_ = tx.Rollback()
		return shapes, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		var shape model.Shape
		if includeGeoJSON {
			shape, err = scanShapeRowWithShape(rows)
		} else {
			shape, err = scanShapeRow(rows)
		}

		if err != nil {
			_ = tx.Rollback()
			return shapes, errors.NewStorageErrorFromError(err)
		}

		shapes = append(shapes, shape)
	}

	return shapes, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqliteStore) ReplaceShapesInShapeCollection(shapeCollectionID int64, userID int64, shapes []*model.Shape) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()

	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	shapeCollectionIDs := make(map[int64]bool)

	for _, shape := range shapes {
		if shapeCollectionIDs[shape.ShapeCollectionID] {
			continue
		}

		err = s.ensureAdminOfShapeCollection(tx, userID, shape.ShapeCollectionID)
		if err != nil {
			return errors.NewStorageError(errors.AccessDeniedError, err)
		}

		shapeCollectionIDs[shape.ShapeCollectionID] = true
	}

	// First we drop all the old shapes
	_, err = tx.Stmt(s.shapeStatements.deleteByShapeCollectionID).Exec(shapeCollectionID)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Error("Error on rollback for replaceShapes", rbErr)
		}
		return errors.NewStorageErrorFromError(err)
	}

	// Insert all new ones
	for _, shape := range shapes {
		shapeStorage, err := shapeStoragefromShapeModel(shape)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error("Error on rollback for replaceShapes", rbErr)
			}
			return errors.NewStorageError(errors.InternalError, err)
		}

		_, err = tx.Stmt(s.shapeStatements.create).Exec(
			shape.ShapeCollectionID,
			shape.Name,
			shape.Properties,
			shapeStorage,
		)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error("Error on rollback for replaceShapes", rbErr)
			}
			return errors.NewStorageErrorFromError(err)
		}
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func scanShapeRow(row rowScanner) (model.Shape, error) {
	shape := model.Shape{
		Properties: geometry.ShapeProperties{},
	}

	err := row.Scan(
		&shape.ID,
		&shape.ShapeCollectionID,
		&shape.Name,
		&shape.Properties,
	)

	if err != nil {
		return model.Shape{}, err
	}

	return shape, err
}

func scanShapeRowWithShape(row rowScanner) (model.Shape, error) {
	shape := model.Shape{
		Properties: geometry.ShapeProperties{},
	}
	shapeStorage := shapeStorageModel{}

	err := row.Scan(
		&shape.ID,
		&shape.ShapeCollectionID,
		&shape.Name,
		&shape.Properties,
		&shapeStorage,
	)

	if err != nil {
		return model.Shape{}, err
	}

	shape.Shape, err = shapeModelFromShapeStorage(&shapeStorage)
	if err != nil {
		return model.Shape{}, err
	}

	shape.Shape.SetID(shape.ID)

	return shape, err
}
