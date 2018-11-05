package sessions

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
)

func assertSessDataWithProvidedAttributes(assert *require.Assertions, sess SessionData, key string, ipAddr string, expiryIn int64) {
	assert.Equal(sess.Key, key)
	assert.NotZero(len(sess.Token))
	// assert.NotZero(sess.ExpiryIn)
	assert.True(sess.ExpiryIn < expiryIn)
	assert.NotNil(sess.ExpiresAt)
	assert.Equal(sess.IP, ipAddr)
	assert.NotZero(sess.Timestamp)
}

func shortBuntDBSessionTest(assert *require.Assertions) {
	// Init a new session with a small timeout
	shortTimeout := int64(2)
	cwd, _ := os.Getwd()
	shortBuntDBSession, err := Init("csrf", false, shortTimeout, cwd)
	// Assert all the parameters
	assert.Nil(err)
	assert.NotNil(shortBuntDBSession.buntDB)
	assert.Equal(shortTimeout, shortBuntDBSession.ExpiryTime)

	// Create a new session
	var key = uuid.NewV4().String()
	newSess, err := shortBuntDBSession.NewSession(key, "127.0.0.1")
	assert.Nil(err)
	// Assert parameters and the fact that the authtoken is still alive
	assertSessDataWithProvidedAttributes(assert, newSess, key, "127.0.0.1", shortTimeout)

	// Ensure that the key is deleted after this timeout
	<-time.After(time.Second * time.Duration(shortTimeout))
	// Fetch the session data here, after the timeout and assert that active is false
	_, err = shortBuntDBSession.FetchSessionData(SessionData{Token: newSess.Token})
	assert.NotNil(err, newSess.Token)

	newSess, err = shortBuntDBSession.CheckStatus(newSess.Token)
	assert.NotNil(err)

	assert.Nil(shortBuntDBSession.buntDB.Close())

	_, err = shortBuntDBSession.NewSession(key, "127.0.0.1")
	assert.NotNil(err)

}

func longBuntDBSessionTest(assert *require.Assertions) {
	// Init a new session with 0 timeout -> and assert that the expiry time is 86400
	var longTimeout = int64(86400)
	cwd, _ := os.Getwd()
	longBuntDBSession, err := Init("sessions", true, 0, cwd)
	assert.Nil(err)
	assert.NotNil(longBuntDBSession.buntDB)
	assert.Equal(longTimeout, longBuntDBSession.ExpiryTime)

	// Create new session
	var key = uuid.NewV4().String()
	newSess, err := longBuntDBSession.NewSession(key, "127.0.0.1")
	assert.Nil(err)
	// Assert parameters and the fact that the authtoken is still alive
	assertSessDataWithProvidedAttributes(assert, newSess, key, "127.0.0.1", longTimeout)

	newSess, err = longBuntDBSession.CheckStatus(newSess.Token)
	assert.Nil(err)

	newSess2, err := longBuntDBSession.FetchSessionData(SessionData{Token: newSess.Token})
	assert.Nil(err, newSess2.Token)

	assert.Equal(newSess.Key, newSess2.Key)
	assert.Equal(newSess.Token, newSess2.Token)
	assert.Equal(newSess.ExpiryIn, newSess2.ExpiryIn)
	assert.Equal(newSess.ExpiresAt, newSess2.ExpiresAt)
	assert.Equal(newSess.IP, newSess2.IP)
	assert.Equal(newSess.Timestamp, newSess2.Timestamp)

	assertSessDataWithProvidedAttributes(assert, newSess2, key, "127.0.0.1", longTimeout)

	// Manually delete the session using the API
	err = longBuntDBSession.DeleteSession(newSess.Token)
	assert.Nil(err)

	// Ensure that this session no longer exists
	newSess, err = longBuntDBSession.CheckStatus(newSess.Token)
	assert.NotNil(err)
}

func TestBuntDBSessions(t *testing.T) {
	assert := require.New(t)
	shortBuntDBSessionTest(assert)
	longBuntDBSessionTest(assert)
}

func BenchmarkBuntDBShort(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	shortTimeout := int64(2)
	cwd, _ := os.Getwd()
	shortBuntDBSession, err := Init("csrf", false, shortTimeout, cwd)
	assert.Nil(err)
	for n := 0; n < b.N; n++ {
		newSess, err := shortBuntDBSession.NewSession(uuid.NewV4().String(), "127.0.0.1")
		assert.Nil(err)
		assert.NotZero(len(newSess.Token))
	}
}

func BenchmarkBuntDBShortFull(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	for n := 0; n < b.N; n++ {
		shortBuntDBSessionTest(assert)
	}
}

func BenchmarkBuntDBLong(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	cwd, _ := os.Getwd()
	longBuntDBSession, err := Init("sessions", true, 0, cwd)
	assert.Nil(err)
	for n := 0; n < b.N; n++ {
		var key = uuid.NewV4().String()
		newSess, err := longBuntDBSession.NewSession(key, "127.0.0.1")
		assert.Nil(err)
		assert.NotZero(len(newSess.Token))
	}
}

func BenchmarkBuntDBLongFull(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	for n := 0; n < b.N; n++ {
		longBuntDBSessionTest(assert)
	}
}
