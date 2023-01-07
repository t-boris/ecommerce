package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
	"github.com/t-boris/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")

		if userID == "" {
			c.Header("Context-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid code"})
			c.Abort()
			return
		}

		hexUserID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
		}

		var addresses models.Address
		addresses.ID = primitive.NewObjectID()

		if err = c.BindJSON(&addresses); err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		matchFilter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: hexUserID}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}
		group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$ID"}, {Key: "count", Value: bson.D{primitive.E{
			Key: "$sum", Value: 1,
		}}}}}}

		pointCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{matchFilter, unwind, group})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
		}

		var addressInfo []bson.M
		if err := pointCursor.All(ctx, &addressInfo); err != nil {
			panic(err)
		}

		var size int32
		for _, addressNo := range addressInfo {
			count := addressNo["count"]
			size = count.(int32)
		}
		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: addresses}}
			update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			c.IndentedJSON(http.StatusBadRequest, "not allowed")
		}
		defer cancel()
		ctx.Done()
	}
}

func EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")
		if userID == "" {
			c.Header("Context-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid"})
			c.Abort()
			return
		}

		hexUserID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var editAddress models.Address
		if err := c.BindJSON(&editAddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}
		filter := bson.D{primitive.E{Key: "_id", Value: hexUserID}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{
			Key: "address.0.houseName", Value: editAddress.House}, {
			Key: "address.0.streetName", Value: editAddress.Street}, {
			Key: "address.0.cityName", Value: editAddress.City}, {
			Key: "address.0.zipCode", Value: editAddress.ZipCode}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "successfully updated home address")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")
		if userID == "" {
			c.Header("Context-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid"})
			c.Abort()
			return
		}

		hexUserID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
			return
		}
		var editAddress models.Address
		if err := c.BindJSON(&editAddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: hexUserID}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{
			Key: "address.1.houseName", Value: editAddress.House}, {
			Key: "address.1.streetName", Value: editAddress.Street}, {
			Key: "address.1.cityName", Value: editAddress.City}, {
			Key: "address.1.zipCode", Value: editAddress.ZipCode}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "successfully updated work address")
	}
}

func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("id")

		if userID == "" {
			c.Header("Context-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid search index"})
			c.Abort()
			return
		}

		addresses := make([]models.Address, 0)
		hexUserID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: hexUserID}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "wrong command")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "successfully deleted")
	}
}
