# Sessions

Package that will be used across Unaxiom to generate and maintain sessions.

## v2

Package has been rewritten using a faster [in-memory DB](https://github.com/tidwall/buntdb). PostgreSQL is no longer required. This version is incompatible with v1.

## Installation

`go get -u github.com/Unaxiom/sessions`

### Dependencies

1. github.com/tidwall/buntdb (for using BuntDB as the backend store)
2. github.com/twinj/uuid
3. github.com/go-redis/redis (for using Redis as the backend store)

### Import

```golang
import (
    "github.com/Unaxiom/sessions"
)
```

#### Usage

```golang
sessionExpiryTime := int64(86400)
dbFolder, _ := os.Getwd()
sessionObject, err := Init("name_of_session", false, sessionExpiryTime, dbFolder)
sessionData, err := sessionObject.NewSession("somekeyhere", "userIPAddress", 0)
fmt.Println("Auth Token is ", sessionData.Token)

// Check the status of the token. An error is returned in case the token does not exist. Returns nil otherwise.
_, err = sessionObject.CheckStatus(sessionData.Token)

// To delete a token
sessionObject.DeleteSession(sessionData.Token)

// For using Redis as the backend store
shortRedisDBSession, err := InitRedis("", "", 0, sessionExpiryTime)
```
