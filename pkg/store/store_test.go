package store

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/store/errors"
	"github.com/eesrc/geo/pkg/store/sqlitestore"
	"github.com/eesrc/geo/pkg/tria/geometry"

	"github.com/stretchr/testify/assert"
)

var (
	testTeam = model.Team{
		Name:        "The A-Team",
		Description: "I pity the fool",
	}
)

func getTestDB() Store {
	db, err := sqlitestore.New(":memory:", true)

	// Outcomment to test towards a local postgresql server
	// db, err := postgresqlstore.New("user=postgres password=test dbname=geo", true)

	if err != nil {
		log.Fatal("Couldn't initialize DB", err)
	}

	return db
}

var testID int64 = 0

func generateTestUser() *model.User {
	testID += 1

	return &model.User{
		GithubID: strconv.FormatInt(testID, 10) + strconv.FormatInt(time.Now().UnixNano(), 10),
	}
}

func generateTestToken() *model.Token {
	testID += 1

	return &model.Token{
		Token:     strconv.FormatInt(testID, 10) + strconv.FormatInt(time.Now().UnixNano(), 10),
		Resource:  "/foo/bar",
		UserID:    0,
		PermWrite: false,
		Created:   time.Now(),
	}
}

func isStorageError(errorType errors.StorageErrorType, err error) bool {
	if storageError, ok := err.(*errors.StorageError); ok {
		return storageError.Type == errorType
	}

	return false
}

func TestUser(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Create a user
	u := generateTestUser()
	id, err := db.CreateUser(u)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, id)

	// Set the id since the CreateUser function does not mutate the
	// user instance it is given.
	u.ID = id

	// Get user
	readUser, err := db.GetUser(id)
	assert.Nil(t, err)
	assert.NotNil(t, readUser)
	assert.Equal(t, u.Name, readUser.Name)
	assert.Equal(t, u.Email, readUser.Email)
	assert.Equal(t, u.Phone, readUser.Phone)
	assert.Equal(t, u.Created.Unix(), readUser.Created.Unix())
	assert.Equal(t, u.GithubID, readUser.GithubID)

	// Update user

	u.Name = "Clown Shoes"
	u.GithubID = "1234"

	// Perform update
	err = db.UpdateUser(u)
	assert.Nil(t, err)

	// Check that the update took
	updatedUser, err := db.GetUser(id)
	assert.Nil(t, err)
	assert.Equal(t, u.Name, updatedUser.Name)
	assert.Equal(t, u.GithubID, updatedUser.GithubID)

	// Delete user
	err = db.DeleteUser(id)
	assert.Nil(t, err)

	// Make sure getting user fails
	_, err = db.GetUser(id)
	assert.NotNil(t, err)

	// Test listing users
	for i := 0; i < 100; i++ {
		u := generateTestUser()
		_, err := db.CreateUser(u)
		assert.Nil(t, err, u)
	}

	users, err := db.ListUsers(0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(users))
}

