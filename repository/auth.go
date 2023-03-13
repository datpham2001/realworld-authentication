package repository

import (
	"net/http"
	"realworld-authentication/dto"
	"realworld-authentication/helper"
	"realworld-authentication/model"
	"realworld-authentication/model/enum"
	"realworld-authentication/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	AUTH_COLLECTION_NAME = "authentication"
)

func NewAuthRepository(db *mongo.Database) AuthRepository {
	r := &Instance{
		db:     db,
		DBName: db.Name(),

		coll:           db.Collection(AUTH_COLLECTION_NAME),
		TemplateObject: &model.User{},
	}

	return r
}

func (r *Instance) SignUp(input *dto.UserSignUpRequest) *helper.APIResponse {
	user := &model.User{
		Email:    input.User.Email,
		Username: input.User.Username,
	}

	_, err := r.QueryOne(model.User{
		ComplexQuery: []*bson.M{
			{
				"$or": []*bson.M{{
					"username": user.Username,
				}, {
					"email": user.Email,
				}},
			},
		},
	})

	if err != nil {
		return &helper.APIResponse{
			Code:      http.StatusBadGateway,
			Status:    helper.APIStatus.Invalid,
			Message:   "Username or email is existed",
			ErrorCode: string(enum.ErrorCodeExisted.UsernameOrEmail),
		}
	}

	user.HashedPassword, err = helper.HashPassword(input.User.Password)
	if err != nil {
		return &helper.APIResponse{
			Code:      http.StatusBadRequest,
			Status:    helper.APIStatus.Invalid,
			Message:   "Cannot hash user password " + err.Error(),
			ErrorCode: string(enum.ErrorCodePackage.Bcrypt),
		}
	}

	user.UserID = utils.GenAccountID()
	user.Status = enum.UserStatus.Active
	user.Role = enum.UserRole.User

	userCreateResp, err := r.Create(user)
	if err != nil {
		return &helper.APIResponse{
			Code:      http.StatusBadRequest,
			Status:    helper.APIStatus.Error,
			Message:   "Failed to create new user " + err.Error(),
			ErrorCode: string(enum.ErrorCodeDatabaseOperation.Create),
		}
	}

	return &helper.APIResponse{
		Code:    http.StatusCreated,
		Status:  helper.APIStatus.Ok,
		Message: "Signup user successfully",
		Data:    userCreateResp.([]*model.User)[0],
	}
}

func (r *Instance) Login(input *dto.UserLoginRequest) *helper.APIResponse {
	existUserResp, err := r.QueryOne(model.User{
		Email: input.User.Email,
	})
	if err != nil {
		return &helper.APIResponse{

			Code:      http.StatusBadRequest,
			Status:    helper.APIStatus.Invalid,
			Message:   "User is not existed",
			ErrorCode: string(enum.ErrorCodeNotExisted.User),
		}
	}

	user := existUserResp.([]*model.User)[0]
	if !helper.VerifyPassword(user.HashedPassword, input.User.Password) {
		return &helper.APIResponse{
			Code:      http.StatusBadRequest,
			Status:    helper.APIStatus.Invalid,
			Message:   "Password is not matched",
			ErrorCode: string(enum.ErrorCodeInvalid.Password),
		}
	}

	tokenMap, err := helper.GenerateJWT(user.Email, user.Username, user.UserID)
	if err != nil {
		return &helper.APIResponse{
			Code:    http.StatusBadRequest,
			Status:  helper.APIStatus.Error,
			Message: "Cannot generate token pair",
		}
	}

	user.AccessToken = tokenMap["accessToken"]

	_, err = r.UpdateOne(model.User{
		ID: user.ID,
	}, model.User{
		RefreshToken: tokenMap["refreshToken"],
	})
	if err != nil {
		return &helper.APIResponse{
			Code:      http.StatusBadRequest,
			Status:    helper.APIStatus.Error,
			Message:   "Fail to create user refresh token " + err.Error(),
			ErrorCode: string(enum.ErrorCodeDatabaseOperation.Update),
		}
	}

	return &helper.APIResponse{
		Code:    http.StatusOK,
		Status:  helper.APIStatus.Ok,
		Message: "Login user successfully",
		Data:    user,
	}
}
