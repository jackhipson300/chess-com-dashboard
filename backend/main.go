package main

import (
	"backend/api"
	"backend/model"
	"backend/types"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

func cleanup(state *types.ServerState) {
	fmt.Println("Cleaning up...")
	for key, db := range state.DBMap {
		fmt.Printf("Closing db: %s\n", key)
		db.Mu.Lock()
		defer db.Mu.Unlock()
		if err := db.Resource.Close(); err != nil {
			fmt.Printf("Error closing db: %s\n", key)
		}
	}
	fmt.Println("Cleanup complete!")
}

func main() {
	state := types.NewServerState()

	if err := model.LoadExistingDbs(&state.DBMap); err != nil {
		fmt.Printf("Fatal error: %s\n", err)
		cleanup(state)
		os.Exit(0)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/setup", api.MakeHandler(state, api.Setup))
	mux.HandleFunc("/gamestats", api.MakeHandler(state, api.GetGameStats))
	mux.HandleFunc("/winstats", api.MakeHandler(state, api.GetWinStats))
	mux.HandleFunc("/lossstats", api.MakeHandler(state, api.GetLossStats))
	mux.HandleFunc("/drawstats", api.MakeHandler(state, api.GetDrawStats))

	handler := cors.Default().Handler(mux)

	const host = "localhost"
	const port = ":8090"
	fmt.Printf("Listening on %s%s...\n", host, port)
	go http.ListenAndServe(port, handler)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	fmt.Println("Shutting down gracefully...")
	cleanup(state)

	os.Exit(0)
}
