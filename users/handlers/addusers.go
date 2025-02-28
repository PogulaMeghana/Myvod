package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/vod/users/dbiface"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// hashPassword hashes the provided password using bcrypt.
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func insertUser(ctx context.Context, user User, collection dbiface.MongoCollectionAPI) (User, *echo.HTTPError) {
	// We'll return the inserted user
	var insertedUser User

	// Check separately if the user email or organization name already exists
	emailCount, err := collection.CountDocuments(ctx, bson.M{"user_name": user.Email})
	if err != nil {
		log.Error("Error checking existing user email:", err)
		return insertedUser, echo.NewHTTPError(http.StatusInternalServerError,
			map[string]string{"message": "Database error while checking user email"})
	}

	orgCount, err := collection.CountDocuments(ctx, bson.M{"organization_name": user.OrganizationName})
	if err != nil {
		log.Error("Error checking existing organization name:", err)
		return insertedUser, echo.NewHTTPError(http.StatusInternalServerError,
			map[string]string{"message": "Database error while checking organization name"})
	}

	// Return specific conflict errors
	if emailCount > 0 {
		log.Warn("User email already exists:", user.Email)
		return insertedUser, echo.NewHTTPError(http.StatusConflict,
			map[string]string{"message": "User with this email already exists"})
	}

	if orgCount > 0 {
		log.Warn("Organization already exists:", user.OrganizationName)
		return insertedUser, echo.NewHTTPError(http.StatusConflict,
			map[string]string{"message": "Organization name already exists"})
	}

	// Hash the user's password before storing
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		log.Error("Error hashing password:", err)
		return insertedUser, echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}
	user.Password = hashedPassword

	// Set user creation and update timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsAdmin = true
	// Insert user into MongoDB
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Error("Error inserting user into database:", err)
		return insertedUser, echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}
	user.Id = result.InsertedID.(primitive.ObjectID)
	// Return the inserted user (or you could return user)
	return user, nil
}

// CreateUser creates a user
func (h *UsersHandler) CreateUser(c echo.Context) error {
	var user User
	c.Echo().Validator = &userValidator{validator: v}
	if err := c.Bind(&user); err != nil {
		log.Errorf("Unable to bind to user struct.", err)
		return c.JSON(http.StatusUnprocessableEntity,
			errorMessage{Message: "Unable to parse the request payload."})
	}
	if err := c.Validate(user); err != nil {
		log.Errorf("Unable to validate the requested body.")
		return c.JSON(http.StatusBadRequest,
			errorMessage{Message: "Unable to validate request body"})
	}
	resUser, httpError := insertUser(context.Background(), user, h.Col)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}
	token, err := user.createToken()
	if err != nil {
		log.Errorf("Unable to generate the token.")
		return echo.NewHTTPError(http.StatusInternalServerError,
			errorMessage{Message: "Unable to generate the token"})
	}
	c.Response().Header().Set("Authorization", "Bearer "+token)
	// ✅ SUCCESS RESPONSE WITH MESSAGE
	responseUser := map[string]interface{}{
		"message": "User created successfully",
		"user_id": resUser.Id.Hex(), // Convert ObjectID to string

	}
	return c.JSON(http.StatusCreated, responseUser)
}
