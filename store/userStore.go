package store

import (
	"context"
	"errors"
	"event-horizon/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

/************************** I DID THE FOLLOWING THINGS IN THIS FILE: *******************************************


1. Created UserStore struct to manage user-related database operations.

2. Implemented NewUserStore constructor to initialize UserStore with MongoDB collection.

3. Developed CreateUser method to add new users, including password hashing and duplicate email checks.

4. Created GetUserByID method to retrieve a user by their ID.

5. Added FindUserByEmail method to find a user by their email address.

6. Implemented VerifyPassword method to compare a plain password with the hashed password stored in the database.


************************************************************************************************************/

type UserStore struct {
	collection *mongo.Collection
}

func NewUserStore(db *mongo.Database) *UserStore {
	return &UserStore{
		collection: db.Collection("Users"),
	}
}

func (s *UserStore) CreateUser(ctx context.Context, user *models.User) error {
	//? HASH THE PASSWORD BEFORE STORING
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	//? COMPARE DUPLICATE EMAILS
	existingUser := models.User{}
	err = s.collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if existingUser.ID != bson.NilObjectID {
		return errors.New("email already exists")
	}

	//? Set creation timestamp
	user.CreatedAt = time.Now()

	//? INSERT THE USER
	result, err := s.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(bson.ObjectID)
	return nil
}

// GetUserByID retrieves a user by their ID
func (s *UserStore) GetUserByID(ctx context.Context, userID bson.ObjectID) (*models.User, error) {
	var user models.User
	err := s.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByEmail FINDS A USER BY THEIR EMAIL
func (s *UserStore) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// VerifyPassword COMPARES A PLAIN PASSWORD WITH THE HASHED PASSWORD
func (s *UserStore) VerifyPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