func TestToken(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Create a user we can assign the token to
	u := generateTestUser()
	id, err := db.CreateUser(u)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, id)

	// Second test user for negative tests
	negativeUser := generateTestUser()
	negativeUserID, err := db.CreateUser(negativeUser)
	assert.Nil(t, err)

	token := generateTestToken()
	token.UserID = id

	// Create
	createdToken, err := db.CreateToken(token)
	assert.Nil(t, err)

	// Read
	readToken, err := db.GetTokenByUserID(createdToken, id)
	assert.Nil(t, err)
	assert.Equal(t, token.Resource, readToken.Resource)
	assert.Equal(t, token.UserID, readToken.UserID)
	assert.Equal(t, token.PermWrite, readToken.PermWrite)
	assert.Equal(t, token.Created.Unix(), readToken.Created.Unix())

	_, err = db.GetTokenByUserID(createdToken, negativeUserID)
	assert.NotNil(t, err, "Should not be able to get a token")
	assert.True(t, isStorageError(errors.AccessDeniedError, err))

	// Update
	updateToken := readToken
	updateToken.Resource = "/new/path"
	err = db.UpdateToken(updateToken)
	assert.Nil(t, err)
	readToken, err = db.GetTokenByUserID(createdToken, id)
	assert.Nil(t, err)
	assert.Equal(t, "/new/path", readToken.Resource)

	negativeUpdateToken := readToken
	negativeUpdateToken.UserID = negativeUserID
	err = db.UpdateToken(negativeUpdateToken)
	assert.NotNil(t, err)
	assert.True(t, isStorageError(errors.AccessDeniedError, err))

	// Deletion
	err = db.DeleteToken(token.Token, negativeUserID)
	assert.NotNil(t, err, "Should not be able to delete token")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should be access denied token")

	err = db.DeleteToken(token.Token, id)
	assert.Nil(t, err)

	_, err = db.GetToken(token.Token)
	assert.NotNil(t, err)

	// Test listing tokens
	for i := 0; i < 100; i++ {
		token := generateTestToken()
		token.UserID = id
		_, err := db.CreateToken(token)
		assert.Nil(t, err)
	}

	tokens, err := db.ListTokensByUserID(id, 0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(tokens))

	tokens, err = db.ListTokensByUserID(negativeUserID, 0, 100)
	assert.Nil(t, err)
	assert.NotEqual(t, 100, len(tokens))

}

func TestTeam(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Create test user
	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	// Second test user for negative tests
	negativeUser := generateTestUser()
	negativeUserID, err := db.CreateUser(negativeUser)
	assert.Nil(t, err)

	// Create
	tea := testTeam
	id, err := db.CreateTeam(&tea)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, id)

	err = db.SetTeamMember(userID, id, true)
	assert.Nil(t, err)

	// Read
	readTeam, err := db.GetTeamByUserID(id, userID)
	assert.Nil(t, err)
	tea.ID = id
	assert.Equal(t, tea, *readTeam)

	_, err = db.GetTeamByUserID(id, negativeUserID)
	assert.NotNil(t, err, "Should not be able to read team")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Update
	tea.Description = "A bloody good time to be had"
	err = db.UpdateTeam(&tea, userID)
	assert.Nil(t, err)

	readUpdatedTeam, err := db.GetTeamByUserID(id, userID)
	assert.Nil(t, err)
	assert.Equal(t, tea, *readUpdatedTeam)

	err = db.UpdateTeam(&tea, negativeUserID)
	assert.NotNil(t, err)
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Delete
	err = db.DeleteTeam(id, negativeUserID)
	assert.NotNil(t, err)
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	err = db.DeleteTeam(id, userID)
	assert.Nil(t, err)

	_, err = db.GetTeamByUserID(id, userID)
	assert.NotNil(t, err)

	// Test listing teams
	for i := 0; i < 100; i++ {
		teamID, err := db.CreateTeam(&model.Team{})
		assert.Nil(t, err)

		err = db.SetTeamMember(userID, teamID, true)
		assert.Nil(t, err)
	}

	teams, err := db.ListTeamsByUserID(userID, 0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(teams))

	teams, err = db.ListTeamsByUserID(negativeUserID, 0, 100)
	assert.Nil(t, err)
	assert.NotEqual(t, 100, len(teams))
}

func TestTeamMember(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	tea := testTeam
	teamID, _ := db.CreateTeam(&tea)

	err = db.SetTeamMember(userID, teamID, true)
	assert.Nil(t, err)

	err = db.RemoveTeamMember(userID, teamID)
	assert.Nil(t, err)
}

