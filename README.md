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

Also create a HIDDEN_CONFIG.go file with package login in the login directory with the following
```go
package login

const(
	JWT_SIGNING_SECRET="..."
)
```

Save and run
```bash
go install
```
Finally launch ```$(GOPATH)/bin/SermireApi```

## Setup for Development

First ensure that ```DEV=true``` is set in the HIDDEN_CONFIG.go file. 

Then ensure that you have a working MongoDB instance on your computer with the ```MONGODB_CONN_INFO``` constant set to the mongodb connection string of that instance. In that instance create a ```Sermire``` database with the following collections

```
Books
Posts
Stickers
Users
```

Optionally data can be exported to json from the live server and then imported locally (done easily using MongoDB Compass).

After this simply run
```
go build
```
and launch the executable. 