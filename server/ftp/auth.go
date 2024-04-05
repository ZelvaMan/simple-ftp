package ftp

func authenticateUser(username string, password string) bool {
	return username == "zm" && password == "password"
}
