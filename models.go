package main

// ToDo represents a task item.
type ToDo struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// CompletedToDo represents a completed task item.
type CompletedToDo struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}
