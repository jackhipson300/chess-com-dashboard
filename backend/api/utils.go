package api

import (
	"backend/types"
	"net/http"
  "fmt"
)

func MakeHandler(
	state *types.ServerState,
	handler func(http.ResponseWriter, *http.Request, *types.ServerState),
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
    fmt.Printf("Request received: %s %s\n", req.Method, req.URL)
		handler(w, req, state)
	}
}
