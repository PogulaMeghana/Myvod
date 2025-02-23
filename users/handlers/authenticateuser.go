package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/vod/users/dbiface"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func authenticateUser(ctx context.Context, reqUser User, collection dbiface.MongoCollectionAPI) (User, *echo.HTTPError) {
	var storedUser User //user in db
	// check whether the user exists or not
	// Retrieve the stored user from the database based on the email.
	// This assumes that your dbiface.MongoCollectionAPI includes a FindOne method.
	result := collection.FindOne(ctx, bson.M{"email": reqUser.Email})
	err := result.Decode(&storedUser)
	if err != nil {
		// If no user is found, or another error occurs, return an unauthorized error.
		// You can check for mongo.ErrNoDocuments if you want to differentiate.
		log.Error("Error retrieving user:", err)
		return storedUser, echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	// Validate the password.
	// Compare the hashed password from storedUser with the plaintext password provided.
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(reqUser.Password)); err != nil {
		// Passwords do not match.
		return storedUser, echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	// If everything is valid, return the stored user.
	return storedUser, nil

}

// AuthnUser authenticates a user
func (h *UsersHandler) AuthnUser(c echo.Context) error {
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
			errorMessage{Message: "Unable to validate request payload"})
	}
	user, httpError := authenticateUser(context.Background(), user, h.Col)
	if httpError != nil {
		return c.JSON(httpError.Code, httpError.Message)
	}
	token, err := user.createToken()
	if err != nil {
		log.Errorf("Unable to generate the token.")
		return c.JSON(http.StatusInternalServerError,
			errorMessage{Message: "Unable to generate the token"})
	}
	c.Response().Header().Set("Authorization", "Bearer "+token)
	return c.JSON(http.StatusOK, User{Email: user.Email})
}
