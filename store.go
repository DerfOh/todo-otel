package main

import "sync"

// Store manages the ToDo items.
type Store struct {
	sync.Mutex
	data  map[int]ToDo
	count int
}

// NewStore creates a new Store.
func NewStore() *Store {
	return &Store{
		data: make(map[int]ToDo),
	}
}

// Add adds a new ToDo item to the store.
func (s *Store) Add(todo ToDo) ToDo {
	s.Lock()
	defer s.Unlock()
	s.count++
	todo.ID = s.count
	s.data[todo.ID] = todo
	return todo
}

// Get retrieves a ToDo item by ID.
func (s *Store) Get(id int) (ToDo, bool) {
	s.Lock()
	defer s.Unlock()
	todo, exists := s.data[id]
	return todo, exists
}

// List returns all ToDo items.
func (s *Store) List() []ToDo {
	s.Lock()
	defer s.Unlock()
	list := make([]ToDo, 0, len(s.data))
	for _, todo := range s.data {
		list = append(list, todo)
	}
	return list
}

// Delete removes a ToDo item by ID. Returns true if deleted, false otherwise.
func (s *Store) Delete(id int) bool {
	s.Lock()
	defer s.Unlock()
	_, exists := s.data[id]
	if exists {
		delete(s.data, id)
	}
	return exists
}

// Update modifies the text of an existing ToDo item. Returns the updated ToDo and true if found, otherwise empty ToDo and false.
func (s *Store) Update(id int, text string) (ToDo, bool) {
	s.Lock()
	defer s.Unlock()
	todo, exists := s.data[id]
	if !exists {
		return ToDo{}, false
	}
	todo.Text = text
	s.data[id] = todo // Update the map
	return todo, true
}

// Complete marks a ToDo item as completed (conceptually, by updating its representation if needed).
// Note: The current implementation in main.go updates the store with a ToDo struct,
// not a CompletedToDo struct. This logic might need refinement depending on requirements.
func (s *Store) Complete(id int) (CompletedToDo, bool) {
	s.Lock()
	defer s.Unlock()
	todo, exists := s.data[id]
	if !exists {
		return CompletedToDo{}, false
	}
	completedTodo := CompletedToDo{ID: todo.ID, Text: todo.Text, Completed: true}
	// Update the store if necessary, or handle completion status differently.
	// The original code overwrites the existing ToDo, which might not be intended.
	// For now, just return the completed representation.
	return completedTodo, true
}

// Search finds ToDo items containing the query text.
func (s *Store) Search(query string) []ToDo {
	s.Lock()
	defer s.Unlock()
	results := []ToDo{}
	for _, todo := range s.data {
		if contains(todo.Text, query) { // Assumes contains is in utils.go or main package
			results = append(results, todo)
		}
	}
	return results
}
