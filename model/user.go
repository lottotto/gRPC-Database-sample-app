package model

type User struct {
	Name   string `json:"name" db:"name"`
	Number int    `json:"number" db:"number"`
}
