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
	Id     string            `json:"id"`
	Status types.SetupStatus `json:"status"`
}

func isSetup(username string) bool {
	dbFilename := fmt.Sprintf("%s.db", utils.Hash(username))
	if _, err := os.Stat(dbFilename); !os.IsNotExist(err) {
		return true
	}
	return false
}

func isSetupInProgress(setupStatuses *types.SetupStatuses, username string) bool {
	requestId := utils.Hash(username)

	setupStatuses.Mu.Lock()
	defer setupStatuses.Mu.Unlock()
	return (*setupStatuses.Resource)[requestId] == types.SetupStatusStarted
}

func performSetupCheck(w http.ResponseWriter, setupStatuses *types.SetupStatuses, username string) error {
	if isSetupInProgress(setupStatuses, username) {
		http.Error(w, "User data setup in progress", http.StatusBadRequest)
		return errors.New("user data setup in progress")
	}

	if !isSetup(username) {
		http.Error(w, "User not setup", http.StatusBadRequest)
		return errors.New("user not setup")
	}

	return nil
}

func validateSetupRequest(w http.ResponseWriter, req *http.Request) (body SetupReqBody, valid bool) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if body.Username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
	}

	valid = true
	return
}

func fullSetup(requestId string, username string, state *types.ServerState) {
	setupStart := time.Now()

	dbFilename := fmt.Sprintf("file:%s.db?_journal_mode=WAL&_synchronous=NORMAL", requestId)
	db, err := sql.Open("sqlite3", dbFilename)
	if err != nil {
		fmt.Println("Error opening db")
		state.SetupStatuses.Mu.Lock()
		defer state.SetupStatuses.Mu.Unlock()
		(*state.SetupStatuses.Resource)[requestId] = types.SetupStatusFailed
		return
	}

	state.DBMap[requestId] = types.NewLockedDB(db)

	model.CreateTables(db)

	fmt.Println("User data request started:")
	requestGamesStart := time.Now()

	archives := model.ListArchives(username)
	allGames := model.GetAllGames(archives)
	duration := time.Since(requestGamesStart)
	fmt.Printf("%d games received in %v!\n", len(allGames), duration)

	fmt.Println("Inserting into db started:")
	insertStart := time.Now()
	insertStats, err := model.InsertUserData(db, requestId, username, allGames, archives)
	if err != nil {
		fmt.Println("Critical failure occurred while inserting into db:", err.Error())
		state.SetupStatuses.Mu.Lock()
		defer state.SetupStatuses.Mu.Unlock()
		(*state.SetupStatuses.Resource)[requestId] = types.SetupStatusFailed
		return
	}

	duration = time.Since(insertStart)
	fmt.Printf("Inserted user data in %v\n", duration)
	fmt.Printf("  %d games inserted\n", insertStats.NumGamesInserted)
	fmt.Printf("  %d games failed to insert\n", insertStats.NumGameInsertErrors)
	fmt.Printf("  %d positions inserted\n", insertStats.NumPositionsInserted)
	fmt.Printf("  %d positions failed to insert\n", insertStats.NumPositionInsertErrors)

	fmt.Printf("Downloaded and saved user data in %v\n", time.Since(setupStart))

	state.SetupStatuses.Mu.Lock()
	defer state.SetupStatuses.Mu.Unlock()
	(*state.SetupStatuses.Resource)[requestId] = types.SetupStatusPending
}

func updateExistingUser(requestId string, username string, state *types.ServerState) {
	db, exists := state.DBMap[requestId]
	if !exists {
		fmt.Printf("Error updating existing user: db for user %s doesn't exist\n", requestId)
	}
	db.Mu.Lock()
	defer db.Mu.Unlock()

	allArchives := model.ListArchives(username)
	latestStoredArchive, _ := model.GetMostRecentArchive(requestId, db.Resource)
	latestDate, _ := archiveToLogicalTimestamp(latestStoredArchive)

	archivesToUpdate := []string{}
	for _, archive := range allArchives {
		date, _ := archiveToLogicalTimestamp(archive)
		if date >= latestDate {
			archivesToUpdate = append(archivesToUpdate, archive)
		}
	}

	fmt.Println("latest archive", latestStoredArchive)
	fmt.Println("archives to update", archivesToUpdate)

	games := model.GetAllGames(archivesToUpdate)
	insertStats, err := model.InsertUserData(db.Resource, requestId, username, games, archivesToUpdate)
	if err != nil {
		fmt.Println("Error updating user data:", err.Error())
	}

	fmt.Println(insertStats)
}

func Setup(w http.ResponseWriter, req *http.Request, state *types.ServerState) {
	state.SetupStatuses.Mu.Lock()
	defer state.SetupStatuses.Mu.Unlock()

	body, valid := validateSetupRequest(w, req)
	if !valid {
		return
	}

	jsonEncoder := json.NewEncoder(w)

	requestId := utils.Hash(body.Username)
	currentStatus := (*state.SetupStatuses.Resource)[requestId]

	// if the setup is pending, we need to update the existing data instead of doing a full setup
	if currentStatus == types.SetupStatusPending {
		(*state.SetupStatuses.Resource)[requestId] = types.SetupStatusUpdating
		jsonEncoder.Encode(SetupResp{
			Id:     requestId,
			Status: types.SetupStatusUpdating,
		})
		go updateExistingUser(requestId, body.Username, state)
		return
	}

	// if setup has already been called, return the existing state
	if currentStatus != "" {
		jsonEncoder.Encode(SetupResp{
			Id:     requestId,
			Status: currentStatus,
		})
		return
	}

	// notify the client that the setup has started
	(*state.SetupStatuses.Resource)[requestId] = types.SetupStatusStarted
	jsonEncoder.Encode(SetupResp{
		Id:     requestId,
		Status: types.SetupStatusStarted,
	})

	go fullSetup(requestId, body.Username, state)
}
