package model

import (
	"bytes"
	"net/http"
	"regexp"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// For a real service, we would presumably want to persist this data, but for
// the sake of this example we will use simple native Go data structures.  We
// do not use the same data structures we present to the user; instead, these
// functions convert to a more efficient internal format.  Of course, there
// tradeoffs in any internal format.  In this case, I have chosen a format
// which provides very fast insertion/lookup by ID, but searching is slower.
// Depending on the use cases we expect most often, we might choose different
// data structures and/or change the API; note the sorting below when
// retrieving sets of lists and tasks.  For tasks, in particular, do duplicate
// task names in the same list make sense?  Would it be better to store the
// tasks in a tree sorted by name and forget the unique ID?  And shouldn't the
// unique IDs be provided by the backend, not the frontend, rather than
// forcing clients to generate their own IDs?  A longer conversation
// could/should be had here.

type task struct {
	name      string
	completed bool
}

type taskmap map[uuid.UUID]task

type list struct {
	name        string
	description string
	tasks       taskmap
}

type listmap map[uuid.UUID]list

// The internal database is a map of UUIDs to lists protected by a reader/
// writer lock.  It is safe for multiple readers to access a Go map at once,
// but we need to stop reading before anyone can write.
var lists = make(listmap)
var lock = sync.RWMutex{}

// These functions all return HTTP status codes; this should ideally be
// be changed to not rely on the HTTP protocol definitions and the API should
// convert to HTTP codes, but I have left this out for this sample service.

// Internal task adding helper.  Doesn't lock; the caller must lock before
// obtaining the taskmap if necessary.
func addTaskHelper(tasks taskmap, model Task) int {
	// Parse the task ID.
	taskid, err := uuid.Parse(model.ID)
	if err != nil {
		return http.StatusBadRequest
	}

	// Check for a conflict.
	if _, ok := tasks[taskid]; ok {
		return http.StatusConflict
	}

	// Add the task.
	tasks[taskid] = task{model.Name, model.Completed}
	return http.StatusCreated
}

// AddList takes a model for a list and adds it to the internal data
// structures.
func AddList(model TodoList) int {
	// Parse the list ID.
	listid, err := uuid.Parse(model.ID)
	if err != nil {
		return http.StatusBadRequest
	}

	// Create a new list and add tasks to it.  Doesn't lock yet; we're not
	// modifying the database.
	newlist := list{model.Name, model.Description, make(taskmap)}
	for _, newtask := range model.Tasks {
		status := addTaskHelper(newlist.tasks, newtask)
		if status != http.StatusCreated {
			return status
		}
	}

	// Lock the database for writing.
	lock.Lock()
	defer lock.Unlock()

	// Check for a conflict before committing.  We could check earlier as well
	// if building the list was difficult, but to avoid a race we must check
	// once we obtain the lock.
	if _, ok := lists[listid]; ok {
		return http.StatusConflict
	}

	// Modify the actual database.
	lists[listid] = newlist
	return http.StatusCreated
}

// AddTask takes a model for a task and adds it to the internal data
// structures.
func AddTask(id string, model Task) int {
	// Parse the list ID.
	listid, err := uuid.Parse(id)
	if err != nil {
		return http.StatusBadRequest
	}

	// Lock the database for writing.
	lock.Lock()
	defer lock.Unlock()

	// Find the list to modify.
	list, ok := lists[listid]
	if !ok {
		return http.StatusBadRequest
	}

	// Modify the actual database.
	return addTaskHelper(list.tasks, model)
}

// SetCompleted takes a model for task completion and modifies the internal
// data structures.
func SetCompleted(id string, taskID string, model CompletedTask) int {
	// Parse the IDs.
	listid, err := uuid.Parse(id)
	if err != nil {
		return http.StatusBadRequest
	}
	taskid, err := uuid.Parse(taskID)
	if err != nil {
		return http.StatusBadRequest
	}

	// Lock the database for writing.
	lock.Lock()
	defer lock.Unlock()

	// Find the task to modify.
	list, ok := lists[listid]
	if !ok {
		return http.StatusBadRequest
	}
	task, ok := list.tasks[taskid]
	if !ok {
		return http.StatusBadRequest
	}

	// Modify the actual database.
	task.completed = model.Completed
	list.tasks[taskid] = task
	return http.StatusCreated
}

// GetList returns a model for a list.
func GetList(id string) (TodoList, int) {
	response := TodoList{}

	// Parse the list ID.
	listid, err := uuid.Parse(id)
	if err != nil {
		return response, http.StatusBadRequest
	}
	response.ID = listid.String() // Use the canonical form

	// Lock the database for reading.
	lock.RLock()
	defer lock.RUnlock()

	// Find the list.
	list, ok := lists[listid]
	if !ok {
		return response, http.StatusNotFound
	}

	// Produce the output model.
	response.Name = list.name
	response.Description = list.description
	response.Tasks = make([]Task, 0, len(list.tasks))
	for taskid, task := range list.tasks {
		response.Tasks = append(response.Tasks, Task{taskid.String(), task.name, task.completed})
	}

	// Sort the tasks by name.  We could provide some different sort options
	// in the API, but for now, this seems like the most reasonable default.  We
	// could provide them unsorted, but this is bad for testing and probably not
	// what users will expect, either.
	sort.Slice(response.Tasks, func(i, j int) bool {
		if response.Tasks[i].Name < response.Tasks[j].Name {
			return true
		}

		// Break ties by ID to provide a stable sort.
		if response.Tasks[i].Name == response.Tasks[j].Name && response.Tasks[i].ID < response.Tasks[j].ID {
			return true
		}

		return false
	})

	return response, http.StatusOK
}

// GetLists returns a model for a range of lists, potentially limited by a
// search term and/or using pagination.  A limit of zero is treated as no
// limit.
func GetLists(searchString string, skip int, limit int) ([]TodoList, int) {
	response := []TodoList{}

	// Check the pagination parameters.
	if skip < 0 || limit < 0 {
		return response, http.StatusBadRequest
	}

	// We'll perform a case-insensitive search across the list names.  This
	// could be written to take into account descriptions, tasks, etc.; a matter
	// for further discussion, as well as whether words should be searched
	// individually, how to treat special characters, etc.  For the present
	// example, this will suffice.  Compile a regular expression to use.
	re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(searchString))
	if err != nil {
		return response, http.StatusBadRequest
	}

	// Lock the database.
	lock.RLock()
	defer lock.RUnlock()

	type searchresult struct {
		name string
		id   uuid.UUID
	}

	// Pagination is an interesting topic.  In most databases, using a limit and
	// offset repeats the work of the search every time; after all, the 51st
	// through 100th items might have changed since a previous query, so from a
	// correctness perspective the work should be repeated.  This is potentially
	// inefficient over a large search space, however, and is also susceptible
	// to drift if new items are inserted/removed between requests.
	// Nevertheless, we will follow this model here while pointing out that the
	// API could be changed to avoid these traps (by providing a record we left
	// off at previously, say)--which would require additional work on the
	// backend, of course.
	results := []searchresult{}
	for listid, list := range lists {
		if re.MatchString(list.name) {
			results = append(results, searchresult{list.name, listid})
		}
	}

	// Sort the results by name.  We could provide some different sort options
	// in the API, but for now, this seems like the most reasonable default.  We
	// could provide them unsorted, but this makes a mockery of pagination.
	sort.Slice(results, func(i, j int) bool {
		if results[i].name < results[j].name {
			return true
		}

		// Break ties by ID to provide a stable sort.
		if results[i].name == results[j].name && bytes.Compare(results[i].id[:], results[j].id[:]) < 0 {
			return true
		}

		return false
	})

	// Slice the result page.  Go is picky and will get mad if we pass indices
	// beyond the bounds of the result set, but we will simply choose to return
	// a number of results below the maximum.
	if limit == 0 {
		limit = len(results) // No limit
	}
	if skip > len(results) {
		skip = len(results)
	}
	limit += skip
	if limit > len(results) {
		limit = len(results)
	}
	results = results[skip:limit]

	// Produce the output model.
	for _, result := range results {
		list := lists[result.id]
		tasks := []Task{}
		for taskid, task := range list.tasks {
			tasks = append(tasks, Task{taskid.String(), task.name, task.completed})
		}
		response = append(response, TodoList{result.id.String(), list.name, list.description, tasks})
	}
	return response, http.StatusOK
}
