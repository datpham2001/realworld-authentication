package model

import (
	"realworld-authentication/model/enum"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID              *primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	CreatedTime     *time.Time          `json:"createdTime,omitempty" bson:"created_time,omitempty"`
	LastUpdatedTime *time.Time          `json:"lastUpdatedTime,omitempty" bson:"last_updated_time,omitempty"`

	UserID         string             `json:"userId,omitempty" bson:"user_id,omitempty"`
	Email          string             `json:"email,omitempty" bson:"email,omitempty"`
	Username       string             `json:"username,omitempty" bson:"username,omitempty"`
	HashedPassword string             `json:"hashedPassword,omitempty" bson:"hashed_password,omitempty"`
	Role           enum.UserRoleValue `json:"role,omitempty" bson:"role,omitempty"`

	ComplexQuery []*bson.M `json:"-" bson:"$and,omitempty"`
}
