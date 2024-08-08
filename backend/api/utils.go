package api

import (
	"backend/types"
	"net/http"
)

func MakeHandler(
	state *types.ServerState,
	handler func(http.ResponseWriter, *http.Request, *types.ServerState),
) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		handler(w, req, state)
	}
}
