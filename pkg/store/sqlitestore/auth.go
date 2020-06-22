package sqlitestore

import (
	"database/sql"
	"errors"

	log "github.com/sirupsen/logrus"
)

type authStatements struct {
	adminOfTeam            *sql.Stmt
	adminOfCollection      *sql.Stmt
	adminOfTracker         *sql.Stmt
	adminOfSubscription    *sql.Stmt
	adminOfShapeCollection *sql.Stmt
	adminOfPosition        *sql.Stmt
}

func (s *sqliteStore) initAuthStatements() error {
	var err error
	if s.authStatements.adminOfTeam, err = s.db.Prepare(`
		SELECT team_members.admin
		FROM team_members
		WHERE
			team_members.team_id = $1
			AND
			team_members.user_id = $2
		`); err != nil {
		return err
	}
	if s.authStatements.adminOfCollection, err = s.db.Prepare(`
		SELECT
			collections.team_id,
			team_members.admin
		FROM
			collections,
			team_members
		WHERE
			collections.team_id = team_members.team_id
			AND
			collections.id = $1
			AND
			team_members.user_id = $2
		`); err != nil {
		return err
	}
	if s.authStatements.adminOfTracker, err = s.db.Prepare(`
		SELECT
			team_members.admin
		FROM
			trackers,
			collections,
			team_members
		WHERE
			trackers.collection_id = collections.id
			AND
			collections.team_id = team_members.team_id
			AND
			trackers.id = $1
			AND
			team_members.user_id = $2
		`); err != nil {
		return err
	}
	if s.authStatements.adminOfPosition, err = s.db.Prepare(`
		SELECT
			team_members.admin
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
			positions.id = $1
			AND
			team_members.user_id = $2
		`); err != nil {
		return err
	}
	if s.authStatements.adminOfSubscription, err = s.db.Prepare(`
		SELECT team_members.admin
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
	if s.authStatements.adminOfShapeCollection, err = s.db.Prepare(`
		SELECT team_members.admin
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

	return nil
}

// ensureAdminOfTeam checks if the user is an admin of a team. Needs transaction
// object. Returns an error if the query fails or if the user is not an admin.
func (s *sqliteStore) ensureAdminOfTeam(tx *sql.Tx, userID int64, teamID int64) error {
	var isAdmin bool
	if err := tx.Stmt(s.authStatements.adminOfTeam).QueryRow(teamID, userID).Scan(&isAdmin); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Errorf("Failed to rollback when ensureAdminOfTeam, %v", txErr)
		}

		return err
	}

	if !isAdmin {
		if err := tx.Rollback(); err != nil {
			log.Errorf("Failed to rollback when ensureAdminOfTeam, %v", err)
		}
		return errors.New("User not admin of team")
	}

	return nil
}

// ensureAdminOfCollection checks if the user is an admin of the team that owns the collection.
// Needs transaction object. Returns an error if the query fails or if the user is not an admin.
func (s *sqliteStore) ensureAdminOfCollection(tx *sql.Tx, userID int64, collectionID int64) (int64, error) {
	var teamID int64
	var isAdmin bool

	if err := tx.Stmt(s.authStatements.adminOfCollection).QueryRow(collectionID, userID).Scan(&teamID, &isAdmin); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Errorf("Failed to rollback when ensureAdminOfCollection, %v", txErr)
		}
		return -1, err
	}

	if !isAdmin {
		if err := tx.Rollback(); err != nil {
			log.Errorf("Failed to rollback when ensureAdminOfCollection, %v", err)
		}
		return -1, errors.New("User not admin of collection")
	}
	return teamID, nil
}

// ensureAdminOfTracker checks if the user is an admin of the team that owns the collection that contains the tracker.
// Needs transaction object. Returns an error if the query fails or if the user is not an admin.
func (s *sqliteStore) ensureAdminOfTracker(tx *sql.Tx, userID int64, trackerID int64) error {
	var isAdmin bool

	if err := tx.Stmt(s.authStatements.adminOfTracker).QueryRow(trackerID, userID).Scan(&isAdmin); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Errorf("Failed to rollback when ensureAdminOfTracker, %v", txErr)
		}
		return err
	}

	if !isAdmin {
		if err := tx.Rollback(); err != nil {
			log.Errorf("Failed to rollback when ensureAdminOfTracker, %v", err)
		}
		return errors.New("User not admin of tracker")
	}

	return nil
}

// ensureAdminOfSubscription checks if the user is an admin of the team that owns the subscription.
// Needs transaction object. Returns an error if the query fails or if the user is not an admin.
func (s *sqliteStore) ensureAdminOfSubscription(tx *sql.Tx, userID int64, subscriptionID int64) error {
	var isAdmin bool

	if err := tx.Stmt(s.authStatements.adminOfSubscription).QueryRow(subscriptionID, userID).Scan(&isAdmin); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Errorf("Failed to rollback when ensureAdminOfSubscription, %v", txErr)
		}
		return err
	}

	if !isAdmin {
		if err := tx.Rollback(); err != nil {
			log.Errorf("Failed to rollback when ensureAdminOfSubscription, %v", err)
		}
		return errors.New("User not admin of subscription")
	}

	return nil
}

// ensureAdminOfShapeCollection checks if the user is an admin of the team that owns the subscription.
// Needs transaction object. Returns an error if the query fails or if the user is not an admin.
func (s *sqliteStore) ensureAdminOfShapeCollection(tx *sql.Tx, userID int64, shapeCollectionID int64) error {
	var isAdmin bool

	if err := tx.Stmt(s.authStatements.adminOfShapeCollection).QueryRow(shapeCollectionID, userID).Scan(&isAdmin); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Errorf("Failed to rollback when ensureAdminOfShapeCollection, %v", txErr)
		}
		return err
	}

	if !isAdmin {
		if err := tx.Rollback(); err != nil {
			log.Errorf("Failed to rollback when ensureAdminOfShapeCollection, %v", err)
		}
		return errors.New("User not admin of shape collection")
	}

	return nil
}

// ensureAdminOfPosition checks if the user is an admin of the team that owns the subscription.
// Needs transaction object. Returns an error if the query fails or if the user is not an admin.
func (s *sqliteStore) ensureAdminOfPosition(tx *sql.Tx, userID int64, shapeCollectionID int64) error {
	var isAdmin bool

	if err := tx.Stmt(s.authStatements.adminOfPosition).QueryRow(shapeCollectionID, userID).Scan(&isAdmin); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			log.Errorf("Failed to rollback when ensureAdminOfPosition, %v", txErr)
		}
		return err
	}

	if !isAdmin {
		if err := tx.Rollback(); err != nil {
			log.Errorf("Failed to rollback when ensureAdminOfPosition, %v", err)
		}
		return errors.New("User not admin of position")
	}

	return nil
}
