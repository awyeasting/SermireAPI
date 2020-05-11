# Sermire API
Repository for the Sermire API server

## Installing
First install all dependencies
```bash
go get "github.com/sirupsen/logrus"
go get "go.mongodb.org/mongo-driver"
go get "github.com/go-chi/chi"
go get "golang.org/x/crypto/bcrypt"
go get "github.com/dgrijalva/jwt-go"
```

Next create a HIDDEN_CONFIG.go file with package main in the root of the directory with the following:
```go
package main

const (
	MONGODB_CONN_INFO="..."
	TLS_CERT_PATH="..."
	TLS_KEY_PATH="..."
	DEV=?					// Mainly whether to use TLS or not
)
```
Save and run
```bash
go install
```
Finally launch ```$(GOPATH)/bin/SermireApi```