package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Note: These handlers assume global variables 'store', 'handlerLatency', 'errorCounter', 'taskCounter'
// and functions 'logWithTrace', 'handleError', 'contains' are accessible within the 'main' package.

func addHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		handlerLatency.Record(r.Context(), float64(duration), metric.WithAttributes(attribute.String("handler", "add")))
	}()
	ctx := r.Context()
	tr := otel.Tracer("todo-service")
	ctx, span := tr.Start(ctx, "addHandler")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.user_agent", r.UserAgent()),
		attribute.String("http.client_ip", r.RemoteAddr),
	)

	var todo ToDo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		handleError(ctx, w, http.StatusBadRequest, "Invalid JSON", err)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "add")))
		return
	}
	span.SetAttributes(attribute.String("todo.text", todo.Text))

	// Using the global store instance
	added := store.Add(todo)
	taskCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("source", "http")))

	logWithTrace(ctx).Str("event", "task_added").Int("todo_id", added.ID).Str("todo_text", added.Text).Msg("Added task")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // Use 201 Created for successful additions
	json.NewEncoder(w).Encode(added)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		handlerLatency.Record(r.Context(), float64(duration), metric.WithAttributes(attribute.String("handler", "list")))
	}()
	ctx := r.Context()
	tr := otel.Tracer("todo-service")
	ctx, span := tr.Start(ctx, "listHandler")
	defer span.End()

	// Using the global store instance
	todos := store.List()

	logWithTrace(ctx).Str("event", "list_tasks").Int("count", len(todos)).Msg("Listed tasks")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		handlerLatency.Record(r.Context(), float64(duration), metric.WithAttributes(attribute.String("handler", "delete")))
	}()
	ctx := r.Context()
	tr := otel.Tracer("todo-service")
	ctx, span := tr.Start(ctx, "deleteHandler")
	defer span.End()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		span.SetAttributes(attribute.String("todo.delete.error", "invalid id"))
		handleError(ctx, w, http.StatusBadRequest, "Invalid ID", err)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "delete")))
		return
	}
	span.SetAttributes(attribute.Int("todo.id", id))

	// Using the global store instance
	deleted := store.Delete(id)

	if deleted {
		logWithTrace(ctx).Str("event", "delete_task").Int("todo_id", id).Msg("Deleted task")
		w.WriteHeader(http.StatusNoContent)
	} else {
		span.SetAttributes(attribute.String("todo.delete.status", "not found"))
		handleError(ctx, w, http.StatusNotFound, "ToDo not found", nil) // Use handleError
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "delete")))
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		handlerLatency.Record(r.Context(), float64(duration), metric.WithAttributes(attribute.String("handler", "update")))
	}()

	ctx := r.Context()
	tr := otel.Tracer("todo-service")
	ctx, span := tr.Start(ctx, "updateHandler")
	defer span.End()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handleError(ctx, w, http.StatusBadRequest, "Invalid ID", err)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "update")))
		return
	}

	var todo ToDo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		handleError(ctx, w, http.StatusBadRequest, "Invalid request body", err)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "update")))
		return
	}

	// Using the global store instance
	updated, exists := store.Update(id, todo.Text)

	if !exists {
		handleError(ctx, w, http.StatusNotFound, "ToDo not found", nil)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "update")))
		return
	}

	span.SetAttributes(attribute.String("todo.text", updated.Text), attribute.Int("todo.id", updated.ID))
	logWithTrace(ctx).Str("event", "update_task").Int("todo_id", updated.ID).Str("todo_text", updated.Text).Msg("Updated task")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		handlerLatency.Record(r.Context(), float64(duration), metric.WithAttributes(attribute.String("handler", "get")))
	}()

	ctx := r.Context()
	tr := otel.Tracer("todo-service")
	ctx, span := tr.Start(ctx, "getHandler")
	defer span.End()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handleError(ctx, w, http.StatusBadRequest, "Invalid ID format", err)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "get")))
		return
	}
	span.SetAttributes(attribute.Int("todo.id", id))

	// Using the global store instance
	todo, exists := store.Get(id)

	if !exists {
		handleError(ctx, w, http.StatusNotFound, "ToDo not found", nil)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "get")))
		return
	}

	span.SetAttributes(attribute.String("todo.text", todo.Text))
	logWithTrace(ctx).Str("event", "get_task").Int("todo_id", todo.ID).Str("todo_text", todo.Text).Msg("Retrieved task")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func completeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		handlerLatency.Record(r.Context(), float64(duration), metric.WithAttributes(attribute.String("handler", "complete")))
	}()

	ctx := r.Context()
	tr := otel.Tracer("todo-service")
	ctx, span := tr.Start(ctx, "completeHandler")
	defer span.End()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handleError(ctx, w, http.StatusBadRequest, "Invalid ID", err)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "complete")))
		return
	}
	span.SetAttributes(attribute.Int("todo.id", id))

	// Using the global store instance
	completedTodo, exists := store.Complete(id)

	if !exists {
		handleError(ctx, w, http.StatusNotFound, "ToDo not found", nil)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "complete")))
		return
	}

	span.SetAttributes(attribute.String("todo.text", completedTodo.Text), attribute.Bool("todo.completed", completedTodo.Completed))
	logWithTrace(ctx).Str("event", "complete_task").Int("todo_id", completedTodo.ID).Msg("Completed task")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedTodo)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		handlerLatency.Record(r.Context(), float64(duration), metric.WithAttributes(attribute.String("handler", "search")))
	}()

	ctx := r.Context()
	tr := otel.Tracer("todo-service")
	ctx, span := tr.Start(ctx, "searchHandler")
	defer span.End()

	query := r.URL.Query().Get("q")
	if query == "" {
		handleError(ctx, w, http.StatusBadRequest, "Query parameter 'q' is required", nil)
		errorCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", "search")))
		return
	}
	span.SetAttributes(attribute.String("search.query", query))

	// Using the global store instance
	results := store.Search(query)

	span.SetAttributes(attribute.Int("search.results", len(results)))
	logWithTrace(ctx).Str("event", "search_tasks").Str("query", query).Int("count", len(results)).Msg("Searched tasks")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
