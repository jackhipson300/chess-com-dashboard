package api

import (
	"backend/types"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

func archiveToLogicalTimestamp(archive string) (date int, err error) {
	regex, err := regexp.Compile("[0-9]{4}/[0-9]{2}$")
	if err != nil {
		return
	}

	match := regex.FindString(archive)
	if match == "" {
		err = fmt.Errorf("invalid archive format: %s", archive)
		return
	}

	parts := strings.Join(strings.Split(match, "/"), "")
	date, _ = strconv.Atoi(parts)

	return
}
