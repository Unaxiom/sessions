### Sessions
Package that will be used across Unaxiom to generate and maintain sessions.

### Installation
`go get -u github.com/Unaxiom/sessions`

### Usage
###### To create the sessions table
```
err := SetupTable(applicationName string, orgName string, production bool, dbHost string, dbPort uint16, databaseName string, dbUser string, dbPassword string, defaultSchema string)
```
###### To start the sessions process
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
    sessionObject, err := sessions.Init(sessionExpiresInSecs int64, sessionLocalTimezoneName string, applicationName string, orgName string, production bool, dbHost string, dbPort uint16, databaseName string, dbUser string, dbPassword string)
    sessionData, err := sessionObject.NewSession("somekeyhere", "userIPAddress")
    fmt.Println("Auth Token is ", sessionData.Token)
}
```
