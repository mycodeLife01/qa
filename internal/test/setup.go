package test

import (
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/mycodeLife01/qa/internal/handler"
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/pkg/client"
	"github.com/mycodeLife01/qa/internal/service/impl"
	"gorm.io/gorm"
)

// SetupTestDB initializes an in-memory SQLite database and migrates models
func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate schemas
	err = db.AutoMigrate(
		&model.User{},
		&model.File{},
		&model.UserFile{},
		&model.Thread{},
		&model.Message{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// SetupRouter initializes the Gin engine with all handlers wired up
// It mocks the PythonClient using the provided baseURL
func SetupRouter(db *gorm.DB, pythonAgentURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Services
	// Create a real PythonClient
	pythonClient := client.NewPythonServiceClient()

	// Initialize Services (Passing nil for COS/Task clients as they shouldn't be used in the chat flow we are testing)
	fileService := impl.NewFileService(db, nil, nil)
	aiService := impl.NewAiService(db, pythonClient, nil, nil)
	threadService := impl.NewThreadService(db)
	messageService := impl.NewMessageService(db)

	// Initialize Handler
	aiHandler := handler.NewAiHandler(aiService, threadService, messageService, fileService)

	// Setup Route
	// We need a way to mock the Auth middleware or bypass it.
	// For testing, we can inject a middleware that sets a specific UserID.
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(1)) // Mock user ID 1
		c.Next()
	})

	r.POST("/ai/ask", aiHandler.Ask)

	return r
}
