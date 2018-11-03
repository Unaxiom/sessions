package sessions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
)

func assertSessDataWithProvidedAttributes(assert *require.Assertions, sess SessionData, key string, ipAddr string, expiryIn int64) {
	assert.Equal(sess.Key, key)
	assert.NotZero(len(sess.Token))
	assert.True(sess.ExpiryIn < expiryIn)
	assert.NotNil(sess.ExpiresAt)
	assert.Equal(sess.IP, ipAddr)
	assert.NotZero(sess.Timestamp)
}

func shortSessionTest(assert *require.Assertions) {
	// Init a new session with a small timeout
	shortTimeout := int64(2)
	shortSession, err := Init("csrf", false, shortTimeout)
	// Assert all the parameters
	assert.Nil(err)
	assert.NotNil(shortSession.DB)
	assert.Equal(shortTimeout, shortSession.ExpiryTime)

	// Create a new session
	var key = uuid.NewV4().String()
	newSess, err := shortSession.NewSession(key, "127.0.0.1")
	assert.Nil(err)
	// Assert parameters and the fact that the authtoken is still alive
	assertSessDataWithProvidedAttributes(assert, newSess, key, "127.0.0.1", shortTimeout)

	// Ensure that the key is deleted after this timeout
	<-time.After(time.Second * time.Duration(shortTimeout))
	// Fetch the session data here, after the timeout and assert that active is false
	_, err = shortSession.FetchSessionData(SessionData{Token: newSess.Token})
	assert.NotNil(err, newSess.Token)

	newSess, err = shortSession.CheckStatus(newSess.Token)
	assert.NotNil(err)

	assert.Nil(shortSession.DB.Close())

	_, err = shortSession.NewSession(key, "127.0.0.1")
	assert.NotNil(err)

}

func longSessionTest(assert *require.Assertions) {
	// Init a new session with 0 timeout -> and assert that the expiry time is 86400
	var longTimeout = int64(86400)
	longSession, err := Init("sessions", true, 0)
	assert.Nil(err)
	assert.NotNil(longSession.DB)
	assert.Equal(longTimeout, longSession.ExpiryTime)

	// Create new session
	var key = uuid.NewV4().String()
	newSess, err := longSession.NewSession(key, "127.0.0.1")
	assert.Nil(err)
	// Assert parameters and the fact that the authtoken is still alive
	assertSessDataWithProvidedAttributes(assert, newSess, key, "127.0.0.1", longTimeout)

	newSess, err = longSession.CheckStatus(newSess.Token)
	assert.Nil(err)

	newSess2, err := longSession.FetchSessionData(SessionData{Token: newSess.Token})
	assert.Nil(err, newSess2.Token)

	assert.Equal(newSess.Key, newSess2.Key)
	assert.Equal(newSess.Token, newSess2.Token)
	assert.Equal(newSess.ExpiryIn, newSess2.ExpiryIn)
	assert.Equal(newSess.ExpiresAt, newSess2.ExpiresAt)
	assert.Equal(newSess.IP, newSess2.IP)
	assert.Equal(newSess.Timestamp, newSess2.Timestamp)

	assertSessDataWithProvidedAttributes(assert, newSess2, key, "127.0.0.1", longTimeout)

	// Manually delete the session using the API
	err = longSession.DeleteSession(newSess.Token)
	assert.Nil(err)

	// Ensure that this session no longer exists
	newSess, err = longSession.CheckStatus(newSess.Token)
	assert.NotNil(err)
}

func TestSessions(t *testing.T) {
	assert := require.New(t)
	shortSessionTest(assert)
	longSessionTest(assert)
}

func BenchmarkShort(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	shortTimeout := int64(2)
	shortSession, err := Init("csrf", false, shortTimeout)
	assert.Nil(err)
	for n := 0; n < b.N; n++ {
		newSess, err := shortSession.NewSession(uuid.NewV4().String(), "127.0.0.1")
		assert.Nil(err)
		assert.NotZero(len(newSess.Token))
	}
}

func BenchmarkShortFull(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	for n := 0; n < b.N; n++ {
		shortSessionTest(assert)
	}
}

func BenchmarkLong(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	longSession, err := Init("sessions", true, 0)
	assert.Nil(err)
	for n := 0; n < b.N; n++ {
		var key = uuid.NewV4().String()
		newSess, err := longSession.NewSession(key, "127.0.0.1")
		assert.Nil(err)
		assert.NotZero(len(newSess.Token))
	}
}

func BenchmarkLongFull(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	for n := 0; n < b.N; n++ {
		longSessionTest(assert)
	}
}