func TestCollection(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Add team
	tea := testTeam
	teamID, _ := db.CreateTeam(&tea)

	// Create test user
	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	err = db.SetTeamMember(userID, teamID, true)
	assert.Nil(t, err)

	// Second test user for negative tests
	negativeUser := generateTestUser()
	negativeUserID, err := db.CreateUser(negativeUser)
	assert.Nil(t, err)

	// Create
	collection := model.Collection{
		TeamID:      teamID,
		Name:        "Trackers United",
		Description: "Just a bunch of sheep in a field",
	}

	collectionID, err := db.CreateCollection(&collection, userID)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, collectionID)

	collection.ID = collectionID

	// Read
	readColl, err := db.GetCollectionByUserID(collectionID, userID)
	assert.Nil(t, err)
	assert.NotNil(t, readColl)
	assert.Equal(t, collection, *readColl)

	_, err = db.GetCollectionByUserID(collectionID, negativeUserID)
	assert.NotNil(t, err)
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Update
	readColl.Name = "Renamed collection"
	err = db.UpdateCollection(readColl, userID)
	assert.Nil(t, err)

	readUpdatedColl, err := db.GetCollectionByUserID(collectionID, userID)
	assert.Nil(t, err)
	assert.Equal(t, readColl, readUpdatedColl)

	err = db.UpdateCollection(readColl, negativeUserID)
	assert.NotNil(t, err)
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Delete
	err = db.DeleteCollection(collectionID, negativeUserID)
	assert.NotNil(t, err)
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	err = db.DeleteCollection(collectionID, userID)
	assert.Nil(t, err)

	_, err = db.GetCollectionByUserID(collectionID, userID)
	assert.NotNil(t, err)
}

func TestTracker(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	team := &model.Team{
		Name:        "my team",
		Description: "some description",
	}
	teamID, err := db.CreateTeam(team)
	assert.Nil(t, err)

	// Create test user
	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	err = db.SetTeamMember(userID, teamID, true)
	assert.Nil(t, err)

	// Second test user for negative tests
	negativeUser := generateTestUser()
	negativeUserID, err := db.CreateUser(negativeUser)
	assert.Nil(t, err)

	collection := &model.Collection{
		TeamID:      teamID,
		Name:        "collection name",
		Description: "collection description",
	}
	collectionID, err := db.CreateCollection(collection, userID)
	assert.Nil(t, err)

	tracker := &model.Tracker{
		CollectionID: collectionID,
		Name:         "Some tracker",
		Description:  "Some description",
	}

	// Create
	_, err = db.CreateTracker(tracker, negativeUserID)
	assert.NotNil(t, err, "Should not allow creating tracker")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	trackerID, err := db.CreateTracker(tracker, userID)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, trackerID)

	// Get
	readTracker, err := db.GetTrackerByUserID(trackerID, userID)
	assert.Nil(t, err)
	assert.NotNil(t, readTracker)
	assert.Equal(t, tracker.Name, readTracker.Name)

	_, err = db.GetTrackerByUserID(trackerID, negativeUserID)
	assert.NotNil(t, err, "Should not allow getting tracker")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Update
	readTracker.Name = "modified name"
	err = db.UpdateTracker(readTracker, userID)
	assert.Nil(t, err)

	updatedTracker, err := db.GetTrackerByUserID(trackerID, userID)
	assert.Nil(t, err)
	assert.Equal(t, readTracker.Name, updatedTracker.Name)

	err = db.UpdateTracker(readTracker, negativeUserID)
	assert.NotNil(t, err, "Should not allow updating of tracker")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Delete
	err = db.DeleteTracker(trackerID, negativeUserID)
	assert.NotNil(t, err, "Should not allow to delete tracker")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	err = db.DeleteTracker(trackerID, userID)
	assert.Nil(t, err)

	_, err = db.GetTrackerByUserID(trackerID, userID)
	assert.NotNil(t, err)

	// Test listing
	for i := 0; i < 100; i++ {
		_, err := db.CreateTracker(&model.Tracker{
			Name:         fmt.Sprintf("Tracker %d", i),
			Description:  fmt.Sprintf("Description of %d", i),
			CollectionID: collectionID,
		}, userID)
		assert.Nil(t, err)
	}
	trackers, err := db.ListTrackersByCollectionID(collectionID, userID, 0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(trackers))

	trackers, err = db.ListTrackersByCollectionID(collectionID, negativeUserID, 0, 100)
	assert.NotNil(t, err)
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")
	assert.NotEqual(t, 100, len(trackers))
}

