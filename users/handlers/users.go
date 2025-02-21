package handlers

import (
	"time"

	"github.com/vod/users/config"
	"github.com/vod/users/dbiface"
)

// User represents a user
type User struct {
	Email     string    `json:"user_name" bson:"user_name" validate:"required,email"`
	Password  string    `json:"password,omitempty" bson:"password" validate:"required,min=8,max=300"`
	IsAdmin   bool      `json:"is_admin,omitempty" bson:"is_admin" validate:"required"`
	Username  string    `json:"username" bson:"username" validate:"required"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

/*
{
	"user_name" : "Ananth",
	"password": "somePassword",
	"is_admin": "true"
}
*/
// UsersHandler users handler
type UsersHandler struct {
	Col dbiface.MongoCollectionAPI
}

type errorMessage struct {
	Message string `json:"message"`
}

var (
	prop config.Properties
)
