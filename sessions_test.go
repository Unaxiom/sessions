package sessions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
)

var appName = "Sessions Module"
var orgName = "Unaxiom"
var dbHost = "localhost"
var dbPort = uint16(5432)
var dbName = "sessions_test"
var dbUser = "sessions"
var dbPassword = "sessions_password"

func assertSessDataWithProvidedAttributes(assert *require.Assertions, sess SessionData, key string, ipAddr string, expiryIn int64) {
	assert.NotZero(sess.ID)
	assert.Equal(sess.Key, key)
	assert.NotZero(len(sess.Token))
	// assert.Equal(sess.ExpiryIn, expiryIn)
	assert.True(sess.ExpiryIn < expiryIn)
	assert.NotNil(sess.ExpiresAt)
	assert.Equal(sess.IP, ipAddr)
	assert.Equal(sess.Active, true)
	assert.NotZero(sess.Timestamp)
}

func shortSessionTest(assert *require.Assertions) {
	// Init a new session with a small timeout
	shortTimeout := int64(2)
	shortSession, err := Init(shortTimeout, "", appName, orgName, false, dbHost, dbPort, dbName, dbUser, dbPassword)
	// Assert all the parameters
	assert.Nil(err)
	assert.NotNil(shortSession.db)
	assert.Equal(shortTimeout, shortSession.sessionExpiryTime)
	assert.Equal("UTC", shortSession.sessionTimezoneName)
	assert.NotNil(shortSession.timezoneLocation)

	// Create a new session
	var key = uuid.NewV4().String()
	newSess, err := shortSession.NewSession(key, "127.0.0.1")
	assert.Nil(err)
	// Assert parameters and the fact that the authtoken is still alive
	assertSessDataWithProvidedAttributes(assert, newSess, key, "127.0.0.1", shortTimeout)

	// Ensure that the key is deleted after this timeout
	<-time.After(time.Second * time.Duration(shortTimeout))
	// Fetch the session data here, after the timeout and assert that active is false
	newSess2, err := shortSession.FetchSessionData(SessionData{Token: newSess.Token})
	assert.Nil(err, newSess.Token)
	assert.False(newSess2.Active)

	newSess, err = shortSession.CheckStatus(newSess.Token)
	assert.NotNil(err)

}

func longSessionTest(assert *require.Assertions) {
	// Init a new session with 0 timeout -> and assert that the expiry time is 86400
	var longTimeout = int64(86400)
	longSession, err := Init(0, "IST", appName, orgName, false, dbHost, dbPort, dbName, dbUser, dbPassword)
	assert.Nil(err)
	assert.NotNil(longSession.db)
	assert.Equal(longTimeout, longSession.sessionExpiryTime)
	assert.Equal("IST", longSession.sessionTimezoneName)
	assert.NotNil(longSession.timezoneLocation)

	// Create new session
	var key = uuid.NewV4().String()
	newSess, err := longSession.NewSession(key, "127.0.0.1")
	assert.Nil(err)
	// Assert parameters and the fact that the authtoken is still alive
	assertSessDataWithProvidedAttributes(assert, newSess, key, "127.0.0.1", longTimeout)

	newSess, err = longSession.CheckStatus(newSess.Token)
	assert.Nil(err)
	// Manually delete the session using the API
	err = longSession.DeleteSession(newSess.Token)
	assert.Nil(err)

	// Ensure that this session no longer exists
	newSess, err = longSession.CheckStatus(newSess.Token)
	assert.NotNil(err)
}

func TestSessions(t *testing.T) {
	assert := require.New(t)

	// Run SetupTable
	err := SetupTable(appName, orgName, false, dbHost, dbPort, dbName, dbUser, dbPassword, "public")
	assert.Nil(err)

	shortSessionTest(assert)
	longSessionTest(assert)
}
