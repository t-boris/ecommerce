package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/t-boris/ecommerce/database"
	"github.com/t-boris/ecommerce/models"
	generate "github.com/t-boris/ecommerce/tokens"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

var UserCollection = database.UserData(database.Client, "Users")
var ProductCollection = database.UserData(database.Client, "Products")
var Validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	msg := ""

	if err != nil {
		msg = "wrong credentials"
		valid = false
	}
	return valid, msg
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := Validate.Struct(user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
			return
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this phone number is already in user"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()
		token, refreshToken, _ := generate.TokenGenerator(*user.Email, *user.FirstName, *user.LastName, user.UserID)
		user.Token = &token
		user.RefreshToken = &refreshToken
		user.UserCart = make([]models.ProductUser, 0)
		user.AddressDetails = make([]models.Address, 0)
		user.OrderStatus = make([]models.Order, 0)

		_, insertErr := UserCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "the user was not created"})
			return
		}

		c.JSON(http.StatusCreated, "successfully signed in")
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		var foundUser models.User

		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user was not found"})
			return
		}

		PasswordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()

		if !PasswordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		token, refreshToken, _ := generate.TokenGenerator(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, foundUser.UserID)
		defer cancel()

		generate.UpdateAllTokens(token, refreshToken, foundUser.UserID)

		c.JSON(http.StatusFound, foundUser)
	}
}

func ProductViewAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var product models.Product
		if err := c.BindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		product.ID = primitive.NewObjectID()
		_, err := ProductCollection.InsertOne(ctx, product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not inserted"})
			return
		}
		c.JSON(http.StatusOK, "success")
	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productList []models.Product
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := ProductCollection.Find(ctx, bson.D{})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something went wrong, please try after some time")
			return
		}

		err = cursor.All(ctx, &productList)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		defer func(cursor *mongo.Cursor, ctx context.Context) {
			err := cursor.Close(ctx)
			if err != nil {
				log.Println(err)
			}
		}(cursor, ctx)

		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "invalid")
			return
		}
		defer cancel()
		c.IndentedJSON(200, productList)
	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchProducts []models.Product
		queryParam := c.Query("name")

		if queryParam == "" {
			log.Println("query is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid search index"})
			c.Abort()
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		searchQueryDB, err := ProductCollection.Find(ctx, bson.M{"productName": bson.M{"$regex": queryParam}})

		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "something went wrong while fetching the data")
			return
		}

		err = searchQueryDB.All(ctx, &searchProducts)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "invalid")
			return
		}
		defer func(searchQueryDB *mongo.Cursor, ctx context.Context) {
			err := searchQueryDB.Close(ctx)
			if err != nil {

			}
		}(searchQueryDB, ctx)

		if err := searchQueryDB.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "invalid request")
			return
		}

		defer cancel()
		c.IndentedJSON(200, searchProducts)
	}
}
