package sessions

import (
	"time"

	"github.com/Unaxiom/ulogger"
	"gopkg.in/jackc/pgx.v2"
)

// db holds the database connection pool that the application is using, so that the same pool could be used for working on sessions
var db *pgx.ConnPool

// log holds the same logger that the parent app uses
var log *ulogger.Logger

// DefaultExpiryTime stores the time after which a session needs to be cleared
var DefaultExpiryTime = 86400

// Timezone stores the timezone that needs to be followed
var Timezone = ""

type Session struct {
	// ID stores the bigserial (PostgreSQL row id) associated with this session
	ID int64 `json:"id,omitempty"`
	// Key stores the string that needs to be encoded/hashed to generate a unique session value
	Key string `json:"key,omitempty"`
	// Value stores the computed value using the key
	Value string `json:"value,omitempty"`
	// ExpiryIn stores the number of seconds after which this session will expire
	ExpiryIn int64 `json:"expiry_in,omitempty"`
	// ExpiresAt stores the timestamp (with full timezone) at which this session will expire
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	// IP stores the IP address of the client requesting a new session
	IP string `json:"ip,omitempty"`
	// Timestamp stores the timestamp in epoch secs when this entry was created
	Timestamp int64 `json:"timestamp,omitempty"`
}
