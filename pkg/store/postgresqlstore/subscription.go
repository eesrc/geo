package postgresqlstore

import (
	"database/sql"
	"fmt"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
	"github.com/eesrc/geo/pkg/sub"
)

type subscriptionStatements struct {
	create                  *sql.Stmt
	get                     *sql.Stmt
	getByUserID             *sql.Stmt
	update                  *sql.Stmt
	delete                  *sql.Stmt
	list                    *sql.Stmt
	listByEntityID          *sql.Stmt
	listByShapeCollectionID *sql.Stmt
	listByUserID            *sql.Stmt
}

func (s *sqlStore) initSubscriptionStatements() error {
	var err error

	if s.subscriptionStatements.create, err = s.db.Prepare(`
	INSERT INTO subscriptions
	(
		name,
		team_id,
		description,
		active,
		output,
		output_config,
		types,
		confidences,
		shape_collection_id,
		trackable_type,
		trackable_id
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11
	) RETURNING id
	`); err != nil {
		return err
	}

	if s.subscriptionStatements.get, err = s.db.Prepare(`
	SELECT
		id,
		team_id,
		name,
		description,
		active,
		output,
		output_config,
		types,
		confidences,
		shape_collection_id,
		trackable_type,
		trackable_id
	FROM subscriptions
	WHERE id = $1
	`); err != nil {
		return err
	}

	if s.subscriptionStatements.getByUserID, err = s.db.Prepare(`
	SELECT
		subscriptions.id,
		subscriptions.team_id,
		subscriptions.name,
		subscriptions.description,
		subscriptions.active,
		subscriptions.output,
		subscriptions.output_config,
		subscriptions.types,
		subscriptions.confidences,
		subscriptions.shape_collection_id,
		subscriptions.trackable_type,
		subscriptions.trackable_id
	FROM
		subscriptions,
		team_members
	WHERE
		subscriptions.team_id = team_members.team_id
		AND
		subscriptions.id = $1
		AND
		team_members.user_id = $2
	`); err != nil {
		return err
	}

	if s.subscriptionStatements.update, err = s.db.Prepare(`
	UPDATE subscriptions
	SET
		name = $1,
		team_id = $2,
		description = $3,
		active = $4,
		output = $5,
		output_config = $6,
		types = $7,
		confidences = $8,
		shape_collection_id = $9,
		trackable_type = $10,
		trackable_id = $11
	WHERE id = $12
	`); err != nil {
		return err
	}

	if s.subscriptionStatements.delete, err = s.db.Prepare(`
	DELETE FROM subscriptions
	WHERE id = $1
	`); err != nil {
		return err
	}

	if s.subscriptionStatements.list, err = s.db.Prepare(`
	SELECT
		id,
		team_id,
		name,
		description,
		active,
		output,
		output_config,
		types,
		confidences,
		shape_collection_id,
		trackable_type,
		trackable_id
	FROM
		subscriptions
	ORDER BY
		id ASC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.subscriptionStatements.listByEntityID, err = s.db.Prepare(`
	SELECT
		id,
		team_id,
		name,
		description,
		active,
		output,
		output_config,
		types,
		confidences,
		shape_collection_id,
		trackable_type,
		trackable_id
	FROM
		subscriptions
	WHERE
		trackable_id = $1
		AND
		trackable_type = $2
	ORDER BY
		id ASC
	LIMIT $3
	OFFSET $4
	`); err != nil {
		return err
	}

	if s.subscriptionStatements.listByShapeCollectionID, err = s.db.Prepare(`
	SELECT
		id,
		team_id,
		name,
		description,
		active,
		output,
		output_config,
		types,
		confidences,
		shape_collection_id,
		trackable_type,
		trackable_id
	FROM
		subscriptions
	WHERE
		shape_collection_id = $1
	ORDER BY
		id ASC
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	if s.subscriptionStatements.listByUserID, err = s.db.Prepare(`
	SELECT
		subscriptions.id,
		subscriptions.team_id,
		subscriptions.name,
		subscriptions.description,
		subscriptions.active,
		subscriptions.output,
		subscriptions.output_config,
		subscriptions.types,
		subscriptions.confidences,
		subscriptions.shape_collection_id,
		subscriptions.trackable_type,
		subscriptions.trackable_id
	FROM
		subscriptions,
		team_members
	WHERE
		team_members.team_id = subscriptions.team_id
		AND
		team_members.user_id = $1
	ORDER BY
		subscriptions.id ASC
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	return err
}

func (s *sqlStore) CreateSubscription(subscription *model.Subscription, userID int64) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return -1, errors.NewStorageErrorFromError(err)
	}

	err = ensureAdminOfSubscriptionResources(s, tx, userID, subscription)
	if err != nil {
		return -1, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	row := tx.Stmt(s.subscriptionStatements.create).QueryRow(
		subscription.Name,
		subscription.TeamID,
		subscription.Description,
		subscription.Active,
		subscription.Output,
		subscription.OutputConfig,
		subscription.Types,
		subscription.Confidences,
		subscription.ShapeCollectionID,
		subscription.TrackableType,
		subscription.TrackableID,
	)

	lastInsertID, err := scanIDRow(row)
	if err != nil {
		_ = tx.Rollback()
		return -1, errors.NewStorageErrorFromError(err)
	}

	return lastInsertID, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) GetSubscription(subscriptionID int64) (*model.Subscription, error) {
	row := s.subscriptionStatements.get.QueryRow(
		subscriptionID,
	)

	subscription, err := scanSubscriptionRow(row)

	return &subscription, errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) GetSubscriptionByUserID(subscriptionID int64, userID int64) (*model.Subscription, error) {
	row := s.subscriptionStatements.getByUserID.QueryRow(
		subscriptionID,
		userID,
	)

	subscription, err := scanSubscriptionRow(row)
	if err == sql.ErrNoRows {
		return &model.Subscription{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	return &subscription, errors.NewStorageErrorFromError(err)
}

func (s *sqlStore) UpdateSubscription(subscription *model.Subscription, userID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfSubscription(tx, userID, subscription.ID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	err = ensureAdminOfSubscriptionResources(s, tx, userID, subscription)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.subscriptionStatements.update).Exec(
		subscription.Name,
		subscription.TeamID,
		subscription.Description,
		subscription.Active,
		subscription.Output,
		subscription.OutputConfig,
		subscription.Types,
		subscription.Confidences,
		subscription.ShapeCollectionID,
		subscription.TrackableType,
		subscription.TrackableID,
		subscription.ID,
	)

	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) DeleteSubscription(subscriptionID int64, userID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfSubscription(tx, userID, subscriptionID)
	if err != nil {
		return errors.NewStorageError(errors.AccessDeniedError, err)
	}

	_, err = tx.Stmt(s.subscriptionStatements.delete).Exec(subscriptionID)
	if err != nil {
		_ = tx.Rollback()
		return errors.NewStorageErrorFromError(err)
	}

	return errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) ListSubscriptionsByCollectionID(collectionID int64, userID int64, offset int64, limit int64) ([]model.Subscription, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return []model.Subscription{}, errors.NewStorageErrorFromError(err)
	}

	_, err = s.ensureAdminOfCollection(tx, userID, collectionID)
	if err != nil {
		return []model.Subscription{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	var subscriptions []model.Subscription
	rows, err := tx.Stmt(s.subscriptionStatements.listByEntityID).Query(
		collectionID,
		string(sub.Collection),
		limit,
		offset,
	)

	if err != nil {
		_ = tx.Rollback()
		return subscriptions, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		subscription, err := scanSubscriptionRow(rows)

		if err != nil {
			_ = tx.Rollback()
			return subscriptions, errors.NewStorageErrorFromError(err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) ListSubscriptionsByTrackerID(trackerID int64, userID int64, offset int64, limit int64) ([]model.Subscription, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return []model.Subscription{}, errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfTracker(tx, userID, trackerID)
	if err != nil {
		return []model.Subscription{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	var subscriptions []model.Subscription
	rows, err := tx.Stmt(s.subscriptionStatements.listByEntityID).Query(
		trackerID,
		string(sub.Tracker),
		limit,
		offset,
	)

	if err != nil {
		_ = tx.Rollback()
		return subscriptions, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		subscription, err := scanSubscriptionRow(rows)

		if err != nil {
			_ = tx.Rollback()
			return subscriptions, errors.NewStorageErrorFromError(err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) ListSubscriptionsByShapeCollectionID(shapeCollectionID int64, userID int64, offset int64, limit int64) ([]model.Subscription, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return []model.Subscription{}, errors.NewStorageErrorFromError(err)
	}

	err = s.ensureAdminOfShapeCollection(tx, userID, shapeCollectionID)
	if err != nil {
		return []model.Subscription{}, errors.NewStorageError(errors.AccessDeniedError, err)
	}

	var subscriptions []model.Subscription
	rows, err := tx.Stmt(s.subscriptionStatements.listByShapeCollectionID).Query(
		shapeCollectionID,
		limit,
		offset,
	)

	if err != nil {
		_ = tx.Rollback()
		return subscriptions, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		subscription, err := scanSubscriptionRow(rows)

		if err != nil {
			_ = tx.Rollback()
			return subscriptions, errors.NewStorageErrorFromError(err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, errors.NewStorageErrorFromError(tx.Commit())
}

func (s *sqlStore) ListSubscriptions(offset int64, limit int64) ([]model.Subscription, error) {
	var subscriptions []model.Subscription
	rows, err := s.subscriptionStatements.list.Query(
		limit,
		offset,
	)

	if err != nil {
		return subscriptions, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		subscription, err := scanSubscriptionRow(rows)

		if err != nil {
			return subscriptions, errors.NewStorageErrorFromError(err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

func (s *sqlStore) ListSubscriptionsByUserID(userID int64, offset int64, limit int64) ([]model.Subscription, error) {
	var subscriptions []model.Subscription
	rows, err := s.subscriptionStatements.listByUserID.Query(
		userID,
		limit,
		offset,
	)

	if err != nil {
		return subscriptions, errors.NewStorageErrorFromError(err)
	}

	defer rows.Close()

	for rows.Next() {
		subscription, err := scanSubscriptionRow(rows)

		if err != nil {
			return subscriptions, errors.NewStorageErrorFromError(err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

func scanSubscriptionRow(row rowScanner) (model.Subscription, error) {
	subscription := model.Subscription{}

	err := row.Scan(
		&subscription.ID,
		&subscription.TeamID,
		&subscription.Name,
		&subscription.Description,
		&subscription.Active,
		&subscription.Output,
		&subscription.OutputConfig,
		&subscription.Types,
		&subscription.Confidences,
		&subscription.ShapeCollectionID,
		&subscription.TrackableType,
		&subscription.TrackableID,
	)

	return subscription, err
}

// ensureAdminOfSubscriptionResources checks if the user is admin of both the team and all resources
// connected to the subscription based on the trackableType. The function rollbacks the transaction
// if an error is returned
func ensureAdminOfSubscriptionResources(s *sqlStore, tx *sql.Tx, userID int64, subscription *model.Subscription) error {
	err := s.ensureAdminOfTeam(tx, userID, subscription.TeamID)
	if err != nil {
		return err
	}

	err = s.ensureAdminOfShapeCollection(tx, userID, subscription.ShapeCollectionID)
	if err != nil {
		return err
	}

	switch subscription.TrackableType {
	case "tracker":
		err = s.ensureAdminOfTracker(tx, userID, subscription.TrackableID)
		if err != nil {
			return err
		}
	case "collection":
		_, err = s.ensureAdminOfCollection(tx, userID, subscription.TrackableID)
		if err != nil {
			return err
		}
	default:
		_ = tx.Rollback()
		return fmt.Errorf("Trackable type %s is not supported", subscription.TrackableType)
	}

	return nil
}
