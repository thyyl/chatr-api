package user

type User struct {
	Id       uint64
	Email    string
	Name     string
	Photo    string
	AuthType AuthType
}

type AuthType string

const (
	LocalAuth  AuthType = "local"
	GoogleAuth AuthType = "google"
)
