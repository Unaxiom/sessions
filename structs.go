package sessions

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/tidwall/buntdb"
)

const (
	buntStorage  = "buntdb"
	redisStorage = "redis"
)

// Session manages all the sessions
type Session struct {
	Name          string
	buntDB        *buntdb.DB
	redisDB       *redis.Client
	ExpiryTime    int64
	StorageEngine string
}

// SessionData stores the parameters in the sessions table
type SessionData struct {
	// Key stores the string that needs to be encoded/hashed to generate a unique session value
	Key string `json:"key,omitempty"`
	// Token stores the computed token using the key
	Token string `json:"token,omitempty"`
	// ExpiryIn stores the number of seconds after which this session will expire
	ExpiryIn int64 `json:"expiry_in,omitempty"`
	// ExpiresAt stores the timestamp (with full timezone) at which this session will expire
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	// IP stores the IP address of the client requesting a new session
	IP string `json:"ip,omitempty"`
	// Timestamp stores the timestamp in epoch secs when this entry was created
	Timestamp int64 `json:"timestamp,omitempty"`
}
