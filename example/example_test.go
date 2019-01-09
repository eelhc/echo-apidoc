package example_test

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	apidoc "github.com/eelhc/echo-apidoc"
	"github.com/eelhc/echo-apidoc/example"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	addr          = "127.0.0.1:8080"
	shutdownTimeo = 5 * time.Second
)

func TestAPIDocExample(t *testing.T) {
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true
	example.AddRoutes(e.Router())

	doc := apidoc.New()
	go func() {
		e.Use(doc.Recorder())
		e.Start(addr)
	}()

	client := http.Client{Timeout: time.Second}

	var err error
	var resp *http.Response
	var req *http.Request
	resp, err = client.Get("http://" + addr + "/users")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	resp, err = client.Post(
		"http://"+addr+"/users",
		"application/json; charset=UTF-8",
		bytes.NewBufferString(`[{"id":"gopher3", "email": "gopher3@gmail.com"}]`),
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	resp, err = client.Get("http://" + addr + "/users/gopher1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	req, err = http.NewRequest(
		http.MethodDelete,
		"http://" + addr + "/users/gopher1",
		nil,
	)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	req, err = http.NewRequest(
		http.MethodDelete,
		"http://" + addr + "/users",
		nil,
	)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	req, err = http.NewRequest(
		http.MethodPut,
		"http://" + addr + "/users",
		bytes.NewBufferString(`[{"id":"gopher1", "email": "gopher111111111@gmail.com"}]`),
	)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)


	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeo)
	defer cancel()
	require.NoError(t, ctx.Err())
	require.NoError(t, e.Shutdown(ctx))
	assert.NoError(t, doc.Write("."))
}
