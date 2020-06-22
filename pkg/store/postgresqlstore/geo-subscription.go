package postgresqlstore

import (
	"database/sql"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
)

type geoSubscriptionStatements struct {
	getBySubscription       *sql.Stmt
	list                    *sql.Stmt
	listByShapeCollectionID *sql.Stmt
}

func (s *sqlStore) initGeoSubscriptionStatements() error {
	var err error

	if s.geoSubscriptionStatements.getBySubscription, err = s.db.Prepare(`
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
		subscriptions.trackable_id,

		shape_collections.id,
		shape_collections.team_id,
		shape_collections.name,
		shape_collections.description
	FROM subscriptions
	LEFT JOIN
		shape_collections
	ON
		shape_collections.id=subscriptions.shape_collection_id
	WHERE
		subscriptions.id = $1
	`); err != nil {
		return err
	}

	if s.geoSubscriptionStatements.list, err = s.db.Prepare(`
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
		subscriptions.trackable_id,

		shape_collections.id,
		shape_collections.team_id,
		shape_collections.name,
		shape_collections.description
	FROM subscriptions
	LEFT JOIN
		shape_collections
	ON
		shape_collections.id=subscriptions.shape_collection_id
	ORDER BY
		subscriptions.id DESC
	LIMIT $1
	OFFSET $2
	`); err != nil {
		return err
	}

	if s.geoSubscriptionStatements.listByShapeCollectionID, err = s.db.Prepare(`
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
		subscriptions.trackable_id,

		shape_collections.id,
		shape_collections.team_id,
		shape_collections.name,
		shape_collections.description
	FROM subscriptions
	LEFT JOIN
		shape_collections
	ON
		shape_collections.id=subscriptions.shape_collection_id
	WHERE
	subscriptions.shape_collection_id = $1
	ORDER BY
		subscriptions.id DESC
	LIMIT $2
	OFFSET $3
	`); err != nil {
		return err
	}

	return err
}

func (s *sqlStore) GetGeoSubscriptionBySubscription(subscriptionID int64) (*model.GeoSubscription, error) {
	row := s.geoSubscriptionStatements.getBySubscription.QueryRow(
		subscriptionID,
	)
	geoSubscription, err := scanGeoSubscriptionRow(row)

	if err != nil {
		return &geoSubscription, errors.NewStorageErrorFromError(err)
	}

	movements, err := s.ListMovementsBySubscriptionID(geoSubscription.Subscription.ID, 0, 100)

	if err != nil {
		return &geoSubscription, errors.NewStorageErrorFromError(err)
	}

	shapes, err := s.ListShapesByShapeCollectionID(geoSubscription.Subscription.ShapeCollectionID, true, 0, 1000000)
	if err != nil {
		return &geoSubscription, errors.NewStorageErrorFromError(err)
	}

	geoSubscription.TrackerMovements = append(geoSubscription.TrackerMovements, movements...)
	geoSubscription.Shapes = append(geoSubscription.Shapes, shapes...)

	return &geoSubscription, nil
}

func (s *sqlStore) ListGeoSubscriptions(offset, limit int64) ([]model.GeoSubscription, error) {
	var geoSubscriptions []model.GeoSubscription

	rows, err := s.geoSubscriptionStatements.list.Query(
		limit,
		offset,
	)

	if err != nil {
		return geoSubscriptions, errors.NewStorageErrorFromError(err)
	}

	for rows.Next() {
		geoSubscription, err := scanGeoSubscriptionRow(rows)

		if err != nil {
			return geoSubscriptions, errors.NewStorageErrorFromError(err)
		}

		geoSubscriptions = append(geoSubscriptions, geoSubscription)
	}

	rows.Close()

	for i := range geoSubscriptions {
		movements, err := s.ListMovementsBySubscriptionID(geoSubscriptions[i].Subscription.ID, 0, 100)

		if err != nil {
			return geoSubscriptions, errors.NewStorageErrorFromError(err)
		}

		shapes, err := s.ListShapesByShapeCollectionID(geoSubscriptions[i].Subscription.ShapeCollectionID, true, 0, 1000000)
		if err != nil {
			return geoSubscriptions, errors.NewStorageErrorFromError(err)
		}

		geoSubscriptions[i].TrackerMovements = append(geoSubscriptions[i].TrackerMovements, movements...)
		geoSubscriptions[i].Shapes = append(geoSubscriptions[i].Shapes, shapes...)
	}

	return geoSubscriptions, nil
}

func (s *sqlStore) ListGeoSubscriptionsByShapeCollectionID(shapeCollectionID int64, offset int64, limit int64) ([]model.GeoSubscription, error) {
	var geoSubscriptions []model.GeoSubscription

	rows, err := s.geoSubscriptionStatements.listByShapeCollectionID.Query(
		shapeCollectionID,
		limit,
		offset,
	)

	if err != nil {
		return geoSubscriptions, errors.NewStorageErrorFromError(err)
	}

	for rows.Next() {
		geoSubscription, err := scanGeoSubscriptionRow(rows)

		if err != nil {
			return geoSubscriptions, errors.NewStorageErrorFromError(err)
		}

		geoSubscriptions = append(geoSubscriptions, geoSubscription)
	}

	rows.Close()

	for i := range geoSubscriptions {
		movements, err := s.ListMovementsBySubscriptionID(geoSubscriptions[i].Subscription.ID, 0, 100)

		if err != nil {
			return geoSubscriptions, errors.NewStorageErrorFromError(err)
		}

		shapes, err := s.ListShapesByShapeCollectionID(geoSubscriptions[i].Subscription.ShapeCollectionID, true, 0, 1000000)
		if err != nil {
			return geoSubscriptions, errors.NewStorageErrorFromError(err)
		}

		geoSubscriptions[i].TrackerMovements = append(geoSubscriptions[i].TrackerMovements, movements...)
		geoSubscriptions[i].Shapes = append(geoSubscriptions[i].Shapes, shapes...)
	}

	return geoSubscriptions, nil
}

func scanGeoSubscriptionRow(row rowScanner) (model.GeoSubscription, error) {
	geoSubscription := model.GeoSubscription{}

	err := row.Scan(
		&geoSubscription.Subscription.ID,
		&geoSubscription.Subscription.TeamID,
		&geoSubscription.Subscription.Name,
		&geoSubscription.Subscription.Description,
		&geoSubscription.Subscription.Active,
		&geoSubscription.Subscription.Output,
		&geoSubscription.Subscription.OutputConfig,
		&geoSubscription.Subscription.Types,
		&geoSubscription.Subscription.Confidences,
		&geoSubscription.Subscription.ShapeCollectionID,
		&geoSubscription.Subscription.TrackableType,
		&geoSubscription.Subscription.TrackableID,

		&geoSubscription.ShapeCollection.ID,
		&geoSubscription.ShapeCollection.TeamID,
		&geoSubscription.ShapeCollection.Name,
		&geoSubscription.ShapeCollection.Description,
	)

	return geoSubscription, err
}
