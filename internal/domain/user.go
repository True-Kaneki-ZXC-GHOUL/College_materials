package domain

type PasswordHash struct {
	Hash []byte
	Salt []byte
}

type User struct {
	ID   int
	Role string
}
