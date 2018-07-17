package swagger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/marvold/todo/model"
	"github.com/stretchr/testify/assert"
)

func TestAPI(t *testing.T) {
	// This is a more limited set of tests than the full tests at the model
	// layer to avoid a lot of redundancy.  As a next step, we could work to
	// mock out the underlying database here to focus on testing routing code.
	// For now, we just don't worry about generating all possible database
	// errors, instead focusing on routing/parsing.
	router := NewRouter()

	// Dummy list.
	newlist := model.TodoList{
		"d290f1ee-6c54-4b01-90e6-d701748f0851",
		"Home",
		"The list of things that need to be done at home\n",
		[]model.Task{},
	}

	// Add a list, succeeds.
	body, _ := json.Marshal(newlist)
	req := httptest.NewRequest("POST", "http://localhost:8080/aweiker/ToDo/1.0.0/lists", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp := rec.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Add a list w/o a body, fails.
	req = httptest.NewRequest("POST", "http://localhost:8080/aweiker/ToDo/1.0.0/lists", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Add a list w/a malformed payload, fails.
	req = httptest.NewRequest("POST", "http://localhost:8080/aweiker/ToDo/1.0.0/lists", strings.NewReader("This isn't right"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Dummy task.
	newtask := model.Task{"0e2ac84f-f723-4f24-878b-44e63e7ae580", "mow the yard", false}

	// Add a task, succeeds.
	body, _ = json.Marshal(newtask)
	req = httptest.NewRequest("POST", "http://localhost:8080/aweiker/ToDo/1.0.0/list/d290f1ee-6c54-4b01-90e6-d701748f0851/tasks", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Add a task w/o a body, fails.
	req = httptest.NewRequest("POST", "http://localhost:8080/aweiker/ToDo/1.0.0/list/d290f1ee-6c54-4b01-90e6-d701748f0851/tasks", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Add a task w/a malformed payload, fails.
	req = httptest.NewRequest("POST", "http://localhost:8080/aweiker/ToDo/1.0.0/list/d290f1ee-6c54-4b01-90e6-d701748f0851/tasks", strings.NewReader("This isn't right"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Completed state.
	completed := model.CompletedTask{true}

	// Complete a task, succeeds.
	body, _ = json.Marshal(completed)
	req = httptest.NewRequest("POST", "http://localhost:8080/aweiker/ToDo/1.0.0/list/d290f1ee-6c54-4b01-90e6-d701748f0851/task/0e2ac84f-f723-4f24-878b-44e63e7ae580/complete", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Complete a task w/o a body, fails.
	req = httptest.NewRequest("POST", "http://localhost:8080/aweiker/ToDo/1.0.0/list/d290f1ee-6c54-4b01-90e6-d701748f0851/task/0e2ac84f-f723-4f24-878b-44e63e7ae580/complete", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Complete a task w/a malformed payload, fails.
	req = httptest.NewRequest("POST", "http://localhost:8080/aweiker/ToDo/1.0.0/list/d290f1ee-6c54-4b01-90e6-d701748f0851/task/0e2ac84f-f723-4f24-878b-44e63e7ae580/complete", strings.NewReader("This isn't right"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Get list, succeeds.
	req = httptest.NewRequest("GET", "http://localhost:8080/aweiker/ToDo/1.0.0/list/d290f1ee-6c54-4b01-90e6-d701748f0851", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Compare results.
	newtask.Completed = true
	newlist.Tasks = append(newlist.Tasks, newtask)
	resultlist := model.TodoList{}
	err := json.NewDecoder(resp.Body).Decode(&resultlist)
	assert.Nil(t, err)
	assert.Equal(t, newlist, resultlist)

	// Get lists, succeeds.
	req = httptest.NewRequest("GET", "http://localhost:8080/aweiker/ToDo/1.0.0/lists", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Compare results.
	newlists := []model.TodoList{newlist}
	resultlists := []model.TodoList{}
	err = json.NewDecoder(resp.Body).Decode(&resultlists)
	assert.Nil(t, err)
	assert.Equal(t, newlists, resultlists)

	// Get lists w/parameters, succeeds.
	req = httptest.NewRequest("GET", "http://localhost:8080/aweiker/ToDo/1.0.0/lists?searchString=Home&skip=0&limit=1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Compare results.
	resultlists = []model.TodoList{}
	err = json.NewDecoder(resp.Body).Decode(&resultlists)
	assert.Nil(t, err)
	assert.Equal(t, newlists, resultlists)

	// Use duplicate params incorrectly.
	req = httptest.NewRequest("GET", "http://localhost:8080/aweiker/ToDo/1.0.0/lists?searchString=Home&searchString=work", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Use non-numeric integer param.
	req = httptest.NewRequest("GET", "http://localhost:8080/aweiker/ToDo/1.0.0/lists?skip=hello", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Use a bad parameter.
	req = httptest.NewRequest("GET", "http://localhost:8080/aweiker/ToDo/1.0.0/lists?fakeparam=1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	resp = rec.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