func TestPosition(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Create tracker
	team := &model.Team{
		Name:        "my team",
		Description: "some description",
	}
	teamID, err := db.CreateTeam(team)
	assert.Nil(t, err)

	// Create test user
	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	err = db.SetTeamMember(userID, teamID, true)
	assert.Nil(t, err)

	// Second test user for negative tests
	negativeUser := generateTestUser()
	negativeUserID, err := db.CreateUser(negativeUser)
	assert.Nil(t, err)

	collection := &model.Collection{
		TeamID:      teamID,
		Name:        "collection name",
		Description: "collection description",
	}
	collectionID, err := db.CreateCollection(collection, userID)
	assert.Nil(t, err)

	tracker := &model.Tracker{
		CollectionID: collectionID,
		Name:         "Some tracker",
		Description:  "Some description",
	}

	trackerID, err := db.CreateTracker(tracker, userID)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, trackerID)

	// Create position
	createPosition := &model.Position{
		TrackerID: trackerID,
		Timestamp: time.Now().UnixNano(),
		Lat:       1.0,
		Lon:       1.0,
		Alt:       1.0,
		Heading:   1.0,
		Speed:     1.0,
		Payload:   []uint8{},
	}
	positionID, err := db.CreatePosition(createPosition, userID)
	createPosition.ID = positionID
	assert.Nil(t, err)
	assert.NotEqual(t, -1, positionID)

	_, err = db.CreatePosition(&model.Position{
		TrackerID: trackerID,
		Timestamp: time.Now().UnixNano(),
		Lat:       1.0,
		Lon:       1.0,
		Alt:       1.0,
		Heading:   1.0,
		Speed:     1.0,
		Payload:   []uint8{},
	}, negativeUserID)
	assert.NotNil(t, err, "Should not be able to create position")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Read position
	readPosition, err := db.GetPositionByUserID(positionID, userID)
	assert.Nil(t, err)
	assert.Equal(t, createPosition, readPosition)

	_, err = db.GetPositionByUserID(positionID, negativeUserID)
	assert.NotNil(t, err, "Should not be able to retrieve position")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Delete position
	err = db.DeletePosition(positionID, negativeUserID)
	assert.NotNil(t, err, "Should not be able to delete position")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	err = db.DeletePosition(positionID, userID)
	assert.Nil(t, err)

	// List position
	for i := 0; i < 100; i++ {
		_, err := db.CreatePosition(&model.Position{
			TrackerID: trackerID,
			Timestamp: time.Now().UnixNano(),
		}, userID)
		assert.Nil(t, err)
	}

	trackers, err := db.ListPositionsByTrackerID(trackerID, userID, 0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(trackers))

	trackers, err = db.ListPositionsByTrackerID(trackerID, negativeUserID, 0, 100)
	assert.NotNil(t, err)
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")
	assert.NotEqual(t, 100, len(trackers))
}

func TestShapeCollections(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Prep data in DB
	team := &model.Team{
		Name:        "my team",
		Description: "some description",
	}
	teamID, err := db.CreateTeam(team)
	assert.Nil(t, err)

	// Create test user
	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	err = db.SetTeamMember(userID, teamID, true)
	assert.Nil(t, err)

	// Second test user for negative tests
	negativeUser := generateTestUser()
	negativeUserID, err := db.CreateUser(negativeUser)
	assert.Nil(t, err)

	// Actual test

	// Create
	shapeCollection := model.ShapeCollection{
		TeamID:      teamID,
		Name:        "ShapeCollection man",
		Description: "Some description",
	}

	_, err = db.CreateShapeCollection(&shapeCollection, negativeUserID)
	assert.NotNil(t, err, "Should not be able to create shape collection")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	shapeCollectionID, err := db.CreateShapeCollection(&shapeCollection, userID)
	assert.Nil(t, err)

	// Set local shapeCollection id to new ID for later comparison
	shapeCollection.ID = shapeCollectionID

	// Retrieve
	retrievedShapeCollection, err := db.GetShapeCollectionByUserID(shapeCollectionID, userID)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, retrievedShapeCollection.ID)
	assert.Equal(t, &shapeCollection, retrievedShapeCollection)

	_, err = db.GetShapeCollectionByUserID(shapeCollectionID, negativeUserID)
	assert.NotNil(t, err, "Should not be able to get shape collection")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Update
	retrievedShapeCollection.Description = "New description"

	err = db.UpdateShapeCollection(retrievedShapeCollection, userID)
	assert.Nil(t, err)

	retrievedShapeCollection, err = db.GetShapeCollectionByUserID(shapeCollectionID, userID)
	assert.Nil(t, err)
	assert.Equal(t, "New description", retrievedShapeCollection.Description)

	err = db.UpdateShapeCollection(retrievedShapeCollection, negativeUserID)
	assert.NotNil(t, err, "Should not be able to update shape collection")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Delete
	err = db.DeleteShapeCollection(shapeCollectionID, negativeUserID)
	assert.NotNil(t, err, "Should not be able to delete shape collection")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	err = db.DeleteShapeCollection(shapeCollectionID, userID)
	assert.Nil(t, err)

	// List
	for i := 0; i < 100; i++ {
		_, err := db.CreateShapeCollection(&shapeCollection, userID)
		assert.Nil(t, err)
	}

	shapes, err := db.ListShapeCollectionsByUserID(userID, 0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(shapes))

	shapes, err = db.ListShapeCollectionsByUserID(negativeUserID, 0, 100)
	assert.Nil(t, err)
	assert.NotEqual(t, 100, len(shapes))
}

