package main

import (
	"backend/api"
	"backend/types"
  "backend/model"
	"database/sql"
	"fmt"
	"net/http"
  "os"

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

  if err := model.LoadExistingDbs(&state); err != nil {
    fmt.Printf("Fatal error: %w\n")
    cleanup(&state)
    os.Exit(0)
  }

	http.HandleFunc("/setup", api.MakeHandler(&state, api.Setup))
	http.HandleFunc("/gamestats", api.MakeHandler(&state, api.GetGameStats))
  http.HandleFunc("/winstats", api.MakeHandler(&state, api.GetWinStats))
  http.HandleFunc("/lossstats", api.MakeHandler(&state, api.GetLossStats))
  http.HandleFunc("/drawstats", api.MakeHandler(&state, api.GetDrawStats))

	const host = "localhost"
	const port = ":8090"
	fmt.Printf("Listening on %s%s...\n", host, port)
	http.ListenAndServe(port, nil)
}
