// Package store defines the storage interface for the Geo server.
package store

import (
	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/postgresqlstore"
	"github.com/eesrc/geo/pkg/store/sqlitestore"
)

// Store defines the persistence layer API
type Store interface {
	// User
	CreateUser(*model.User) (int64, error)
	GetUser(id int64) (*model.User, error)
	GetUserByGithubID(id string) (*model.User, error)
	GetUserByConnectID(id string) (*model.User, error)
	UpdateUser(*model.User) error
	DeleteUser(id int64) error

	ListUsers(offset int64, limit int64) ([]model.User, error)

	// Token
	CreateToken(*model.Token) (string, error)
	UpdateToken(token *model.Token) error
	GetToken(token string) (*model.Token, error)
	GetTokenByUserID(token string, userID int64) (*model.Token, error)
	DeleteToken(token string, userID int64) error

	ListTokens(offset int64, limit int64) ([]model.Token, error)
	ListTokensByUserID(userId int64, offset int64, limit int64) ([]model.Token, error)

	// Team
	CreateTeam(team *model.Team) (int64, error)
	GetTeam(id int64) (*model.Team, error)
	GetTeamByUserID(teamId int64, userID int64) (*model.Team, error)
	UpdateTeam(team *model.Team, userID int64) error
	DeleteTeam(id int64, userID int64) error

	ListTeams(offset int64, limit int64) ([]model.Team, error)
	ListTeamsByUserID(userID int64, offset int64, limit int64) ([]model.Team, error)

	SetTeamMember(userID int64, teamID int64, admin bool) error
	RemoveTeamMember(user int64, team int64) error

	// Collection
	CreateCollection(collection *model.Collection, userID int64) (int64, error)
	GetCollection(collectionID int64) (*model.Collection, error)
	GetCollectionByUserID(collectionID int64, userID int64) (*model.Collection, error)
	UpdateCollection(collection *model.Collection, userID int64) error
	DeleteCollection(collectionID int64, userID int64) error

	ListCollections(offset int64, limit int64) ([]model.Collection, error)
	ListCollectionsByUserID(userID int64, offset int64, limit int64) ([]model.Collection, error)

	// Tracker
	CreateTracker(tracker *model.Tracker, userID int64) (int64, error)
	GetTracker(id int64) (*model.Tracker, error)
	GetTrackerByUserID(id int64, userID int64) (*model.Tracker, error)
	UpdateTracker(tracker *model.Tracker, userID int64) error
	DeleteTracker(id int64, userID int64) error

	ListTrackers(offset int64, limit int64) ([]model.Tracker, error)
	ListTrackersByCollectionID(collectionID int64, userID int64, offset int64, limit int64) ([]model.Tracker, error)

	// ShapeCollection
	CreateShapeCollection(shapeCollection *model.ShapeCollection, userID int64) (int64, error)
	GetShapeCollection(shapeCollectionID int64) (*model.ShapeCollection, error)
	GetShapeCollectionByUserID(shapecollectionID int64, userID int64) (*model.ShapeCollection, error)
	UpdateShapeCollection(shapeCollection *model.ShapeCollection, userID int64) error
	DeleteShapeCollection(shapeCollectionID int64, userID int64) error

	ListShapeCollections(offset int64, limit int64) ([]model.ShapeCollection, error)
	ListShapeCollectionsByTeamID(teamID int64, offset int64, limit int64) ([]model.ShapeCollection, error)
	ListShapeCollectionsByUserID(userID int64, offset int64, limit int64) ([]model.ShapeCollection, error)

	// Shape
	GetShape(shapeCollectionID int64, shapeID int64, includeGeoJSON bool) (*model.Shape, error)
	GetShapeByUserID(shapeCollectionID int64, shapeID int64, userID int64, includeGeoJSON bool) (*model.Shape, error)
	CreateShape(shape *model.Shape, userID int64) (int64, error)
	CreateShapes(shapes []*model.Shape, userID int64) error
	UpdateShape(shape *model.Shape, userID int64) error
	DeleteShape(shapeCollectionID, shapeID int64, userID int64) error

	ListShapes(includeGeoJSON bool, offset, limit int64) ([]model.Shape, error)
	ListShapesByShapeCollectionID(shapeCollectionID int64, includeGeoJSON bool, offset int64, limit int64) ([]model.Shape, error)
	ListShapesByShapeCollectionIDAndUserID(shapeCollectionID int64, userID int64, includeGeoJSON bool, offset int64, limit int64) ([]model.Shape, error)

	ReplaceShapesInShapeCollection(shapeCollectionID int64, userID int64, shapes []*model.Shape) error

	// Position
	CreatePosition(position *model.Position, userID int64) (int64, error)
	GetPosition(id int64) (*model.Position, error)
	GetPositionByUserID(id int64, userID int64) (*model.Position, error)
	DeletePosition(positionID int64, userID int64) error

	ListPositions(offset int64, limit int64) ([]model.Position, error)
	ListPositionsByTrackerID(trackerID int64, userID int64, offset int64, limit int64) ([]model.Position, error)

	// Position movement
	InsertMovement(*model.TrackerMovement) error
	InsertMovements([]model.TrackerMovement) error

	ListMovementsBySubscriptionID(subscriptionID int64, offset int64, limit int64) ([]model.TrackerMovement, error)

	// Subscription
	CreateSubscription(subscription *model.Subscription, userID int64) (int64, error)
	GetSubscription(subscriptionID int64) (*model.Subscription, error)
	GetSubscriptionByUserID(subscriptionID int64, userID int64) (*model.Subscription, error)
	UpdateSubscription(subscription *model.Subscription, userID int64) error
	DeleteSubscription(subscriptionID int64, userID int64) error

	// Listing subscriptions
	ListSubscriptions(offset int64, limit int64) ([]model.Subscription, error)
	ListSubscriptionsByShapeCollectionID(shapeCollectionID int64, userID int64, offset int64, limit int64) ([]model.Subscription, error)
	ListSubscriptionsByCollectionID(collectionID int64, userID int64, offset int64, limit int64) ([]model.Subscription, error)
	ListSubscriptionsByTrackerID(trackerID int64, userID int64, offset int64, limit int64) ([]model.Subscription, error)
	ListSubscriptionsByUserID(userID int64, offset int64, limit int64) ([]model.Subscription, error)

	// GeoSubscriptions
	GetGeoSubscriptionBySubscription(subscriptionID int64) (*model.GeoSubscription, error)

	// List GeoSubscriptions
	ListGeoSubscriptions(offset, limit int64) ([]model.GeoSubscription, error)
	ListGeoSubscriptionsByShapeCollectionID(shapeCollectionID int64, offset int64, limit int64) ([]model.GeoSubscription, error)

	Close() error
}

// New creates a new Store backed by given driver.
func New(dbDriver string, connectionString string, create bool) (Store, error) {
	var store Store

	switch dbDriver {
	case "sqlite3":
		sqliteStore, err := sqlitestore.New(connectionString, create)

		if err != nil {
			return nil, err
		}

		store = sqliteStore
	case "postgres":
		postgresStore, err := postgresqlstore.New(connectionString, create)

		if err != nil {
			return nil, err
		}

		store = postgresStore
	default:
		log.Fatalf("Unsupported DB driver %s", dbDriver)
	}

	return store, nil
}
