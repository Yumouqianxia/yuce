package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ExampleSuccess demonstrates how to send a successful response.
func ExampleSuccess() {
	// This would typically be inside a Gin handler
	var c *gin.Context

	// Example user data
	user := map[string]interface{}{
		"id":       1,
		"username": "john_doe",
		"email":    "john@example.com",
	}

	// Send success response
	Success(c, http.StatusOK, "User retrieved successfully", user)

	// This produces JSON output:
	// {
	//   "success": true,
	//   "message": "User retrieved successfully",
	//   "data": {
	//     "id": 1,
	//     "username": "john_doe",
	//     "email": "john@example.com"
	//   }
	// }
}

// ExampleError demonstrates how to send an error response.
func ExampleError() {
	var c *gin.Context

	// Send error response
	Error(c, http.StatusBadRequest, "Invalid input", "Email format is invalid")

	// This produces JSON output:
	// {
	//   "success": false,
	//   "message": "Invalid input",
	//   "error": {
	//     "code": 400,
	//     "message": "Invalid input",
	//     "details": "Email format is invalid"
	//   }
	// }
}

// ExampleUserHandler demonstrates a complete user handler using response utilities.
func ExampleUserHandler() {
	// Example Gin handler function
	getUserHandler := func(c *gin.Context) {
		userID := c.Param("id")

		// Validate user ID
		if userID == "" {
			BadRequest(c, "User ID is required")
			return
		}

		// Simulate user lookup
		user, err := findUserByID(userID)
		if err != nil {
			if err == ErrUserNotFound {
				NotFound(c, "User")
				return
			}
			InternalError(c, "Failed to retrieve user")
			return
		}

		// Return successful response
		OK(c, "User retrieved successfully", user)
	}

	// Example usage in router setup
	var router *gin.Engine
	router.GET("/users/:id", getUserHandler)
}

// ExampleCreateUserHandler demonstrates handling user creation with validation.
func ExampleCreateUserHandler() {
	createUserHandler := func(c *gin.Context) {
		var req CreateUserRequest

		// Bind and validate request
		if err := c.ShouldBindJSON(&req); err != nil {
			ValidationError(c, err.Error())
			return
		}

		// Check if user already exists
		if userExists(req.Email) {
			Conflict(c, "User with this email already exists")
			return
		}

		// Create user
		user, err := createUser(&req)
		if err != nil {
			InternalError(c, "Failed to create user")
			return
		}

		// Return created response
		Created(c, "User created successfully", user)
	}

	var router *gin.Engine
	router.POST("/users", createUserHandler)
}

// ExampleAuthMiddleware demonstrates using response utilities in middleware.
func ExampleAuthMiddleware() {
	authMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			token := c.GetHeader("Authorization")

			if token == "" {
				Unauthorized(c, "Authorization token is required")
				c.Abort()
				return
			}

			// Validate token
			user, err := validateToken(token)
			if err != nil {
				Unauthorized(c, "Invalid or expired token")
				c.Abort()
				return
			}

			// Check permissions
			if !user.HasPermission("read_users") {
				Forbidden(c, "Insufficient permissions")
				c.Abort()
				return
			}

			// Set user in context
			c.Set("user", user)
			c.Next()
		}
	}

	var router *gin.Engine
	router.Use(authMiddleware())
}

// Example types and functions for demonstration
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (u *User) HasPermission(permission string) bool {
	// Mock permission check
	return true
}

var ErrUserNotFound = fmt.Errorf("user not found")

func findUserByID(id string) (*User, error) {
	// Mock user lookup
	if id == "999" {
		return nil, ErrUserNotFound
	}
	return &User{ID: 1, Username: "john", Email: "john@example.com"}, nil
}

func userExists(email string) bool {
	// Mock existence check
	return email == "existing@example.com"
}

func createUser(req *CreateUserRequest) (*User, error) {
	// Mock user creation
	return &User{
		ID:       1,
		Username: req.Username,
		Email:    req.Email,
	}, nil
}

func validateToken(token string) (*User, error) {
	// Mock token validation
	if token == "Bearer invalid" {
		return nil, fmt.Errorf("invalid token")
	}
	return &User{ID: 1, Username: "john"}, nil
}
