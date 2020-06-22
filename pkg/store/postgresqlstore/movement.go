package postgresqlstore

import (
	"database/sql"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
)

type movementStatements struct {
	create               *sql.Stmt
	listBySubscriptionID *sql.Stmt
}

func (s *sqlStore) initMovementStatements() error {
	var err error

	if s.movementStatements.create, err = s.db.Prepare(`
	INSERT INTO position_movements (
		tracker_id,
		subscription_id,
		position_id,
		shape_id,
		movement
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5
	)
	`); err != nil {
		return err
	}

	if s.movementStatements.listBySubscriptionID, err = s.db.Prepare(`
	SELECT
		tracker_id,
		subscription_id,
		position_id,
		shape_id,
		movement
	FROM
		position_movements
	WHERE subscription_id = $1
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	return err
}

func (s *sqlStore) InsertMovement(movement *model.TrackerMovement) error {
	_, err := s.movementStatements.create.Exec(
		movement.TrackerID,
		movement.SubscriptionID,
		movement.PositionID,
		movement.ShapeID,
		movement.Movements,
	)

	return errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) InsertMovements(movements []model.TrackerMovement) error {
	tx, err := s.db.Begin()

	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	for _, movement := range movements {
		_, err = tx.Stmt(s.movementStatements.create).Exec(
			movement.TrackerID,
			movement.SubscriptionID,
			movement.PositionID,
			movement.ShapeID,
			movement.Movements,
		)

		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Error("Error on rollback for replace movements", rbErr)
			}
			return errors.NewStorageErrorFromError(err)
		}
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) ListMovementsBySubscriptionID(subscriptionID int64, offset int64, limit int64) ([]model.TrackerMovement, error) {
	var positionMovements []model.TrackerMovement

	rows, err := s.movementStatements.listBySubscriptionID.Query(
		subscriptionID,
		limit,
		offset,
	)

	if err != nil {
		return positionMovements, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		movements, err := scanMovementRow(rows)

		if err != nil {
			return positionMovements, errors.NewStorageErrorFromError(err)
		}

		positionMovements = append(positionMovements, movements)
	}

	return positionMovements, nil
}

func scanMovementRow(row rowScanner) (model.TrackerMovement, error) {
	trackerMovement := model.TrackerMovement{}

	err := row.Scan(
		&trackerMovement.TrackerID,
		&trackerMovement.SubscriptionID,
		&trackerMovement.PositionID,
		&trackerMovement.ShapeID,
		&trackerMovement.Movements,
	)

	return trackerMovement, err
}
