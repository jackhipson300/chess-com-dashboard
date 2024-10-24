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

func handleSetupError(requestId string, err error, setupStatuses *types.SetupStatuses) {
	fmt.Println("Error during setup:", err.Error())
	setupStatuses.Mu.Lock()
	defer setupStatuses.Mu.Unlock()
	(*setupStatuses.Resource)[requestId] = types.SetupStatusFailed
}

func printInsertStats(stats model.InsertStatistics) {
	fmt.Printf("  %d games inserted\n", stats.NumGamesInserted)
	fmt.Printf("  %d games failed to insert\n", stats.NumGameInsertErrors)
	fmt.Printf("  %d positions inserted\n", stats.NumPositionsInserted)
	fmt.Printf("  %d positions failed to insert\n", stats.NumPositionInsertErrors)
}

func fullSetup(requestId string, username string, state *types.ServerState) {
	setupStart := time.Now()

	dbFilename := fmt.Sprintf("file:%s.db?_journal_mode=WAL&_synchronous=NORMAL", requestId)
	db, err := sql.Open("sqlite3", dbFilename)
	if err != nil {
		handleSetupError(requestId, fmt.Errorf("error opening db: %w", err), &state.SetupStatuses)
		return
	}

	state.DBMap[requestId] = types.NewLockedDB(db)

	model.CreateTables(db)

	fmt.Println("User data request started:")
	requestGamesStart := time.Now()

	archives, err := model.ListArchives(username)
	if err != nil {
		handleSetupError(requestId, fmt.Errorf("error opening db: %w", err), &state.SetupStatuses)
		return
	}

	allGames := model.GetAllGames(archives)
	duration := time.Since(requestGamesStart)
	fmt.Printf("%d games received in %v!\n", len(allGames), duration)

	fmt.Println("Inserting into db started:")
	insertStart := time.Now()
	insertStats, err := model.InsertUserData(db, requestId, username, allGames, archives)
	if err != nil {
		handleSetupError(requestId, fmt.Errorf("error inserting user data: %w", err), &state.SetupStatuses)
		return
	}

	duration = time.Since(insertStart)
	fmt.Printf("Inserted user data in %v\n", duration)
	printInsertStats(insertStats)
	fmt.Printf("Downloaded and saved user data in %v\n", time.Since(setupStart))

	state.SetupStatuses.Mu.Lock()
	defer state.SetupStatuses.Mu.Unlock()
	(*state.SetupStatuses.Resource)[requestId] = types.SetupStatusComplete
}

func updateExistingUser(requestId string, username string, state *types.ServerState) {
	db, exists := state.DBMap[requestId]
	if !exists {
		fmt.Printf("Error updating existing user: db for user %s doesn't exist\n", requestId)
	}
	db.Mu.Lock()
	defer db.Mu.Unlock()

	allArchives, err := model.ListArchives(username)
	if err != nil {
		handleSetupError(requestId, fmt.Errorf("error listing archives: %w", err), &state.SetupStatuses)
		return
	}

	latestStoredArchive, err := model.GetMostRecentArchive(requestId, db.Resource)
	if err != nil {
		handleSetupError(requestId, fmt.Errorf("error getting most recent archive: %w", err), &state.SetupStatuses)
		return
	}

	latestDate, err := archiveToLogicalTimestamp(latestStoredArchive)
	if err != nil {
		handleSetupError(requestId, fmt.Errorf("error converting latest archive to date: %w", err), &state.SetupStatuses)
		return
	}

	archivesToUpdate := []string{}
	for _, archive := range allArchives {
		date, err := archiveToLogicalTimestamp(archive)
		if err != nil {
			handleSetupError(requestId, fmt.Errorf("error converting archive to date: %w", err), &state.SetupStatuses)
			return
		}

		if date >= latestDate {
			archivesToUpdate = append(archivesToUpdate, archive)
		}
	}

	games := model.GetAllGames(archivesToUpdate)
	insertStats, err := model.InsertUserData(db.Resource, requestId, username, games, archivesToUpdate)
	if err != nil {
		err = fmt.Errorf("error inserting user data: %w", err)
		handleSetupError(requestId, err, &state.SetupStatuses)
		return
	}

	printInsertStats(insertStats)
	state.SetupStatuses.Mu.Lock()
	defer state.SetupStatuses.Mu.Unlock()
	(*state.SetupStatuses.Resource)[requestId] = types.SetupStatusComplete
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
