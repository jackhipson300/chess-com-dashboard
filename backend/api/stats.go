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

type WinStats struct {
  NumResigns int `json:"resigns"`
  NumCheckmates int `json:"checkmates"`
  NumAbandons int `json:"abandons"`
  NumTimeouts int `json:"timeouts"`
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
		COALESCE(SUM(CASE WHEN g.winner = $1 THEN 1 ELSE NULL END), 0) as wins, 
		COALESCE(SUM(CASE WHEN g.winner IS NOT NULL AND g.winner != $1 THEN 1 ELSE NULL END), 0) as losses,
		COALESCE(SUM(CASE WHEN g.winner IS NULL THEN 1 ELSE NULL END), 0) as draws 
	FROM (
		SELECT DISTINCT time_class
		FROM games
	) tc
  LEFT JOIN games g ON tc.time_class = g.time_class
  GROUP BY tc.time_class
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

func GetWinStats(w http.ResponseWriter, req *http.Request, state *types.ServerState) {
	if !req.URL.Query().Has("username") {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	username := req.URL.Query().Get("username")

	if err := performSetupCheck(w, username); err != nil {
		fmt.Printf("Error getting win stats for user \"%s\": %s\n", username, err)
		return
	}

  queryStr := `
  SELECT
    tc.time_class,
    COALESCE(SUM(CASE WHEN g.result = 'resigned' THEN 1 ELSE NULL END), 0) as resigns,
    COALESCE(SUM(CASE WHEN g.result = 'checkmated' THEN 1 ELSE NULL END), 0) as checkmates,
    COALESCE(SUM(CASE WHEN g.result = 'abandoned' THEN 1 ELSE NULL END), 0) as abandons,
    COALESCE(SUM(CASE WHEN g.result = 'timeout' THEN 1 ELSE NULL END), 0) as timeouts
  FROM (
    SELECT DISTINCT time_class FROM games
  ) tc 
  LEFT JOIN
    games g ON tc.time_class = g.time_class AND g.winner = $1
  GROUP BY tc.time_class
  `

	db := state.DBMap[utils.Hash(username)]
	if db == nil {
		fmt.Println("Error making win stats query: db not found")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(queryStr, username)
	if err != nil {
		fmt.Printf("Error making win stats query: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
    
  response := make(map[string]WinStats)
  for rows.Next() {
    var timeClass string 
    var resigns int 
    var checkmates int 
    var abandons int 
    var timeouts int 

    if err := rows.Scan(&timeClass, &resigns, &checkmates, &abandons, &timeouts); err != nil {
      fmt.Printf("Error parsing win stats query result: %s\n", err)
      http.Error(w, "Internal server error", http.StatusInternalServerError)
      return
    }

    response[timeClass] = WinStats{
      NumResigns: resigns,
      NumCheckmates: checkmates,
      NumAbandons: abandons,
      NumTimeouts: timeouts,
    }
  }

  if err := json.NewEncoder(w).Encode(response); err != nil {
    fmt.Printf("Error encoding win stats query result: %s\n", err)
    http.Error(w, "Internal server error", http.StatusInternalServerError)
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
