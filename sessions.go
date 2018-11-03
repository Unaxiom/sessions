package sessions

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"encoding/json"

	"github.com/tidwall/buntdb"

	"github.com/twinj/uuid"
)

// Init initialises a session object with a name, a bool representing if the corresponding file needs
// to be saved to disk, and a key expiry time in seconds.
func Init(name string, persistSessionToDisk bool, expiryInSecs int64) (*Session, error) {
	sessionObject := new(Session)
	sessionObject.Name = name
	if expiryInSecs != 0 {
		sessionObject.ExpiryTime = expiryInSecs
	} else {
		sessionObject.ExpiryTime = int64(86400)
	}
	var err error
	var filename = ""

	// Make the directory here
	cwd, _ := os.Getwd()
	os.MkdirAll(filepath.Join(cwd, "sessionsdb"), os.ModePerm)

	if persistSessionToDisk {
		filename = filepath.Join(cwd, "sessionsdb", sessionObject.Name+".db")
	} else {
		filename = ":memory:"
	}

	sessionObject.DB, err = buntdb.Open(filename)
	return sessionObject, err
}

// FetchSessionData returns the data of the particular session token
func (sessionObject *Session) FetchSessionData(session SessionData) (SessionData, error) {
	err := sessionObject.DB.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(session.Token)
		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(val), &session)
		return err
	})
	if err != nil {
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
	assignedSession.ExpiresAt = time.Now().Add(time.Second * time.Duration(sessionObject.ExpiryTime))
	assignedSession.Timestamp = time.Now().Unix()

	data, _ := json.Marshal(assignedSession)

	// Insert this into the database here
	err := sessionObject.DB.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(assignedSession.Token, string(data), &buntdb.SetOptions{Expires: true, TTL: time.Second * time.Duration(sessionObject.ExpiryTime)})
		return err
	})

	if err != nil {
		return assignedSession, err
	}
	// Calculate SessionData.ExpiresIn from SessionData.ExpiresAt (after adjusting it to the specified time zone)
	assignedSession.ExpiryIn = calculateExpiresIn(assignedSession.ExpiresAt)

	return assignedSession, nil
}

// CheckStatus accepts a session token, and returns the session object, along with an error, if any
func (sessionObject *Session) CheckStatus(sessionToken string) (SessionData, error) {
	var session SessionData

	err := sessionObject.DB.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(sessionToken)
		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(val), &session)
		return err
	})
	if err != nil {
		return session, err
	}

	session.ExpiryIn = calculateExpiresIn(session.ExpiresAt)
	return session, nil
}

// DeleteSession deletes the session with the specific sessionToken and returns an error, if any.
func (sessionObject *Session) DeleteSession(sessionToken string) error {
	err := sessionObject.DB.Update(func(tx *buntdb.Tx) error {
		tx.Delete(sessionToken)
		return nil
	})
	return err
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
