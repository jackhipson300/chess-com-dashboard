package api

import (
	"backend/model"
	"backend/types"
	"backend/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type SetupReqBody struct {
	Username string `json:"username"`
}

type SetupResp struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

func isSetup(username string) bool {
	dbFilename := fmt.Sprintf("%s.db", utils.Hash(username))
	if _, err := os.Stat(dbFilename); !os.IsNotExist(err) {
		return true
	}
	return false
}

func isSetupInProgress(setupRequests *map[string]string, username string) bool {
	requestId := utils.Hash(username)
	return (*setupRequests)[requestId] == "Started"
}

func performSetupCheck(w http.ResponseWriter, setupRequests *map[string]string, username string) error {
	if isSetupInProgress(setupRequests, username) {
		http.Error(w, "User data setup in progress", http.StatusBadRequest)
		return errors.New("user data setup in progress")
	}

	if !isSetup(username) {
		http.Error(w, "User not setup", http.StatusBadRequest)
		return errors.New("user not setup")
	}

	return nil
}

func Setup(w http.ResponseWriter, req *http.Request, state *types.ServerState) {
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

	requestId := utils.Hash(body.Username)
	if state.SetupRequests[requestId] != "" {
		json.NewEncoder(w).Encode(SetupResp{
			Id:     requestId,
			Status: state.SetupRequests[requestId],
		})
		return
	}

	state.SetupRequests[requestId] = "Started"
	json.NewEncoder(w).Encode(SetupResp{
		Id:     requestId,
		Status: "Started",
	})

	go func() {
		setupStart := time.Now()

		dbFilename := fmt.Sprintf("./%s.db", requestId)
		db, err := sql.Open("sqlite3", dbFilename)
		if err != nil {
			fmt.Println("Error opening db")
			state.SetupRequests[requestId] = "Failed"
			return
		}
		state.DBMap[requestId] = db

		model.CreateTables(db)

		fmt.Println("User data request started:")
		requestGamesStart := time.Now()
		allGames := model.GetAllGames(body.Username)
		duration := time.Since(requestGamesStart)
		fmt.Printf("%d games received in %v!\n", len(allGames), duration)

		fmt.Println("Inserting into db started:")
		insertStart := time.Now()
		insertStats, err := model.InsertUserData(db, allGames)
		if err != nil {
			fmt.Println("Critical failure occurred while inserting into db")
			state.SetupRequests[requestId] = "Failed"
			return
		}

		duration = time.Since(insertStart)
		fmt.Printf("Inserted user data in %v\n", duration)
		fmt.Printf("  %d games inserted\n", insertStats.NumGamesInserted)
		fmt.Printf("  %d games failed to insert\n", insertStats.NumGameInsertErrors)
		fmt.Printf("  %d positions inserted\n", insertStats.NumPositionsInserted)
		fmt.Printf("  %d positions failed to insert\n", insertStats.NumPositionInsertErrors)

		fmt.Printf("Downloaded and saved user data in %v\n", time.Since(setupStart))

    state.SetupRequests[requestId] = "Complete"
	}()
}
