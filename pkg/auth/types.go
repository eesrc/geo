package auth

type AuthProvider string

const (
	AuthGithub  AuthProvider = "github"
	AuthConnect AuthProvider = "connect"
	AuthToken   AuthProvider = "token"
)
