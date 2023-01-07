package tokens

import (
	"context"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/t-boris/ecommerce/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	UID       string
	jwt.StandardClaims
}

var userData = database.UserData(database.Client, "Users")
var secretKey = os.Getenv("SECRET_KEY")

func TokenGenerator(email string, firstname string, lastname string, uid string) (string, string, error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstname,
		LastName:  lastname,
		UID:       uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS384, refreshClaims).SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "the tokens is invalid"
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "tokens is already expired"
		return
	}
	return claims, msg
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{Key: "tokens", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refreshToken", Value: signedRefreshToken})
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updatedAt", Value: updatedAt})

	upsert := true

	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := userData.UpdateOne(ctx, filter, bson.D{
		{Key: "$set", Value: updateObj},
	}, &opt)
	if err != nil {
		log.Panic(err)
		return
	}

}