func TestShapes(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Prep data in DB
	team := &model.Team{
		Name:        "my team",
		Description: "some description",
	}
	teamID, err := db.CreateTeam(team)
	assert.Nil(t, err)

	// Create test user
	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	err = db.SetTeamMember(userID, teamID, true)
	assert.Nil(t, err)

	// Second test user for negative tests
	negativeUser := generateTestUser()
	negativeUserID, err := db.CreateUser(negativeUser)
	assert.Nil(t, err)

	shapeCollection := model.ShapeCollection{
		TeamID:      teamID,
		Name:        "ShapeCollection man",
		Description: "Some description",
	}
	shapeCollectionID, err := db.CreateShapeCollection(&shapeCollection, userID)
	assert.Nil(t, err)

	// Actual test

	// Create
	shape := model.Shape{
		ShapeCollectionID: shapeCollectionID,
		Name:              "Shapy shapeson",
		Properties: geometry.ShapeProperties{
			"foo": "bar",
		},
		Shape: &geometry.Circle{
			ID:     1,
			Name:   "Circle circleson",
			Origo:  geometry.Point{X: 0, Y: 0},
			Radius: 1337,
			Properties: geometry.ShapeProperties{
				"foo": "bar",
			},
		},
	}

	_, err = db.CreateShape(&shape, negativeUserID)
	assert.NotNil(t, err, "Should not be able to create shape")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	shapeID, err := db.CreateShape(&shape, userID)
	assert.Nil(t, err)

	// Set local shape id to new ID for later comparison
	shape.ID = shapeID
	shape.Shape.SetID(shapeID)

	// Retrieve
	retrievedShape, err := db.GetShapeByUserID(shapeCollectionID, shapeID, userID, true)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, retrievedShape.ID)
	assert.Equal(t, &shape, retrievedShape)

	_, err = db.GetShapeByUserID(shapeCollectionID, shapeID, negativeUserID, true)
	assert.NotNil(t, err, "Should not be able to get shape")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Update
	retrievedShape.Name = "New name"

	err = db.UpdateShape(retrievedShape, userID)
	assert.Nil(t, err)

	retrievedShape, err = db.GetShapeByUserID(shapeCollectionID, shapeID, userID, true)
	assert.Nil(t, err)
	assert.Equal(t, "New name", retrievedShape.Name)

	err = db.UpdateShape(retrievedShape, negativeUserID)
	assert.NotNil(t, err, "Should not be able to update shape")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Delete
	err = db.DeleteShape(shapeCollectionID, shapeID, negativeUserID)
	assert.NotNil(t, err, "Should not be able to delete shape")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	err = db.DeleteShape(shapeCollectionID, shapeID, userID)
	assert.Nil(t, err)

	// List
	for i := 0; i < 100; i++ {
		_, err := db.CreateShape(&shape, userID)
		assert.Nil(t, err)
	}

	shapes, err := db.ListShapesByShapeCollectionIDAndUserID(shapeCollectionID, userID, true, 0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(shapes))

	shapes, err = db.ListShapesByShapeCollectionIDAndUserID(shapeCollectionID, negativeUserID, true, 0, 100)
	assert.NotNil(t, err)
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")
	assert.NotEqual(t, 100, len(shapes))
}

