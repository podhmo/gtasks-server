---
title: gtask-server
version: 0.0.0
---

# gtask-server

gtask-server

- [paths](#paths)
- [schemas](#schemas)

## paths

| endpoint | operationId | tags | summary |
| --- | --- | --- | --- |
| `GET /` | [main.MarkdownAPI.ListTaskList](#mainmarkdownapilisttasklist-get-)  | `main` |  |
| `GET /api/tasklist` | [main.TaskListAPI.List](#maintasklistapilist-get-apitasklist)  | `main` |  |
| `GET /api/tasklist/{tasklistId}` | [main.TaskAPI.List](#maintaskapilist-get-apitasklisttasklistid)  | `main` |  |
| `GET /{tasklistId}` | [main.MarkdownAPI.DetailTaskList](#mainmarkdownapidetailtasklist-get-tasklistid)  | `main` |  |


### main.MarkdownAPI.ListTaskList `GET /`



| name | value | 
| --- | --- |
| operationId | main.MarkdownAPI.ListTaskList |
| endpoint | `GET /` |
| tags | `main` |



#### output (application/json)

```go


// GET / (default)
// default error
type OutputDefault struct {	// ErrorResponse
	code integer

	error string

	detail? []string
}
```
### main.TaskListAPI.List `GET /api/tasklist`



| name | value | 
| --- | --- |
| operationId | main.TaskListAPI.List |
| endpoint | `GET /api/tasklist` |
| tags | `main` |



#### output (application/json)

```go

// GET /api/tasklist (200)
type Output200 []struct {	// TaskList
	// Etag: ETag of the resource.
	etag? string

	// Id: Task list identifier.
	id? string

	// Kind: Type of the resource. This is always "tasks#taskList".
	kind? string

	// SelfLink: URL pointing to this task list. Used to retrieve, update,
	// or delete this task list.
	selfLink? string

	// Title: Title of the task list.
	title? string

	// Updated: Last modification time of the task list (as a RFC 3339
	// timestamp).
	updated? string

	// HTTPStatusCode is the server's response status code. When using a
	// resource method's Do call, this will always be in the 2xx range.
	HTTPStatusCode integer

	// Header contains the response header fields from the server.
	Header map[string][]string
}

// GET /api/tasklist (default)
// default error
type OutputDefault struct {	// ErrorResponse
	code integer

	error string

	detail? []string
}
```
### main.TaskAPI.List `GET /api/tasklist/{tasklistId}`



| name | value | 
| --- | --- |
| operationId | main.TaskAPI.List |
| endpoint | `GET /api/tasklist/{tasklistId}` |
| tags | `main` |


#### input (application/json)

```go
// GET /api/tasklist/{tasklistId}
type Input struct {
	tasklistId string `in:"path"`
}
```

#### output (application/json)

```go

// GET /api/tasklist/{tasklistId} (200)
type Output200 []struct {	// Task
	// Completed: Completion date of the task (as a RFC 3339 timestamp).
	// This field is omitted if the task has not been completed.
	completed? string

	// Deleted: Flag indicating whether the task has been deleted. The
	// default is False.
	deleted? boolean

	// Due: Due date of the task (as a RFC 3339 timestamp). Optional. The
	// due date only records date information; the time portion of the
	// timestamp is discarded when setting the due date. It isn't possible
	// to read or write the time that a task is due via the API.
	due? string

	// Etag: ETag of the resource.
	etag? string

	// Hidden: Flag indicating whether the task is hidden. This is the case
	// if the task had been marked completed when the task list was last
	// cleared. The default is False. This field is read-only.
	hidden? boolean

	// Id: Task identifier.
	id? string

	// Kind: Type of the resource. This is always "tasks#task".
	kind? string

	// Links: Collection of links. This collection is read-only.
	links? []struct {	// TaskLinks
		// Description: The description. In HTML speak: Everything between <a>
		// and </a>.
		description? string

		// Link: The URL.
		link? string

		// Type: Type of the link, e.g. "email".
		type? string
	}

	// Notes: Notes describing the task. Optional.
	notes? string

	// Parent: Parent task identifier. This field is omitted if it is a
	// top-level task. This field is read-only. Use the "move" method to
	// move the task under a different parent or to the top level.
	parent? string

	// Position: String indicating the position of the task among its
	// sibling tasks under the same parent task or at the top level. If this
	// string is greater than another task's corresponding position string
	// according to lexicographical ordering, the task is positioned after
	// the other task under the same parent task (or at the top level). This
	// field is read-only. Use the "move" method to move the task to another
	// position.
	position? string

	// SelfLink: URL pointing to this task. Used to retrieve, update, or
	// delete this task.
	selfLink? string

	// Status: Status of the task. This is either "needsAction" or
	// "completed".
	status? string

	// Title: Title of the task.
	title? string

	// Updated: Last modification time of the task (as a RFC 3339
	// timestamp).
	updated? string

	// HTTPStatusCode is the server's response status code. When using a
	// resource method's Do call, this will always be in the 2xx range.
	HTTPStatusCode integer

	// Header contains the response header fields from the server.
	Header map[string][]string
}

// GET /api/tasklist/{tasklistId} (default)
// default error
type OutputDefault struct {	// ErrorResponse
	code integer

	error string

	detail? []string
}
```
### main.MarkdownAPI.DetailTaskList `GET /{tasklistId}`



| name | value | 
| --- | --- |
| operationId | main.MarkdownAPI.DetailTaskList |
| endpoint | `GET /{tasklistId}` |
| tags | `main` |


#### input (application/json)

```go
// GET /{tasklistId}
type Input struct {
	tasklistId string `in:"path"`
}
```

#### output (application/json)

```go


// GET /{tasklistId} (default)
// default error
type OutputDefault struct {	// ErrorResponse
	code integer

	error string

	detail? []string
}
```



----------------------------------------

## schemas

| name | summary |
| --- | --- |
| [ErrorResponse](#errorresponse) | represents a normal error response type |
| [Task](#task) |  |
| [TaskLinks](#tasklinks) |  |
| [TaskList](#tasklist) |  |



### ErrorResponse

```go
// ErrorResponse represents a normal error response type
type ErrorResponse struct {
	code integer

	error string

	detail? []string
}
```

- [output of main.MarkdownAPI.ListTaskList (default)](#mainmarkdownapilisttasklist-get-)
- [output of main.TaskListAPI.List (default)](#maintasklistapilist-get-apitasklist)
- [output of main.TaskAPI.List (default)](#maintaskapilist-get-apitasklisttasklistid)
- [output of main.MarkdownAPI.DetailTaskList (default)](#mainmarkdownapidetailtasklist-get-tasklistid)

### Task

```go
type Task struct {
	// Completed: Completion date of the task (as a RFC 3339 timestamp).
	// This field is omitted if the task has not been completed.
	completed? string

	// Deleted: Flag indicating whether the task has been deleted. The
	// default is False.
	deleted? boolean

	// Due: Due date of the task (as a RFC 3339 timestamp). Optional. The
	// due date only records date information; the time portion of the
	// timestamp is discarded when setting the due date. It isn't possible
	// to read or write the time that a task is due via the API.
	due? string

	// Etag: ETag of the resource.
	etag? string

	// Hidden: Flag indicating whether the task is hidden. This is the case
	// if the task had been marked completed when the task list was last
	// cleared. The default is False. This field is read-only.
	hidden? boolean

	// Id: Task identifier.
	id? string

	// Kind: Type of the resource. This is always "tasks#task".
	kind? string

	// Links: Collection of links. This collection is read-only.
	links? []struct {	// TaskLinks
		// Description: The description. In HTML speak: Everything between <a>
		// and </a>.
		description? string

		// Link: The URL.
		link? string

		// Type: Type of the link, e.g. "email".
		type? string
	}

	// Notes: Notes describing the task. Optional.
	notes? string

	// Parent: Parent task identifier. This field is omitted if it is a
	// top-level task. This field is read-only. Use the "move" method to
	// move the task under a different parent or to the top level.
	parent? string

	// Position: String indicating the position of the task among its
	// sibling tasks under the same parent task or at the top level. If this
	// string is greater than another task's corresponding position string
	// according to lexicographical ordering, the task is positioned after
	// the other task under the same parent task (or at the top level). This
	// field is read-only. Use the "move" method to move the task to another
	// position.
	position? string

	// SelfLink: URL pointing to this task. Used to retrieve, update, or
	// delete this task.
	selfLink? string

	// Status: Status of the task. This is either "needsAction" or
	// "completed".
	status? string

	// Title: Title of the task.
	title? string

	// Updated: Last modification time of the task (as a RFC 3339
	// timestamp).
	updated? string

	// HTTPStatusCode is the server's response status code. When using a
	// resource method's Do call, this will always be in the 2xx range.
	HTTPStatusCode integer

	// Header contains the response header fields from the server.
	Header map[string][]string
}
```


### TaskLinks

```go
type TaskLinks struct {
	// Description: The description. In HTML speak: Everything between <a>
	// and </a>.
	description? string

	// Link: The URL.
	link? string

	// Type: Type of the link, e.g. "email".
	type? string
}
```


### TaskList

```go
type TaskList struct {
	// Etag: ETag of the resource.
	etag? string

	// Id: Task list identifier.
	id? string

	// Kind: Type of the resource. This is always "tasks#taskList".
	kind? string

	// SelfLink: URL pointing to this task list. Used to retrieve, update,
	// or delete this task list.
	selfLink? string

	// Title: Title of the task list.
	title? string

	// Updated: Last modification time of the task list (as a RFC 3339
	// timestamp).
	updated? string

	// HTTPStatusCode is the server's response status code. When using a
	// resource method's Do call, this will always be in the 2xx range.
	HTTPStatusCode integer

	// Header contains the response header fields from the server.
	Header map[string][]string
}
```
