package models

import "errors"
// UserBasic represents the basic information of a user.
type UserBasic struct {
	UserID   UserID `json:"UserID"`
	Username string `json:"Username"`
}

// User represents the detailed information of a user.
type User struct {
	UserBasic
	Email           string                     `json:"Email"`
	GamesPlayed     []GameID                   `json:"GamesPlayed"`
}

// NewUserBasic creates a new UserBasic with initialized fields
func NewUserBasic(id UserID, username string) (*UserBasic, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	return &UserBasic{
		UserID:   id,
		Username: username,
	}, nil
}

// NewUser creates a new User with initialized fields
func NewUser(id UserID, username, email string, gamesPlayed []GameID) (*User, error) {
	userBasic, err := NewUserBasic(id, username)
	if err != nil {
		return nil, err
	}

	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	if gamesPlayed == nil {
		gamesPlayed = []GameID{}
	}

	return &User{
		UserBasic:      *userBasic,
		Email:          email,
		GamesPlayed:    gamesPlayed,
	}, nil
}