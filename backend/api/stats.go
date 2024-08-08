package api

import (
	"backend/types"
	"backend/utils"
	"encoding/json"
	"fmt"
	"net/http"
)

type GameStats struct {
	NumWins   int `json:"wins"`
	NumLosses int `json:"losses"`
	NumDraws  int `json:"draws"`
}

type GetGameStatsResponse struct {
	Classical GameStats `json:"classical"`
	Rapid     GameStats `json:"rapid"`
	Blitz     GameStats `json:"blitz"`
	Bullet    GameStats `json:"bullet"`
}

func GetGameStats(w http.ResponseWriter, req *http.Request, state *types.ServerState) {
	if !req.URL.Query().Has("username") {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	username := req.URL.Query().Get("username")

	if err := performSetupCheck(w, username); err != nil {
		fmt.Printf("Error getting game stats for user \"%s\": %s\n", username, err)
		return
	}

	queryStr := `
	SELECT 
		tc.time_class, 
		COALESCE(w.wins, 0) as wins, 
		COALESCE(l.losses, 0) as losses,
		COALESCE(d.draws, 0) as draws 
	FROM (
		SELECT DISTINCT time_class
		FROM games
	) tc
	LEFT JOIN (
		SELECT time_class, count(*) as wins
		FROM games 
		WHERE winner = $1
		GROUP BY time_class
	) w ON tc.time_class = w.time_class
	LEFT JOIN (
		SELECT time_class, count(*) as losses 
		FROM games 
		WHERE winner IS NOT NULL AND winner != $1 
		GROUP BY time_class
	) l ON tc.time_class = l.time_class
	LEFT JOIN (
		SELECT time_class, count(*) as draws 
		FROM games 
		WHERE winner IS NULL 
		GROUP BY time_class 
	) d ON tc.time_class = d.time_class
	`

	db := state.DBMap[utils.Hash(username)]
	if db == nil {
		fmt.Println("Error making game stats query: db not found")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(queryStr, username)
	if err != nil {
		fmt.Printf("Error making game stats query: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	response := make(map[string]GameStats)
	for rows.Next() {
		var timeClass string
		var wins int
		var losses int
		var draws int

		if err := rows.Scan(&timeClass, &wins, &losses, &draws); err != nil {
			fmt.Printf("Error parsing game stats query result: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response[timeClass] = GameStats{
			NumWins:   wins,
			NumLosses: losses,
			NumDraws:  draws,
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding game stats query result: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetWinStats(w http.ResponseWriter, req *http.Request) {
	if !req.URL.Query().Has("username") {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	username := req.URL.Query().Get("username")

	if err := performSetupCheck(w, username); err != nil {
		fmt.Printf("Error getting win stats for user \"%s\": %s\n", username, err)
		return
	}
}

func GetLossStats(w http.ResponseWriter, req *http.Request) {
	if !req.URL.Query().Has("username") {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	username := req.URL.Query().Get("username")

	if err := performSetupCheck(w, username); err != nil {
		fmt.Printf("Error getting loss stats for user \"%s\": %s\n", username, err)
		return
	}
}

func GetDrawStats(w http.ResponseWriter, req *http.Request) {
	if !req.URL.Query().Has("username") {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	username := req.URL.Query().Get("username")

	if err := performSetupCheck(w, username); err != nil {
		fmt.Printf("Error getting draw stats for user \"%s\": %s\n", username, err)
		return
	}
}
