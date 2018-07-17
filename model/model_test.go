package model

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestAddList(t *testing.T) {
	// Dummy list.
	newlist := TodoList{
		"d290f1ee-6c54-4b01-90e6-d701748f0851",
		"Home",
		"The list of things that need to be done at home\n",
		[]Task{Task{"0e2ac84f-f723-4f24-878b-44e63e7ae580", "mow the yard", true}},
	}

	// Add this list; succeeds.
	status := AddList(newlist)
	assert.Equal(t, http.StatusCreated, status)

	// Check list values.
	actuallist, status := GetList("d290f1ee-6c54-4b01-90e6-d701748f0851")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, newlist, actuallist)

	// Add it again; fails due to conflict.
	status = AddList(newlist)
	assert.Equal(t, http.StatusConflict, status)

	// Add a list with an invalid ID; fails.
	newlist.ID = "This is not a valid UUID"
	status = AddList(newlist)
	assert.Equal(t, http.StatusBadRequest, status)

	// Add a list with a new ID; succeeds.
	newlist.ID = "d290f1ee-6c54-4b01-90e6-d701748f0852"
	status = AddList(newlist)
	assert.Equal(t, http.StatusCreated, status)

	// Check list values.
	actuallist, status = GetList("d290f1ee-6c54-4b01-90e6-d701748f0852")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, newlist, actuallist)

	// Teardown.
	lists = make(listmap)
}

func TestAddTask(t *testing.T) {
	// Dummy list.
	newlist := TodoList{
		"d290f1ee-6c54-4b01-90e6-d701748f0851",
		"Home",
		"The list of things that need to be done at home\n",
		[]Task{},
	}

	// Add this list; succeeds.
	status := AddList(newlist)
	assert.Equal(t, http.StatusCreated, status)

	// Dummy task.
	newtask := Task{"0e2ac84f-f723-4f24-878b-44e63e7ae580", "mow the yard", true}

	// Add this task; succeeds.
	status = AddTask("d290f1ee-6c54-4b01-90e6-d701748f0851", newtask)
	assert.Equal(t, http.StatusCreated, status)

	// Check list values.
	newlist.Tasks = []Task{newtask}
	actuallist, status := GetList("d290f1ee-6c54-4b01-90e6-d701748f0851")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, newlist, actuallist)

	// Add it again; fails due to conflict.
	status = AddTask("d290f1ee-6c54-4b01-90e6-d701748f0851", newtask)
	assert.Equal(t, http.StatusConflict, status)

	// Add it to an invalid list; fails.
	status = AddTask("d290f1ee-6c54-4b01-90e6-d701748f0852", newtask)
	assert.Equal(t, http.StatusBadRequest, status)

	// Add a task with an invalid ID; fails.
	newtask.ID = "This is not a valid UUID"
	status = AddTask("d290f1ee-6c54-4b01-90e6-d701748f0851", newtask)
	assert.Equal(t, http.StatusBadRequest, status)

	// Add a task with a new ID; succeeds.
	newtask.ID = "0e2ac84f-f723-4f24-878b-44e63e7ae581"
	status = AddTask("d290f1ee-6c54-4b01-90e6-d701748f0851", newtask)
	assert.Equal(t, http.StatusCreated, status)

	// Check list values.
	newlist.Tasks = append(newlist.Tasks, newtask)
	actuallist, status = GetList("d290f1ee-6c54-4b01-90e6-d701748f0851")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, newlist, actuallist)

	// Teardown.
	lists = make(listmap)
}