func TestSubscriptions(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Prep data in DB
	team := &model.Team{
		Name:        "my team",
		Description: "some description",
	}
	teamID, err := db.CreateTeam(team)
	assert.Nil(t, err)

	// Create test user
	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	err = db.SetTeamMember(userID, teamID, true)
	assert.Nil(t, err)

	// Second test user for negative tests
	negativeUser := generateTestUser()
	negativeUserID, err := db.CreateUser(negativeUser)
	assert.Nil(t, err)

	collection := &model.Collection{
		TeamID:      teamID,
		Name:        "collection name",
		Description: "collection description",
	}
	collectionID, err := db.CreateCollection(collection, userID)
	assert.Nil(t, err)

	tracker := &model.Tracker{
		CollectionID: collectionID,
		Name:         "Tracker",
	}
	trackerID, err := db.CreateTracker(tracker, userID)
	assert.Nil(t, err)

	shapeCollection := model.ShapeCollection{
		TeamID:      teamID,
		Name:        "ShapeCollection man",
		Description: "Some description",
	}
	shapeCollectionID, err := db.CreateShapeCollection(&shapeCollection, userID)
	assert.Nil(t, err)

	// Actual test

	// Create
	subscriptionCollection := model.Subscription{
		TeamID:      teamID,
		Name:        "Susbcriptions awyeh",
		Description: "Some description",
		Active:      true,
		Output:      "webhook",
		OutputConfig: model.OutputConfig{
			"configParam": "bar",
		},
		Types:             model.MovementList{"inside", "outside"},
		Confidences:       model.ConfidenceList{"high", "medium"},
		ShapeCollectionID: shapeCollectionID,
		TrackableType:     "collection",
		TrackableID:       collectionID,
	}
	subscriptionCollectionID, err := db.CreateSubscription(&subscriptionCollection, userID)
	assert.Nil(t, err)

	_, err = db.CreateSubscription(&subscriptionCollection, negativeUserID)
	assert.NotNil(t, err, "Should not be able to create subscription")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	subscriptionTracker := model.Subscription{
		TeamID:      teamID,
		Name:        "Susbcriptions awyeh",
		Description: "Some description",
		Active:      true,
		Output:      "webhook",
		OutputConfig: model.OutputConfig{
			"configParam": "bar",
		},
		Types:             model.MovementList{"inside", "outside"},
		Confidences:       model.ConfidenceList{"high", "medium"},
		ShapeCollectionID: shapeCollectionID,
		TrackableType:     "tracker",
		TrackableID:       trackerID,
	}
	subscriptionTrackerID, err := db.CreateSubscription(&subscriptionTracker, userID)
	assert.Nil(t, err)

	_, err = db.CreateSubscription(&subscriptionTracker, negativeUserID)
	assert.NotNil(t, err, "Should not be able to create subscription")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Set local shape id to new ID for later comparison
	subscriptionCollection.ID = subscriptionCollectionID
	subscriptionTracker.ID = subscriptionTrackerID

	// Retrieve
	retrievedSubscriptionCollection, err := db.GetSubscriptionByUserID(subscriptionCollectionID, userID)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, retrievedSubscriptionCollection.ID)
	assert.Equal(t, &subscriptionCollection, retrievedSubscriptionCollection)

	_, err = db.GetSubscriptionByUserID(subscriptionCollectionID, negativeUserID)
	assert.NotNil(t, err, "Should not be able to get subscription")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	retrievedSubscriptionTracker, err := db.GetSubscriptionByUserID(subscriptionTrackerID, userID)
	assert.Nil(t, err)
	assert.NotEqual(t, -1, retrievedSubscriptionTracker.ID)
	assert.Equal(t, &subscriptionTracker, retrievedSubscriptionTracker)

	_, err = db.GetSubscriptionByUserID(subscriptionTrackerID, negativeUserID)
	assert.NotNil(t, err, "Should not be able to get subscription")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	// Delete
	err = db.DeleteSubscription(subscriptionCollectionID, negativeUserID)
	assert.NotNil(t, err, "Should not be able to delete subscription")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	err = db.DeleteSubscription(subscriptionCollectionID, userID)
	assert.Nil(t, err)

	err = db.DeleteSubscription(subscriptionTrackerID, negativeUserID)
	assert.NotNil(t, err, "Should not be able to delete subscription")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	err = db.DeleteSubscription(subscriptionTrackerID, userID)
	assert.Nil(t, err)

	// List
	for i := 0; i < 100; i++ {
		_, err := db.CreateSubscription(&subscriptionCollection, userID)
		assert.Nil(t, err)
		_, err = db.CreateSubscription(&subscriptionTracker, userID)
		assert.Nil(t, err)
	}

	subscriptions, err := db.ListSubscriptionsByCollectionID(collectionID, userID, 0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(subscriptions))

	_, err = db.ListSubscriptionsByCollectionID(collectionID, negativeUserID, 0, 100)
	assert.NotNil(t, err, "Should not be allowed to list by collection ID")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	subscriptionsByShapeCollection, err := db.ListSubscriptionsByShapeCollectionID(shapeCollectionID, userID, 0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(subscriptionsByShapeCollection))

	_, err = db.ListSubscriptionsByShapeCollectionID(shapeCollectionID, negativeUserID, 0, 100)
	assert.NotNil(t, err, "Should not be allowed to list by shape collection ID")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")

	shapesTracker, err := db.ListSubscriptionsByTrackerID(trackerID, userID, 0, 100)
	assert.Nil(t, err)
	assert.Equal(t, 100, len(shapesTracker))

	_, err = db.ListSubscriptionsByTrackerID(trackerID, negativeUserID, 0, 100)
	assert.NotNil(t, err, "Should not be allowed to list by tracker ID")
	assert.True(t, isStorageError(errors.AccessDeniedError, err), "Should return an access denied error")
}

