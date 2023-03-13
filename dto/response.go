package dto

import "realworld-authentication/model"

type UserSignUpResponse struct {
	User *model.User `json:"user"`
}

type UserLoginResponse struct {
	User struct {
		Email       string `json:"email,omitempty"`
		Username    string `json:"username,omitempty"`
		AccessToken string `json:"accessToken,omitempty"`
	} `json:"user"`
}