func TestSetCompleted(t *testing.T) {
	// Dummy list.
	newlist := TodoList{
		"d290f1ee-6c54-4b01-90e6-d701748f0851",
		"Home",
		"The list of things that need to be done at home\n",
		[]Task{Task{"0e2ac84f-f723-4f24-878b-44e63e7ae580", "mow the yard", false}},
	}

	// Add this list; succeeds.
	status := AddList(newlist)
	assert.Equal(t, http.StatusCreated, status)

	// Set the task to complete; succeeds.
	completed := CompletedTask{true}
	status = SetCompleted("d290f1ee-6c54-4b01-90e6-d701748f0851", "0e2ac84f-f723-4f24-878b-44e63e7ae580", completed)
	assert.Equal(t, http.StatusCreated, status)

	// Check list values.
	newlist.Tasks[0].Completed = true
	actuallist, status := GetList("d290f1ee-6c54-4b01-90e6-d701748f0851")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, newlist, actuallist)

	// Set a task on an invalid list to complete; fails.
	status = SetCompleted("d290f1ee-6c54-4b01-90e6-d701748f0852", "0e2ac84f-f723-4f24-878b-44e63e7ae580", completed)
	assert.Equal(t, http.StatusBadRequest, status)

	// Set an invalid task to complete; fails.
	status = SetCompleted("d290f1ee-6c54-4b01-90e6-d701748f0851", "0e2ac84f-f723-4f24-878b-44e63e7ae581", completed)
	assert.Equal(t, http.StatusBadRequest, status)

	// Teardown.
	lists = make(listmap)
}

func TestGetList(t *testing.T) {
	// Getting lists successfully is covered by other tests.

	// Get a non-existent list; fails.
	_, status := GetList("d290f1ee-6c54-4b01-90e6-d701748f0851")
	assert.Equal(t, http.StatusNotFound, status)

	// Get a list with an invalid ID; fails.
	_, status = GetList("This is not a valid UUID")
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestGetLists(t *testing.T) {
	// An initial UUID.
	id, _ := uuid.Parse("d290f1ee-6c54-4b01-90e6-d701748f0851")

	// Dummy list #1.
	homelist := TodoList{
		id.String(),
		"Home",
		"The list of things that need to be done at home\n",
		[]Task{},
	}
	id[15]++ // Increment UUID; remains unique in the scope of this test

	// Dummy list #2.
	worklist := TodoList{
		id.String(),
		"Work",
		"The list of things that need to be done at work\n",
		[]Task{},
	}
	id[15]++

	// Add these lists; succeeds.
	status := AddList(homelist)
	assert.Equal(t, http.StatusCreated, status)
	status = AddList(worklist)
	assert.Equal(t, http.StatusCreated, status)

	// Retrieve these lists; succeeds.
	response, status := GetLists("", 0, 0)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, []TodoList{homelist, worklist}, response)

	// Retrieve the home list; succeeds.
	response, status = GetLists("Home", 0, 0)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, []TodoList{homelist}, response)

	// Retrieve the work list; succeeds.
	response, status = GetLists("Work", 0, 0)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, []TodoList{worklist}, response)

	// Retrieve all lists containing an O; succeeds.
	// (Yes, this is kind of a lame search.  I'd try to leverage the underlying
	// database in a more sophisticated service, but for the sake of showing my
	// ideas quickly, here we are.)
	response, status = GetLists("O", 0, 0)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, []TodoList{homelist, worklist}, response)

	// Retrieve a first page of lists containing an O; succeeds.
	response, status = GetLists("O", 0, 1)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, []TodoList{homelist}, response)

	// Retrieve a second page of lists containing an O; succeeds.
	response, status = GetLists("O", 1, 1)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, []TodoList{worklist}, response)

	// Retrieve a third page of lists containing an O; succeeds but is empty.
	response, status = GetLists("O", 2, 1)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, []TodoList{}, response)

	// Add a bunch of additional lists.
	homelists := []TodoList{homelist}
	worklists := []TodoList{worklist}
	for i := 0; i < 5; i++ {
		homelist.ID = id.String()
		status = AddList(homelist)
		assert.Equal(t, http.StatusCreated, status)
		homelists = append(homelists, homelist)
		id[15]++

		worklist.ID = id.String()
		status = AddList(worklist)
		assert.Equal(t, http.StatusCreated, status)
		worklists = append(worklists, worklist)
		id[15]++
	}

	// Retrieve all home lists; succeeds.
	response, status = GetLists("Home", 0, 0)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, homelists, response)

	// Retrieve all work lists; succeeds.
	response, status = GetLists("Work", 0, 0)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, worklists, response)

	// Pass bad parameters; fails.
	response, status = GetLists("", -1, -1)
	assert.Equal(t, http.StatusBadRequest, status)

	// Teardown.
	lists = make(listmap)
}
