package sessions

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Unaxiom/ulogger"
	"github.com/apratheek/schemamagic"
	"github.com/twinj/uuid"
	"gopkg.in/jackc/pgx.v2"
)

func init() {

}

// Init accepts number of seconds after which the session expires (if set to 0, default value of 86400 is used), the timezone (if set to "", default timezone - UTC is used), and sets up the required variables. It also starts the expiry timers of existing sessions.
func Init(sessionExpiresInSecs int64, sessionLocalTimezoneName string, applicationName string, orgName string, production bool, dbHost string, dbPort uint16, databaseName string, dbUser string, dbPassword string) (*Session, error) {
	sessionObject := new(Session)
	log = ulogger.New()
	log.SetLogLevel(ulogger.InfoLevel)
	log.RemoteAvailable = production
	log.ApplicationName = applicationName + " sessions"
	log.OrganizationName = orgName
	if production {
		ulogger.RemoteURL = "https://logs.unaxiom.com/newlog"
	}
	var err error
	sessionObject.db, err = schemamagic.SetupDB(dbHost, dbPort, databaseName, dbUser, dbPassword)
	if err != nil {
		return nil, err
	}

	if sessionExpiresInSecs != 0 {
		sessionObject.sessionExpiryTime = sessionExpiresInSecs
	} else {
		sessionObject.sessionExpiryTime = int64(86400)
	}
	if sessionLocalTimezoneName != "" {
		sessionObject.sessionTimezoneName = sessionLocalTimezoneName
	} else {
		sessionObject.sessionTimezoneName = "UTC"
	}

	// Load the time zone here
	sessionObject.timezoneLocation, err = time.LoadLocation(sessionObject.sessionTimezoneName)
	if err != nil {
		log.Infoln("Couldn't load timezone location --> ", sessionObject.sessionTimezoneName, ". Error is ", err, ".")
		sessionObject.timezoneLocation, _ = time.LoadLocation("UTC")
	}

	// Fetch all the active sessions here
	allSessions, err := sessionObject.FetchAllSessions()
	if err != nil {
		log.Errorln("While trying to fetch all sessions during Setup, error is ", err)
		return nil, err
	}

	// Calculate the expiry timers and delete the sessions accordingly
	for _, session := range allSessions {
		go func(session SessionData) {
			<-time.After(time.Second * time.Duration(session.ExpiryIn))
			go sessionObject.DeleteSession(session.Token)
		}(session)
	}
	return sessionObject, nil
}

// SetupTable accepts all the parameters and creates/updates the Sessions table. Don't call it along with Init(). Init() and SetupTable() should be called separately.
func SetupTable(applicationName string, orgName string, production bool, dbHost string, dbPort uint16, databaseName string, dbUser string, dbPassword string, defaultSchema string) error {
	log = ulogger.New()
	log.SetLogLevel(ulogger.InfoLevel)
	log.RemoteAvailable = production
	log.ApplicationName = applicationName + " sessions"
	log.OrganizationName = orgName
	if production {
		ulogger.RemoteURL = "https://logs.unaxiom.com/newlog"
	}
	var err error
	db, err = schemamagic.SetupDB(dbHost, dbPort, databaseName, dbUser, dbPassword)
	if err != nil {
		return err
	}
	err = createSessionsTable(db, databaseName, defaultSchema)
	if err != nil {
		return err
	}
	return nil
}

// createSessionsTable accepts the schema of the database and creates the Sessions table using the database parameters passed during Setup. Setup needs to be called prior. Otherwise, method may crash.
func createSessionsTable(db *pgx.ConnPool, databaseName string, defaultSchema string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	table := tableSessions(tx, defaultSchema, databaseName)
	table.Begin()
	commitErr := tx.Commit()
	if commitErr != nil {
		tx.Rollback()
		return commitErr
	}
	return nil
}

