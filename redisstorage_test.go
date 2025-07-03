package sessions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/twinj/uuid"
)

func shortRedisDBSessionTest(assert *require.Assertions) {
	// InitRedis a new session with a small timeout
	shortTimeout := int64(2)
	shortRedisDBSession, err := InitRedis("", "", 0, shortTimeout)
	// Assert all the parameters
	assert.Nil(err)
	assert.NotNil(shortRedisDBSession.redisDB)
	assert.Equal(shortTimeout, shortRedisDBSession.ExpiryTime)

	// Create a new session
	var key = uuid.NewV4().String()
	newSess, err := shortRedisDBSession.NewSession(key, "127.0.0.1", 0)
	assert.Nil(err)
	// Assert parameters and the fact that the authtoken is still alive
	assertSessDataWithProvidedAttributes(assert, newSess, key, "127.0.0.1", shortTimeout)

	// Ensure that the key is deleted after this timeout
	<-time.After(time.Millisecond * time.Duration(shortTimeout*1000+20))
	// Fetch the session data here, after the timeout and assert that active is false
	_, err = shortRedisDBSession.FetchSessionData(SessionData{Token: newSess.Token})
	assert.NotNil(err, newSess.Token)

	newSess, err = shortRedisDBSession.CheckStatus(newSess.Token)
	assert.NotNil(err)

	assert.Nil(shortRedisDBSession.redisDB.Close())

	_, err = shortRedisDBSession.NewSession(key, "127.0.0.1", 0)
	assert.NotNil(err)

}

func longRedisDBSessionTest(assert *require.Assertions) {
	// InitRedis a new session with 0 timeout -> and assert that the expiry time is 86400
	var longTimeout = int64(86400)
	longRedisDBSession, err := InitRedis("", "", 0, 0)
	assert.Nil(err)
	assert.NotNil(longRedisDBSession.redisDB)
	assert.Equal(longTimeout, longRedisDBSession.ExpiryTime)

	// Create new session
	var key = uuid.NewV4().String()
	newSess, err := longRedisDBSession.NewSession(key, "127.0.0.1", 0)
	assert.Nil(err)
	// Assert parameters and the fact that the authtoken is still alive
	assertSessDataWithProvidedAttributes(assert, newSess, key, "127.0.0.1", longTimeout)

	newSess, err = longRedisDBSession.CheckStatus(newSess.Token)
	assert.Nil(err)

	newSess2, err := longRedisDBSession.FetchSessionData(SessionData{Token: newSess.Token})
	assert.Nil(err, newSess2.Token)

	assert.Equal(newSess.Key, newSess2.Key)
	assert.Equal(newSess.Token, newSess2.Token)
	assert.Equal(newSess.ExpiryIn, newSess2.ExpiryIn)
	assert.Equal(newSess.ExpiresAt, newSess2.ExpiresAt)
	assert.Equal(newSess.IP, newSess2.IP)
	assert.Equal(newSess.Timestamp, newSess2.Timestamp)

	assertSessDataWithProvidedAttributes(assert, newSess2, key, "127.0.0.1", longTimeout)

	// Manually delete the session using the API
	err = longRedisDBSession.DeleteSession(newSess.Token)
	assert.Nil(err)

	// Ensure that this session no longer exists
	newSess, err = longRedisDBSession.CheckStatus(newSess.Token)
	assert.NotNil(err)
}

func configurableDBSessionTest(assert *require.Assertions) {
	// InitRedis a new session with 0 timeout -> and assert that the expiry time is 86400
	var longTimeout = uint64(1000)
	longRedisDBSession, err := InitRedis("", "", 0, 0)
	assert.Nil(err)
	assert.NotNil(longRedisDBSession.redisDB)
	assert.Equal(uint64(86400), uint64(longRedisDBSession.ExpiryTime))

	// Create new session
	var key = uuid.NewV4().String()
	newSess, err := longRedisDBSession.NewSession(key, "127.0.0.1", longTimeout)
	assert.Nil(err)
	// Assert parameters and the fact that the authtoken is still alive
	assertSessDataWithProvidedAttributes(assert, newSess, key, "127.0.0.1", int64(longTimeout))
}

func TestRedisDBSessions(t *testing.T) {
	assert := require.New(t)
	shortRedisDBSessionTest(assert)
	longRedisDBSessionTest(assert)
	configurableDBSessionTest(assert)
}

func BenchmarkRedisDBShort(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	shortTimeout := int64(2)
	shortRedisDBSession, err := InitRedis("", "", 0, shortTimeout)
	assert.Nil(err)
	for n := 0; n < b.N; n++ {
		newSess, err := shortRedisDBSession.NewSession(uuid.NewV4().String(), "127.0.0.1", 0)
		assert.Nil(err)
		assert.NotZero(len(newSess.Token))
	}
}

func BenchmarkRedisDBShortFull(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	for n := 0; n < b.N; n++ {
		shortRedisDBSessionTest(assert)
	}
}

func BenchmarkRedisDBLong(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	longRedisDBSession, err := InitRedis("", "", 0, 0)
	assert.Nil(err)
	for n := 0; n < b.N; n++ {
		var key = uuid.NewV4().String()
		newSess, err := longRedisDBSession.NewSession(key, "127.0.0.1", 0)
		assert.Nil(err)
		assert.NotZero(len(newSess.Token))
	}
}

func BenchmarkRedisDBLongFull(b *testing.B) {
	t := new(testing.T)
	assert := require.New(t)
	for n := 0; n < b.N; n++ {
		longRedisDBSessionTest(assert)
	}
}