func TestGeoSubscriptions(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Prep data in DB
	team := &model.Team{
		Name:        "my team",
		Description: "some description",
	}
	teamID, err := db.CreateTeam(team)
	assert.Nil(t, err)

	// Create test user
	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	err = db.SetTeamMember(userID, teamID, true)
	assert.Nil(t, err)

	collection := &model.Collection{
		TeamID:      teamID,
		Name:        "collection name",
		Description: "collection description",
	}
	collectionID, err := db.CreateCollection(collection, userID)
	assert.Nil(t, err)
	collection.ID = collectionID

	shapeCollection := model.ShapeCollection{
		TeamID:      teamID,
		Name:        "ShapeCollection man",
		Description: "Some description",
	}
	shapeCollectionID, err := db.CreateShapeCollection(&shapeCollection, userID)
	assert.Nil(t, err)
	shapeCollection.ID = shapeCollectionID

	subscription := model.Subscription{
		TeamID:      teamID,
		Name:        "Susbcriptions awyeh",
		Description: "Some description",
		Active:      true,
		Output:      "webhook",
		OutputConfig: model.OutputConfig{
			"configParam": "bar",
		},
		Types:             model.MovementList{"inside", "outside"},
		Confidences:       model.ConfidenceList{"high", "medium"},
		ShapeCollectionID: shapeCollectionID,
		TrackableType:     "collection",
		TrackableID:       collectionID,
	}
	subscriptionID, err := db.CreateSubscription(&subscription, userID)
	assert.Nil(t, err)
	subscription.ID = subscriptionID

	// Actual test

	// List
	geoSubscriptions, err := db.ListGeoSubscriptions(0, 1)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(geoSubscriptions))
	assert.Equal(t, subscription, geoSubscriptions[0].Subscription)
	assert.Equal(t, shapeCollection, geoSubscriptions[0].ShapeCollection)
}

