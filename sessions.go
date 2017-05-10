package sessions

import (
	"time"

	"github.com/Unaxiom/ulogger"
	"gopkg.in/jackc/pgx.v2"
)

func init() {

}

// Setup accepts a db conn pool, and a ulogger instance and sets up the required variables. It also starts the expiry timers of existing sessions.
func Setup(dbParent *pgx.ConnPool, logParent *ulogger.Logger) {
	db = dbParent
	log = logParent
	// Fetch all the active sessions here

	// Calculate the expiry timers and delete the sessions accordingly
}

// FetchAllSessions returns all the live sessions.
func FetchAllSessions() []Session {
	allSessions := make([]Session, 0)
	return allSessions
}

// NewSession accepts the key to encode, along with the client's IP Address, inserts into the sessions table, sets the delete session timer, and returns a Session struct.
func NewSession(key string, ipAddress string) Session {
	var assignedSession Session

	// Calculate the session value here

	// Insert this into the database here

	// Calculate Session.ExpiresIn from Session.ExpiresAt (after adjusting it to the specified time zone)

	// Set up the delete timer here

	return assignedSession

}

// DeleteSession deletes the session with the specific sessionID and returns an error, if any.
func DeleteSession(sessionID int64) error {
	tx, _ := db.Begin()
	defer tx.Rollback()

	_, err := tx.Exec(`
		DELETE FROM sessions WHERE id = $1
	`, sessionID)
	if err != nil {
		tx.Rollback()
		log.Errorln("While deleting a session, error is ", err)
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		tx.Rollback()
		log.Errorln("While deleting a session, error is ", commitErr)
		return commitErr
	}
	return nil
}

// calculateExpiresIn calculates the time after which a session needs to be deleted, from its full-timezone timestamp.
func calculateExpiresIn(expiresAt time.Time) int64 {
	var expiresIn int64
	return expiresIn
}
