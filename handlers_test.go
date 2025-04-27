package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Setup common test resources if needed, e.g., a mock store or initializing globals for tests
func setupTest() {
	// Initialize global store for testing purposes
	// WARNING: This uses global state, which is not ideal for parallel tests.
	// Consider using dependency injection and mocks for better test isolation.
	store = NewStore()
	// Initialize metrics/tracer stubs if handlers depend on them directly and they aren't mocked
	// initMetrics() // Might start HTTP server, be careful
	// initTracer()() // Might try to connect to collector
}

func TestGetHandler_InvalidID(t *testing.T) {
	setupTest() // Ensure store is initialized

	req, _ := http.NewRequest("GET", "/get?id=invalid", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getHandler) // Assuming getHandler is accessible

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestGetHandler_NotFound(t *testing.T) {
	setupTest() // Ensure store is initialized and empty

	req, _ := http.NewRequest("GET", "/get?id=999", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getHandler) // Assuming getHandler is accessible

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// Add more tests for other handlers (add, list, delete, update, complete, search)
// Example for AddHandler:
/*
func TestAddHandler(t *testing.T) {
    setupTest()

    payload := `{"text": "Test Add"}`
    req, _ := http.NewRequest("POST", "/add", strings.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(addHandler)

    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusCreated { // Expect 201 Created
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
    }

    // Check response body
    var addedTodo ToDo
    if err := json.NewDecoder(rr.Body).Decode(&addedTodo); err != nil {
        t.Fatalf("Could not decode response: %v", err)
    }
    if addedTodo.Text != "Test Add" {
        t.Errorf("handler returned wrong text: got %v want %v", addedTodo.Text, "Test Add")
    }
    if addedTodo.ID == 0 {
        t.Errorf("handler returned zero ID")
    }
}
*/