func TestMovement(t *testing.T) {
	db := getTestDB()
	defer db.Close()

	// Prep data in DB
	team := &model.Team{
		Name:        "my team",
		Description: "some description",
	}
	teamID, err := db.CreateTeam(team)
	assert.Nil(t, err)

	// Create test user
	u := generateTestUser()
	userID, err := db.CreateUser(u)
	assert.Nil(t, err)

	err = db.SetTeamMember(userID, teamID, true)
	assert.Nil(t, err)

	collection := &model.Collection{
		TeamID:      teamID,
		Name:        "collection name",
		Description: "collection description",
	}
	collectionID, err := db.CreateCollection(collection, userID)
	assert.Nil(t, err)

	tracker := &model.Tracker{
		CollectionID: collectionID,
		Name:         "Tracker",
	}
	trackerID, err := db.CreateTracker(tracker, userID)
	assert.Nil(t, err)

	shapeCollection := model.ShapeCollection{
		TeamID:      teamID,
		Name:        "ShapeCollection man",
		Description: "Some description",
	}
	shapeCollectionID, err := db.CreateShapeCollection(&shapeCollection, userID)
	assert.Nil(t, err)
	shapeCollection.ID = shapeCollectionID

	subscription := model.Subscription{
		TeamID:      teamID,
		Name:        "Susbcriptions awyeh",
		Description: "Some description",
		Active:      true,
		Output:      "webhook",
		OutputConfig: model.OutputConfig{
			"configParam": "bar",
		},
		Types:             model.MovementList{"inside", "outside"},
		Confidences:       model.ConfidenceList{"high", "medium"},
		ShapeCollectionID: shapeCollectionID,
		TrackableType:     "collection",
		TrackableID:       collectionID,
	}
	subscriptionID, err := db.CreateSubscription(&subscription, userID)
	assert.Nil(t, err)

	shape := model.Shape{
		ShapeCollectionID: shapeCollectionID,
		Name:              "Shapy shapeson",
		Properties: geometry.ShapeProperties{
			"foo": "bar",
		},
		Shape: &geometry.Circle{
			ID:     1,
			Name:   "Circle circleson",
			Origo:  geometry.Point{X: 0, Y: 0},
			Radius: 1337,
			Properties: geometry.ShapeProperties{
				"foo": "bar",
			},
		},
	}

	shapeID, err := db.CreateShape(&shape, userID)
	assert.Nil(t, err)

	positionID, err := db.CreatePosition(&model.Position{
		TrackerID: trackerID,
		Timestamp: time.Now().UnixNano(),
		Lat:       1.0,
		Lon:       1.0,
		Alt:       1.0,
		Heading:   1.0,
		Speed:     1.0,
	}, userID)
	assert.Nil(t, err)

	// Actual test

	// Insert
	trackerMovement := model.TrackerMovement{
		TrackerID:      trackerID,
		SubscriptionID: subscriptionID,
		ShapeID:        shapeID,
		PositionID:     positionID,
		Movements: model.MovementList{
			"inside",
			"entered",
		},
	}

	err = db.InsertMovement(&trackerMovement)
	assert.Nil(t, err)

	// Insert multiple
	movements := make([]model.TrackerMovement, 0)
	for i := 0; i < 100; i++ {
		positionID, err := db.CreatePosition(&model.Position{
			TrackerID: trackerID,
			Timestamp: time.Now().UnixNano(),
			Lat:       1.0,
			Lon:       1.0,
			Alt:       1.0,
			Heading:   1.0,
			Speed:     1.0,
		}, userID)
		assert.Nil(t, err)

		movements = append(movements, model.TrackerMovement{
			TrackerID:      trackerID,
			SubscriptionID: subscriptionID,
			ShapeID:        shapeID,
			PositionID:     positionID,
			Movements: model.MovementList{
				"inside",
				"entered",
			},
		})
	}

	err = db.InsertMovements(movements)
	assert.Nil(t, err)

	// List by subscription
	lastMovements, err := db.ListMovementsBySubscriptionID(subscriptionID, 0, 1000)
	assert.Nil(t, err)
	assert.Equal(t, 101, len(lastMovements))
}
