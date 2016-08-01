# harbor-auth-client

Grab the library via go get:
```
go get github.com/turnerlabs/harbor-auth-client
```
This library wraps the 3 auth calls for harbor auth.


Login
```
Login(username string, password string) (string, bool, error)
```
returns the token, whether the call was successful, and the error if one exists.


IsAuthenticated
```
IsAuthenticated(username string, token string) (bool, error)
```
returns whether the call was successful, and the error if one exists.


Logout
```
Logout(username string, token string) (bool, error)
```
returns whether the call was successful, and the error if one exists.


The code can be tested like this:
```
godep go test -v ./... -url=<auth service url> -username=<valid username> -password=<valid password>
```
