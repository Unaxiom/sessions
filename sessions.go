package sessions

import (
	"time"

	"crypto/sha256"

	"fmt"

	"github.com/Unaxiom/ulogger"
	"github.com/twinj/uuid"
	"gopkg.in/jackc/pgx.v2"
)

func init() {

}

// Setup accepts a db conn pool, a ulogger instance, number of seconds after which the session expires (if set to 0, default value of 86400 is used), the timezone (if set to "", default timezone - UTC is used), and sets up the required variables. It also starts the expiry timers of existing sessions.
func Setup(dbParent *pgx.ConnPool, logParent *ulogger.Logger, sessionExpiresInSecs int64, sessionLocalTimezoneName string) {
	db = dbParent
	log = logParent
	if sessionExpiresInSecs != 0 {
		sessionExpiryTime = sessionExpiresInSecs
	}
	if sessionLocalTimezoneName != "" {
		sessionTimezoneName = sessionLocalTimezoneName
	}

	// Load the time zone here
	var err error
	timezoneLocation, err = time.LoadLocation(sessionTimezoneName)
	if err != nil {
		log.Errorln("Couldn't load timezone location --> ", sessionTimezoneName, ". Error is ", err, ".")
		timezoneLocation, _ = time.LoadLocation("UTC")
	}

	// Fetch all the active sessions here
	allSessions, err := FetchAllSessions()
	if err != nil {
		log.Errorln("While trying to fetch all sessions during Setup, error is ", err)
	}

	// Calculate the expiry timers and delete the sessions accordingly
	for _, session := range allSessions {
		go func() {
			<-time.After(time.Second * time.Duration(session.ExpiryIn))
			go DeleteSession(session.Token)
		}()
	}
}

// FetchAllSessions returns all the live sessions.
func FetchAllSessions() ([]Session, error) {
	allSessions := make([]Session, 0)
	rows, err := db.Query(`
		SELECT id, key, token, expires_at, ip, timestamp FROM sessions WHERE active = True
	`)
	defer rows.Close()
	if err != nil {
		log.Errorln("While trying to fetch all sessions in FetchAllSessions, error is ", err)
		return allSessions, err
	}
	for rows.Next() {
		var session Session
		rows.Scan(&session.ID, &session.Key, &session.Token, &session.ExpiresAt, &session.IP, &session.Timestamp)
		session.ExpiryIn = calculateExpiresIn(session.ExpiresAt)
		allSessions = append(allSessions, session)
	}
	return allSessions, nil
}

// FetchSessionData returns the data of the particular session token
func FetchSessionData(session Session) (Session, error) {
	err := db.QueryRow(`
		SELECT id, key, token, expires_at, ip, active, timestamp FROM sessions WHERE token = $1
	`, session.Token).Scan(&session.ID, &session.Key, &session.Token, &session.ExpiresAt, &session.IP, &session.Active, &session.Timestamp)
	if err != nil {
		log.Errorln("While fetching individual session data with token --> ", session.Token, " error is ", err)
		return session, err
	}
	session.ExpiryIn = calculateExpiresIn(session.ExpiresAt)
	return session, nil
}

// NewSession accepts the key to encode, along with the client's IP Address, inserts into the sessions table, sets the delete session timer, and returns a Session struct.
func NewSession(key string, ipAddress string) (Session, error) {
	var assignedSession Session
	assignedSession.Key = key
	assignedSession.IP = ipAddress

	// Calculate the session token here
	assignedSession.Token = calculateHash(assignedSession.Key, assignedSession.IP)
	// Calculate the ExpiresAt token here
	assignedSession.ExpiresAt = time.Now().In(timezoneLocation).Add(time.Second * time.Duration(sessionExpiryTime))

	// Insert this into the database here
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Errorln("Couldn't fetch a database connection. Error is --> ", err)
		return assignedSession, err
	}

	err = tx.QueryRow(`
		INSERT INTO sessions (key, token, expires_at, ip) VALUES ($1, $2, $3, $4) RETURNING id, timestamp
	`, assignedSession.Key, assignedSession.Token, assignedSession.ExpiresAt, assignedSession.IP).Scan(&assignedSession.ID, &assignedSession.Timestamp)

	if err != nil {
		log.Errorln("While inserting a new session in the sessions table, error is ", err)
		return assignedSession, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		tx.Rollback()
		log.Errorln("While inserting into sessions table, error is ", commitErr)
		return assignedSession, commitErr
	}
	// Calculate Session.ExpiresIn from Session.ExpiresAt (after adjusting it to the specified time zone)
	assignedSession.ExpiryIn = calculateExpiresIn(assignedSession.ExpiresAt)

	// Set up the delete timer here
	go func() {
		<-time.After(time.Second * time.Duration(assignedSession.ExpiryIn))
		go DeleteSession(assignedSession.Token)
	}()

	return assignedSession, nil
}

// CheckStatus accepts a session token, and returns the session object, along with an error, if any
func CheckStatus(sessionToken string) (Session, error) {
	var session Session
	err := db.QueryRow(`
		SELECT id, key, token, expires_at, ip, timestamp FROM sessions WHERE token = $1 AND active = True
	`, sessionToken).Scan(&session.ID, &session.Key, &session.Token, &session.ExpiresAt, &session.IP, &session.Timestamp)
	if err != nil {
		return session, err
	}
	session.ExpiryIn = calculateExpiresIn(session.ExpiresAt)
	return session, nil
}

// DeleteSession deletes the session with the specific sessionToken and returns an error, if any.
func DeleteSession(sessionToken string) error {
	tx, _ := db.Begin()
	defer tx.Rollback()

	_, err := tx.Exec(`
		UPDATE sessions SET active = False WHERE token = $1
	`, sessionToken)
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
	// var expiresIn int64
	// Calculate the current time in the specified timezone
	// Find the difference between expiredAt and the current time
	// difference := expiresAt.Sub(time.Now().In(timezoneLocation))
	// Find the difference in seconds and return the value
	// return expiresIn
	// log.Infoln("Expires At timezone is ", expiresAt.Location().String(), " and timezoneLocation is ", timezoneLocation.String())
	// zone, offset := expiresAt.Zone()
	// log.Infoln("Zone is ", zone, " and offset is ", offset)
	// return int64(expiresAt.Sub(time.Now().In(timezoneLocation))) / (1000 * 1000 * 1000)
	return int64(expiresAt.Sub(time.Now())) / (1000 * 1000 * 1000)
}

// calculateHash accepts the key, and the ip address of the client, and generates an SHA
func calculateHash(key string, ipAddress string) string {
	h := sha256.New()
	h.Write([]byte(key + uuid.NewV4().String() + ipAddress))
	return fmt.Sprintf("%x", h.Sum(nil))
}
