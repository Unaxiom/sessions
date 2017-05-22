### Sessions
Package that will be used across Unaxiom to generate and maintain sessions.

### Installation
`go get -u github.com/Unaxiom/sessions`

### Usage
```
import (
    "github.com/Unaxiom/sessions"
    "github.com/Unaxiom/ulogger"
    "gopkg.in/jackc/pgx.v2"
)

func init() {
    // Set up the database and the logger objects
    sessionExpiryTime := int64(86400)
    localTimeZone := "UTC"
    sessions.Setup(dbObject, uloggerObject, sessionExpiryTime, localTimeZone)
}
```