// FetchAllSessions returns all the live sessions.
func (sessionObject *Session) FetchAllSessions() ([]SessionData, error) {
	allSessions := make([]SessionData, 0)
	rows, err := sessionObject.db.Query(`
		SELECT id, key, token, expires_at, ip, timestamp FROM sessions WHERE active = True
	`)
	defer rows.Close()
	if err != nil {
		log.Errorln("While trying to fetch all sessions in FetchAllSessions, error is ", err)
		return allSessions, err
	}
	for rows.Next() {
		var session SessionData
		rows.Scan(&session.ID, &session.Key, &session.Token, &session.ExpiresAt, &session.IP, &session.Timestamp)
		session.ExpiryIn = calculateExpiresIn(session.ExpiresAt)
		allSessions = append(allSessions, session)
	}
	return allSessions, nil
}

// FetchSessionData returns the data of the particular session token
func (sessionObject *Session) FetchSessionData(session SessionData) (SessionData, error) {
	err := sessionObject.db.QueryRow(`
		SELECT id, key, token, expires_at, ip, active, timestamp FROM sessions WHERE token = $1
	`, session.Token).Scan(&session.ID, &session.Key, &session.Token, &session.ExpiresAt, &session.IP, &session.Active, &session.Timestamp)
	if err != nil {
		log.Errorln("While fetching individual session data with token --> ", session.Token, " error is ", err)
		return session, err
	}
	session.ExpiryIn = calculateExpiresIn(session.ExpiresAt)
	return session, nil
}

// NewSession accepts the key to encode, along with the client's IP Address, inserts into the sessions table, sets the delete session timer, and returns a SessionData struct.
func (sessionObject *Session) NewSession(key string, ipAddress string) (SessionData, error) {
	var assignedSession SessionData
	assignedSession.Key = key
	assignedSession.IP = ipAddress

	// Calculate the session token here
	assignedSession.Token = calculateHash(assignedSession.Key, assignedSession.IP)
	// Calculate the ExpiresAt token here
	assignedSession.ExpiresAt = time.Now().In(sessionObject.timezoneLocation).Add(time.Second * time.Duration(sessionObject.sessionExpiryTime))

	// Insert this into the database here
	tx, err := sessionObject.db.Begin()
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
	// Calculate SessionData.ExpiresIn from SessionData.ExpiresAt (after adjusting it to the specified time zone)
	assignedSession.ExpiryIn = calculateExpiresIn(assignedSession.ExpiresAt)
	assignedSession.Active = true

	// Set up the delete timer here
	go func() {
		<-time.After(time.Second * time.Duration(assignedSession.ExpiryIn))
		go sessionObject.DeleteSession(assignedSession.Token)
	}()

	return assignedSession, nil
}

// CheckStatus accepts a session token, and returns the session object, along with an error, if any
func (sessionObject *Session) CheckStatus(sessionToken string) (SessionData, error) {
	var session SessionData
	err := sessionObject.db.QueryRow(`
		SELECT id, key, token, expires_at, ip, timestamp FROM sessions WHERE token = $1 AND active = True
	`, sessionToken).Scan(&session.ID, &session.Key, &session.Token, &session.ExpiresAt, &session.IP, &session.Timestamp)
	if err != nil {
		return session, err
	}
	session.ExpiryIn = calculateExpiresIn(session.ExpiresAt)
	return session, nil
}

// DeleteSession deletes the session with the specific sessionToken and returns an error, if any.
func (sessionObject *Session) DeleteSession(sessionToken string) error {
	tx, _ := sessionObject.db.Begin()
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
	return int64(expiresAt.Sub(time.Now())) / (1000 * 1000 * 1000)
}

// calculateHash accepts the key, and the ip address of the client, and generates an SHA
func calculateHash(key string, ipAddress string) string {
	h := sha256.New()
	h.Write([]byte(key + uuid.NewV4().String() + ipAddress))
	return fmt.Sprintf("%x", h.Sum(nil))
}
