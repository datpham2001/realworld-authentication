package repository

import (
	"realworld-authentication/dto"
	"realworld-authentication/helper"
)

type AuthRepository interface {
	SignUp(input *dto.UserSignUpRequest) *helper.APIResponse
	Login(input *dto.UserLoginRequest) *helper.APIResponse
}
