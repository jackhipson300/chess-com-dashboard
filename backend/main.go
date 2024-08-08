package main

import (
	"backend/api"
	"backend/types"
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func cleanup(state *types.ServerState) {
	fmt.Println("Cleaning up...")
	for key, db := range state.DBMap {
		fmt.Printf("Closing db: %s\n", key)
		if err := db.Close(); err != nil {
			fmt.Printf("Error closing db: %s\n", key)
		}
	}
	fmt.Println("Cleanup complete!")
}

func main() {
	state := types.ServerState{
		DBMap: make(map[string]*sql.DB),
	}
	defer cleanup(&state)

	http.HandleFunc("/setup", api.MakeHandler(&state, api.Setup))
	http.HandleFunc("/gamestats", api.MakeHandler(&state, api.GetGameStats))

	const host = "localhost"
	const port = ":8090"
	fmt.Printf("Listening on %s%s...\n", host, port)
	http.ListenAndServe(port, nil)
}
