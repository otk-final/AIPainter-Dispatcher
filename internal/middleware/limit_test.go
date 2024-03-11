package middleware

import (
	"github.com/gorilla/mux"
	"net/http"
	"testing"
)

func TestRouteMatch(t *testing.T) {
	r := mux.NewRouter()

	r.Methods("POST").Path("/xxs/a/{prompt_id}")
	r.Path("/xxs/b")
	r.Path("/xxs/c")
	r.Path("/xxs/d")

	req, _ := http.NewRequest("POST", "http://localhost/xxs/a/123/a", nil)
	matched := r.Match(req, &mux.RouteMatch{})
	t.Log(matched)
}
