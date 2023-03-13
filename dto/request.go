package dto

import "realworld-authentication/model"

type UserSignUpRequest struct {
	User struct {
		Email    string `json:"email,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"user"`
}

type UserLoginRequest struct {
	User struct {
		Email    string `json:"email,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"user"`
}

type UserUpdateRequest struct {
	User struct {
		model.User
	} `json:"user"`
}
