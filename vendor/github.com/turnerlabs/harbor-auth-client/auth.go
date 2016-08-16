package harborauth

//Auth -
type Auth interface {
	// Login -
	Login(username string, password string) (string, bool, error)
	// Logout
	Logout(username string, token string) (bool, error)
	// IsAuthenticated -
	IsAuthenticated(username string, token string) (bool, error)
}
