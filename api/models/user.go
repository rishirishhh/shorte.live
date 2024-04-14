package models

import (
	"context"
	"fmt"
	"time"

	"github.com/ivinayakg/shorte.live/api/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

func CreateUser(email string, name string, picture string) (*database.User, error) {
	createdAt := database.UnixTime(time.Now().Unix())
	user := database.User{Name: name, Email: email, Picture: picture, CreatedAt: createdAt}
	ctx := context.TODO()

	res, err := database.CurrentDb.User.InsertOne(ctx, user)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	user.ID = res.InsertedID.(primitive.ObjectID)
	fmt.Printf("User created with id %v\n", user.ID)

	return &user, nil
}

func GetUser(email string) (*database.User, error) {
	user := new(database.User)

	ctx := context.TODO()
	userFilter := bson.M{"email": email}

	err := database.CurrentDb.User.FindOne(ctx, userFilter).Decode(user)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Printf("User found with id %v\n", user.ID)
	return user, nil
}
