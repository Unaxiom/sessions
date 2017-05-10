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
    sessions.Setup(dbObject, uloggerObject)
}
```
