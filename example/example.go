package example

import (
	"net/http"

	"github.com/labstack/echo"
)

func AddRoutes(r *echo.Router) {
	r.Add("POST", "/users", CreateUsers)
	r.Add("PUT", "/users", UpdateUsers)
	r.Add("DELETE", "/users", DeleteAllUsers)
	r.Add("DELETE", "/users/:id", DeleteUserWithID)
	r.Add("GET", "/users", ListUsers)
	r.Add("GET", "/users/:id", GetUserWithID)
}

type User struct {
	ID    string `json:"id"`
	EMAIL string `json:"email"`
}

var dumyUsers map[string]User = map[string]User{
	"gopher1": User{ID: "gopher1", EMAIL: "gopher1@gmail.com"},
	"gopher2": User{ID: "gopher2", EMAIL: "gopher2@gmail.com"},
}

// ListUsers return all of user informations.
func ListUsers(c echo.Context) error {
	return c.JSON(http.StatusOK, dumyUsers)
}

// GetUserWithID return a user's information with ID.
func GetUserWithID(c echo.Context) error {
	id := c.Param("id")
	user, ok := dumyUsers[id]
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, user)
}

// CreateUsers create users with request data.
func CreateUsers(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// UpdateUsers update requested users information.
func UpdateUsers(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// DeleteUserWithID delete an user's information with ID.
func DeleteUserWithID(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// DeleteAllUsers delete all of user informations.
func DeleteAllUsers(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

type Controller struct {
}

func (ctl Controller) AddRoutes(r *echo.Router) {
	r.Add("GET", "/withctl/users", ctl.ListUsers)
}

// ListUsers for controller method test
func (ctl Controller) ListUsers(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
