package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	FirstName      *string            `json:"firstName" validate:"required,min=2,max=30"`
	LastName       *string            `json:"lastName" validate:"required,min=2,max=30"`
	Password       *string            `json:"password" validate:"required,min=6"`
	Email          *string            `json:"email" validate:"email, required"`
	Phone          *string            `json:"phone" validate:"required"`
	Token          *string            `json:"tokens"`
	RefreshToken   *string            `json:"refreshToken"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	UserID         string             `json:"userId"`
	UserCart       []ProductUser      `json:"userCart" bson:"userCart"`
	AddressDetails []Address          `json:"address" bson:"address"`
	OrderStatus    []Order            `json:"orders" bson:"orders"`
}

type Product struct {
	ID     primitive.ObjectID `json:"_id" bson:"_id"`
	Name   *string            `json:"productName"`
	Price  uint64             `json:"price"`
	Rating uint8              `json:"rating"`
	Image  *string            `json:"image"`
}

type ProductUser struct {
	ID     primitive.ObjectID `json:"_id" bson:"_id"`
	Name   *string            `json:"productName" bson:"productName"`
	Price  uint64             `json:"price" bson:"price"`
	Rating uint8              `json:"rating" bson:"rating"`
	Image  *string            `json:"image" bson:"image"`
}

type Address struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id"`
	House   *string            `json:"houseName" bson:"houseName"`
	Street  *string            `json:"streetName" bson:"streetName"`
	City    *string            `json:"cityName" bson:"cityName"`
	ZipCode *string            `json:"zipCode" bson:"zipCode"`
}
type Order struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	Cart          []ProductUser      `json:"orderList" bson:"orderList"`
	OrderedAt     time.Time          `json:"orderedAt" bson:"orderedAt"`
	Price         uint64             `json:"totalPrice" bson:"totalPrice"`
	Discount      int64              `json:"discount" bson:"discount"`
	PaymentMethod Payment            `json:"paymentMethod" bson:"paymentMethod"`
}

type Payment struct {
	Digital bool
	COD     bool
}
