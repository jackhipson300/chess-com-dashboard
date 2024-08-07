package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SetupReqBody struct {
	Username string `json:"username"`
}

type SetupResp struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

var pendingSetupRequests = make(map[string]string)

func setup(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body SetupReqBody
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if body.Username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
	}

	requestId := hash(body.Username)
	if pendingSetupRequests[requestId] != "" {
		json.NewEncoder(w).Encode(SetupResp{
			Id:     requestId,
			Status: pendingSetupRequests[requestId],
		})
		return
	}

	dbFilename := fmt.Sprintf("./%s.db", requestId)
	if _, err := os.Stat(dbFilename); !os.IsNotExist(err) {
		json.NewEncoder(w).Encode(SetupResp{
			Id:     requestId,
			Status: "Complete",
		})
		return
	}

	pendingSetupRequests[requestId] = "Started"
	json.NewEncoder(w).Encode(SetupResp{
		Id:     requestId,
		Status: "Started",
	})

	go func() {
		setupStart := time.Now()

		db, err := sql.Open("sqlite3", dbFilename)
		if err != nil {
			fmt.Println("Error opening db")
			panic(err)
		}
		defer db.Close()

		createTables(db)

		fmt.Println("User data request started:")
		requestGamesStart := time.Now()
		allGames := getAllGames(body.Username)
		duration := time.Since(requestGamesStart)
		fmt.Printf("%d games received in %v!\n", len(allGames), duration)

		fmt.Println("Inserting into db started:")
		insertStart := time.Now()
		insertStats, err := insertUserData(db, allGames)
		if err != nil {
			fmt.Println("Critical failure occurred while inserting into db")
			panic(err)
		}
		duration = time.Since(insertStart)
		fmt.Printf("Inserted user data in %v\n", duration)
		fmt.Printf("  %d games inserted\n", insertStats.numGamesInserted)
		fmt.Printf("  %d games failed to insert\n", insertStats.numGameInsertErrors)
		fmt.Printf("  %d positions inserted\n", insertStats.numPositionsInserted)
		fmt.Printf("  %d positions failed to insert\n", insertStats.numPositionInsertErrors)

		fmt.Printf("Downloaded and saved user data in %v\n", time.Since(setupStart))
	}()
}

func main() {
	http.HandleFunc("/setup", setup)

	const host = "localhost"
	const port = ":8090"
	fmt.Printf("Listening on %s%s...\n", host, port)
	http.ListenAndServe(port, nil)
}
