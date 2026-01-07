package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// CloseNotifyingRecorder implements http.CloseNotifier, which is used by Gin's SSE implementation
// Note: http.CloseNotifier is deprecated in Go 1.8+, but some frameworks still rely on it or
// check for it. Gin uses context cancellation now, but check httptest limitations.
// Gin's context.Stream uses w.CloseNotify() if available.
type CloseNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closeNotifyChan chan bool
}

func NewCloseNotifyingRecorder() *CloseNotifyingRecorder {
	return &CloseNotifyingRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closeNotifyChan:  make(chan bool, 1),
	}
}

// CloseNotify implements http.CloseNotifier
func (r *CloseNotifyingRecorder) CloseNotify() <-chan bool {
	return r.closeNotifyChan
}

// Flush implements http.Flusher
func (r *CloseNotifyingRecorder) Flush() {
	// httptest.ResponseRecorder already implements Flush if we embed it,
	// but the embedded pointer might need explicit delegation if interface check fails?
	// httptest.ResponseRecorder satisfies http.Flusher.
	r.ResponseRecorder.Flush()
}

func TestAiHandler_Ask_NewThread(t *testing.T) {
	// 1. Setup DB
	db, err := test.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup DB: %v", err)
	}

	// 2. Seed Data
	// User
	user := model.User{
		Model:    gorm.Model{ID: 1},
		Username: "testuser",
		Email:    "test@example.com",
	}
	db.Create(&user)

	// File (Must exist and be SUCCESS)
	file := model.File{
		ContentHash: "hash123",
		ObjectKey:   "key123",
		BucketName:  "bucket",
		FileType:    "txt",
		Status:      "SUCCESS",
	}
	db.Create(&file)

	// 3. Mock Python Agent Server
	// The real handler expects a stream response from Python Agent
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		if r.URL.Path != "/chat" {
			t.Errorf("Expected path /chat, got %s", r.URL.Path)
		}

		// Verify request body contains thread_id and doc_hash
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		if reqBody["doc_content_hash"] != "hash123" {
			t.Errorf("Expected hash123, got %v", reqBody["doc_content_hash"])
		}

		// Simulate SSE Stream response
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Send data chunks
		// Important: use data: prefix
		fmt.Fprintf(w, "data: Hello\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		fmt.Fprintf(w, "data:  World\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}))
	defer ts.Close()

	// Update global config to point to mock server
	config.C.Services.QAAgentURL = ts.URL

	// 4. Setup Router
	r := test.SetupRouter(db, ts.URL)

	// 5. Perform Request (New Thread)
	reqBody := map[string]string{
		"question":          "Hi",
		"file_content_hash": "hash123",
	}
	jsonData, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/ai/ask", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Use custom recorder
	w := NewCloseNotifyingRecorder()

	r.ServeHTTP(w, req)

	// 6. Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify SSE response
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "data:Hello") // Gin SSE implementation might strip space? No, c.SSEvent writes "data:" + message + "\n\n"
	// Wait, c.SSEvent(name, message) -> "event: name\ndata: message\n\n"
	// In our handler: c.SSEvent("message", token)
	// Output:
	// event: message
	// data: Hello
	//
	// event: message
	// data:  World
	//
	assert.Contains(t, responseBody, "data:Hello")
	assert.Contains(t, responseBody, "data: World") // Leading space preserved if string

	// Verify DB Side Effects
	// 1. Thread Created
	var thread model.Thread
	err = db.First(&thread).Error
	assert.NoError(t, err)
	assert.Equal(t, uint(1), thread.UserID)
	assert.Equal(t, file.ID, thread.FileID)
	assert.Equal(t, "Hi", thread.Title)

	// 2. Messages Created
	var messages []model.Message
	err = db.Where("thread_id = ?", thread.ID).Order("id asc").Find(&messages).Error
	assert.NoError(t, err)
	require.Len(t, messages, 2)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "Hi", messages[0].Content)
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, "Hello World", messages[1].Content) // "Hello" + " World"
}

func TestAiHandler_Ask_ExistingThread(t *testing.T) {
	// 1. Setup DB
	db, err := test.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup DB: %v", err)
	}

	// 2. Seed Data
	db.Create(&model.User{Model: gorm.Model{ID: 1}, Username: "testuser"})
	file := model.File{ContentHash: "hash123", Status: "SUCCESS"}
	db.Create(&file)

	thread := model.Thread{
		UUID:   "uuid-123",
		UserID: 1,
		FileID: file.ID,
		Title:  "Old Thread",
	}
	db.Create(&thread)

	// 3. Mock Python Agent
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify thread_id is passed
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Note: The backend logic passes thread UUID to python agent
		if reqBody["thread_id"] != "uuid-123" {
			t.Errorf("Expected thread_id uuid-123, got %v", reqBody["thread_id"])
		}

		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: New Response\n\n")
	}))
	defer ts.Close()
	config.C.Services.QAAgentURL = ts.URL

	// 4. Setup Router
	r := test.SetupRouter(db, ts.URL)

	// 5. Perform Request (Existing Thread)
	reqBody := map[string]string{
		"question":  "Follow up",
		"thread_id": "uuid-123",
		// file_content_hash is optional now if thread exists, but handler might check logic
		"file_content_hash": "hash123", // Providing it to avoid "InvalidParams" fallback error in current logic
	}
	jsonData, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/ai/ask", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := NewCloseNotifyingRecorder()

	r.ServeHTTP(w, req)

	// 6. Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify DB Messages
	var messages []model.Message
	err = db.Where("thread_id = ?", thread.ID).Order("id asc").Find(&messages).Error
	assert.NoError(t, err)
	require.Len(t, messages, 2)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "Follow up", messages[0].Content)
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, "New Response", messages[1].Content)
}
