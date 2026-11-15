package domain

import "time"

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleTeacher Role = "teacher"
	RoleStudent Role = "student"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	FullName     string
	Role         Role
	GroupID      *int64
	IsBlocked    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}