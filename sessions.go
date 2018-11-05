package sessions

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"encoding/json"

	"github.com/go-redis/redis"
	"github.com/tidwall/buntdb"

	"github.com/twinj/uuid"
)

// Init initialises a session object with buntdb as the storage engine a name, a bool representing if the corresponding file needs
// to be saved to disk, a key expiry time in seconds, and the location of the folder where the database file is stored
func Init(name string, persistSessionToDisk bool, expiryInSecs int64, dbFolderLocation string) (*Session, error) {
	sessionObject := new(Session)
	sessionObject.Name = name
	sessionObject.StorageEngine = buntStorage
	if expiryInSecs != 0 {
		sessionObject.ExpiryTime = expiryInSecs
	} else {
		sessionObject.ExpiryTime = int64(86400)
	}
	var err error
	var filename = ""

	if persistSessionToDisk {
		// Make the directory here
		os.MkdirAll(filepath.Join(dbFolderLocation, "sessionsdb"), os.ModePerm)
		filename = filepath.Join(dbFolderLocation, "sessionsdb", sessionObject.Name+".db")
	} else {
		filename = ":memory:"
	}

	sessionObject.buntDB, err = buntdb.Open(filename)
	return sessionObject, err
}

// InitRedis initialises a session object with Redis as the storage engine. Pass "" for redisAddr & password, and 0 for db for defaults.
func InitRedis(redisAddr string, password string, db int, expiryInSecs int64) (*Session, error) {
	sessionObject := new(Session)
	sessionObject.Name = "Redis"
	sessionObject.StorageEngine = redisStorage
	if expiryInSecs != 0 {
		sessionObject.ExpiryTime = expiryInSecs
	} else {
		sessionObject.ExpiryTime = int64(86400)
	}

	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	sessionObject.redisDB = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})

	var err error
	return sessionObject, err
}

// FetchSessionData returns the data of the particular session token
func (sessionObject *Session) FetchSessionData(session SessionData) (SessionData, error) {
	return sessionObject.CheckStatus(session.Token)
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
	if sessionObject.StorageEngine == buntStorage {
		err := sessionObject.buntDB.Update(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set(assignedSession.Token, string(data), &buntdb.SetOptions{Expires: true, TTL: time.Second * time.Duration(sessionObject.ExpiryTime)})
			return err
		})

		if err != nil {
			return assignedSession, err
		}
	} else if sessionObject.StorageEngine == redisStorage {
		err := sessionObject.redisDB.Set(assignedSession.Token, string(data), time.Second*time.Duration(sessionObject.ExpiryTime))
		if err.Err() != nil {
			return assignedSession, err.Err()
		}
	}
	// Calculate SessionData.ExpiresIn from SessionData.ExpiresAt (after adjusting it to the specified time zone)
	assignedSession.ExpiryIn = calculateExpiresIn(assignedSession.ExpiresAt)

	return assignedSession, nil
}

// CheckStatus accepts a session token, and returns the session object, along with an error, if any
func (sessionObject *Session) CheckStatus(sessionToken string) (SessionData, error) {
	var session SessionData
	if sessionObject.StorageEngine == buntStorage {
		err := sessionObject.buntDB.View(func(tx *buntdb.Tx) error {
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
	} else if sessionObject.StorageEngine == redisStorage {
		val := sessionObject.redisDB.Get(sessionToken)
		err := json.Unmarshal([]byte(val.Val()), &session)
		if err != nil {
			return session, err
		}
	}

	session.ExpiryIn = calculateExpiresIn(session.ExpiresAt)
	return session, nil
}

// DeleteSession deletes the session with the specific sessionToken and returns an error, if any.
func (sessionObject *Session) DeleteSession(sessionToken string) error {
	var err error
	if sessionObject.StorageEngine == buntStorage {
		err = sessionObject.buntDB.Update(func(tx *buntdb.Tx) error {
			tx.Delete(sessionToken)
			return nil
		})
	} else if sessionObject.StorageEngine == redisStorage {
		errR := sessionObject.redisDB.Del(sessionToken)
		err = errR.Err()
	}
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
