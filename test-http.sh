#!/bin/bash

# Validate and list all ToDos
echo "Listing all ToDos:"
response=$(curl -s http://localhost:8080/list)
if ! echo "$response" | jq empty 2>/dev/null; then
  echo "Error: Invalid JSON response from API"
  exit 1
fi
echo "$response" | jq
echo ""

# Add new ToDos
echo "Adding new ToDos:"
curl -s -X POST localhost:8080/add -d '{"text":"Write code"}' -H "Content-Type: application/json" | jq
curl -s -X POST localhost:8080/add -d '{"text":"Test code"}' -H "Content-Type: application/json" | jq
curl -s -X POST localhost:8080/add -d '{"text":"Run code"}' -H "Content-Type: application/json" | jq
curl -s -X POST localhost:8080/add -d '{"text":"Ship code"}' -H "Content-Type: application/json" | jq
curl -s -X POST localhost:8080/add -d '{"text":"Shit histograms"}' -H "Content-Type: application/json" | jq
curl -s -X POST localhost:8080/add -d '{"text":"Build something"}' -H "Content-Type: application/json" | jq
echo ""

# List all ToDos again
echo "Listing all ToDos after adding:"
response=$(curl -s http://localhost:8080/list)
if ! echo "$response" | jq empty 2>/dev/null; then
  echo "Error: Invalid JSON response from API"
  exit 1
fi
echo "$response" | jq
echo ""

# Find and delete ToDos containing the word "Shit"
echo "Finding and deleting ToDos containing the word 'Shit':"
echo "$response" | jq -c '.[]' | while read -r todo; do
  id=$(echo "$todo" | jq -r '.id')
  text=$(echo "$todo" | jq -r '.text')
  if [[ $text == *"Shit"* ]]; then
    echo "Deleting ToDo with ID $id and text '$text'"
    curl -s -X DELETE "http://localhost:8080/delete?id=$id" | jq
  fi
done
echo ""

# Update and mark complete all ToDos containing the word "Write"
echo "Updating and marking complete all ToDos containing the word 'Write':"
echo "$response" | jq -c '.[]' | while read -r todo; do
  id=$(echo "$todo" | jq -r '.id')
  text=$(echo "$todo" | jq -r '.text')
  if [[ $text == *"Write"* ]]; then
    echo "Updating ToDo with ID $id and text '$text' to mark as complete"
    curl -s -X PUT "http://localhost:8080/update?id=$id" \
      -d "{\"text\":\"$text\"}" \
      -H "Content-Type: application/json" | jq
    curl -s -X POST "http://localhost:8080/complete?id=$id" | jq
  fi
done
echo ""

# Update ToDos containing "Build" to say "Build and rebuild"
echo "Updating ToDos containing 'Build' to say 'Build and rebuild':"
echo "$response" | jq -c '.[]' | while read -r todo; do
  id=$(echo "$todo" | jq -r '.id')
  text=$(echo "$todo" | jq -r '.text')
  if [[ $text == *"Build"* ]]; then
    new_text="Build and rebuild"
    echo "Updating ToDo with ID $id and text '$text' to '$new_text'"
    curl -s -X PUT "http://localhost:8080/update?id=$id" \
      -d "{\"text\":\"$new_text\"}" \
      -H "Content-Type: application/json" | jq
  fi
done
echo ""

# List all ToDos after updates
echo "Listing all ToDos after updates:"
response=$(curl -s http://localhost:8080/list)
if ! echo "$response" | jq empty 2>/dev/null; then
  echo "Error: Invalid JSON response from API"
  exit 1
fi
echo "$response" | jq
echo ""

# Test the "search" endpoint
echo "Testing the 'search' endpoint:"
search_query="Write"
response=$(curl -s "http://localhost:8080/search?q=$search_query")
if ! echo "$response" | jq empty 2>/dev/null; then
  echo "Error: Invalid JSON response from 'search' endpoint"
  exit 1
fi
echo "Search results for query '$search_query':"
echo "$response" | jq
echo ""

# Test the "get" endpoint
echo "Testing the 'get' endpoint:"
todo_id=1
response=$(curl -s "http://localhost:8080/get?id=$todo_id")
if ! echo "$response" | jq empty 2>/dev/null; then
  echo "Error: Invalid JSON response from 'get' endpoint"
  exit 1
fi
echo "Details for ToDo with ID $todo_id:"
echo "$response" | jq
echo ""

# Test invalid ID for "get" endpoint
echo "Testing invalid ID for 'get' endpoint:"
response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/get?id=invalid")
if [ "$response" -ne 400 ]; then
  echo "Test failed: Expected 400, got $response"
else
  echo "Test passed: Invalid ID for 'get' endpoint"
fi

# Test non-existent ID for "get" endpoint
echo "Testing non-existent ID for 'get' endpoint:"
response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/get?id=999")
if [ "$response" -ne 404 ]; then
  echo "Test failed: Expected 404, got $response"
else
  echo "Test passed: Non-existent ID for 'get' endpoint"
fi