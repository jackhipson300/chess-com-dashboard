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
	Total     int `json:"total"`
}

type WinLossStats struct {
	NumResigns    int `json:"resigns"`
	NumCheckmates int `json:"checkmates"`
	NumAbandons   int `json:"abandons"`
	NumTimeouts   int `json:"timeouts"`
	Total         int `json:"total"`
}

type DrawStats struct {
	NumRepetitions            int `json:"repetitions"`
	NumInsufficients          int `json:"insufficients"`
	NumTimeoutVsInsufficients int `json:"timeoutVsInsufficients"`
	NumStalemates             int `json:"stalemates"`
	NumAgrees                 int `json:"agrees"`
	Num50Rules                int `json:"fiftyMoveRules"`
	Total                     int `json:"total"`
}

func GetGameStats(w http.ResponseWriter, req *http.Request, state *types.ServerState) {
	if !req.URL.Query().Has("username") {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	username := req.URL.Query().Get("username")

	if err := performSetupCheck(w, &state.SetupStatuses, username); err != nil {
		fmt.Printf("Error getting game stats for user \"%s\": %s\n", username, err)
		return
	}

	queryStr := `
	SELECT 
		tc.time_class, 
		COALESCE(SUM(CASE WHEN g.winner = $1 THEN 1 ELSE NULL END), 0) as wins, 
		COALESCE(SUM(CASE WHEN g.winner IS NOT NULL AND g.winner != $1 THEN 1 ELSE NULL END), 0) as losses,
		COALESCE(SUM(CASE WHEN g.winner IS NULL THEN 1 ELSE NULL END), 0) as draws,
		COUNT(*) as total
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
	db.Mu.Lock()
	defer db.Mu.Unlock()

	rows, err := db.Resource.Query(queryStr, username)
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
		var total int

		if err := rows.Scan(&timeClass, &wins, &losses, &draws, &total); err != nil {
			fmt.Printf("Error parsing game stats query result: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response[timeClass] = GameStats{
			NumWins:   wins,
			NumLosses: losses,
			NumDraws:  draws,
			Total:     total,
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

	if err := performSetupCheck(w, &state.SetupStatuses, username); err != nil {
		fmt.Printf("Error getting win stats for user \"%s\": %s\n", username, err)
		return
	}

	queryStr := `
  SELECT
    tc.time_class,
    COALESCE(SUM(CASE WHEN g.result = 'resigned' THEN 1 ELSE NULL END), 0) as resigns,
    COALESCE(SUM(CASE WHEN g.result = 'checkmated' THEN 1 ELSE NULL END), 0) as checkmates,
    COALESCE(SUM(CASE WHEN g.result = 'abandoned' THEN 1 ELSE NULL END), 0) as abandons,
    COALESCE(SUM(CASE WHEN g.result = 'timeout' THEN 1 ELSE NULL END), 0) as timeouts,
		COUNT(*) as total
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
	db.Mu.Lock()
	defer db.Mu.Unlock()

	rows, err := db.Resource.Query(queryStr, username)
	if err != nil {
		fmt.Printf("Error making win stats query: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	response := make(map[string]WinLossStats)
	for rows.Next() {
		var timeClass string
		var resigns int
		var checkmates int
		var abandons int
		var timeouts int
		var total int

		if err := rows.Scan(&timeClass, &resigns, &checkmates, &abandons, &timeouts, &total); err != nil {
			fmt.Printf("Error parsing win stats query result: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response[timeClass] = WinLossStats{
			NumResigns:    resigns,
			NumCheckmates: checkmates,
			NumAbandons:   abandons,
			NumTimeouts:   timeouts,
			Total:         total,
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding win stats query result: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetLossStats(w http.ResponseWriter, req *http.Request, state *types.ServerState) {
	if !req.URL.Query().Has("username") {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	username := req.URL.Query().Get("username")

	if err := performSetupCheck(w, &state.SetupStatuses, username); err != nil {
		fmt.Printf("Error getting loss stats for user \"%s\": %s\n", username, err)
		return
	}

	queryStr := `
  SELECT
    tc.time_class,
    COALESCE(SUM(CASE WHEN g.result = 'resigned' THEN 1 ELSE NULL END), 0) as resigns,
    COALESCE(SUM(CASE WHEN g.result = 'checkmated' THEN 1 ELSE NULL END), 0) as checkmates,
    COALESCE(SUM(CASE WHEN g.result = 'abandoned' THEN 1 ELSE NULL END), 0) as abandons,
    COALESCE(SUM(CASE WHEN g.result = 'timeout' THEN 1 ELSE NULL END), 0) as timeouts,
		COUNT(*) as total
  FROM (
    SELECT DISTINCT time_class FROM games
  ) tc 
  LEFT JOIN
    games g ON tc.time_class = g.time_class AND g.winner != $1
  GROUP BY tc.time_class
  `

	db := state.DBMap[utils.Hash(username)]
	if db == nil {
		fmt.Println("Error making loss stats query: db not found")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	db.Mu.Lock()
	defer db.Mu.Unlock()

	rows, err := db.Resource.Query(queryStr, username)
	if err != nil {
		fmt.Printf("Error making loss stats query: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	response := make(map[string]WinLossStats)
	for rows.Next() {
		var timeClass string
		var resigns int
		var checkmates int
		var abandons int
		var timeouts int
		var total int

		if err := rows.Scan(&timeClass, &resigns, &checkmates, &abandons, &timeouts, &total); err != nil {
			fmt.Printf("Error parsing loss stats query result: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response[timeClass] = WinLossStats{
			NumResigns:    resigns,
			NumCheckmates: checkmates,
			NumAbandons:   abandons,
			NumTimeouts:   timeouts,
			Total:         total,
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding loss stats query result: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func GetDrawStats(w http.ResponseWriter, req *http.Request, state *types.ServerState) {
	if !req.URL.Query().Has("username") {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}
	username := req.URL.Query().Get("username")

	if err := performSetupCheck(w, &state.SetupStatuses, username); err != nil {
		fmt.Printf("Error getting draw stats for user \"%s\": %s\n", username, err)
		return
	}

	queryStr := `
  SELECT
    tc.time_class,
    COALESCE(SUM(CASE WHEN g.result = 'repetition' THEN 1 ELSE NULL END), 0) as repetitions,
    COALESCE(SUM(CASE WHEN g.result = 'insufficient' THEN 1 ELSE NULL END), 0) as insufficients,
    COALESCE(SUM(CASE WHEN g.result = 'timevsinsufficient' THEN 1 ELSE NULL END), 0) as timeoutVsInsufficients,
    COALESCE(SUM(CASE WHEN g.result = 'stalemate' THEN 1 ELSE NULL END), 0) as stalemates,
    COALESCE(SUM(CASE WHEN g.result = 'agreed' THEN 1 ELSE NULL END), 0) as agrees,
    COALESCE(SUM(CASE WHEN g.result = '50move' THEN 1 ELSE NULL END), 0) as fiftyMoveRules,
		COUNT(*) as total
  FROM (
    SELECT DISTINCT time_class FROM games
  ) tc 
  LEFT JOIN
    games g ON tc.time_class = g.time_class AND g.winner IS NULL
  GROUP BY tc.time_class
  `

	db := state.DBMap[utils.Hash(username)]
	if db == nil {
		fmt.Println("Error making draw stats query: db not found")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	db.Mu.Lock()
	defer db.Mu.Unlock()

	rows, err := db.Resource.Query(queryStr)
	if err != nil {
		fmt.Printf("Error making draw stats query: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	response := make(map[string]DrawStats)
	for rows.Next() {
		var timeClass string
		var repetitions int
		var insufficients int
		var timeoutVsInsufficients int
		var stalemates int
		var agrees int
		var fiftyMoveRules int
		var total int

		if err := rows.Scan(
			&timeClass,
			&repetitions,
			&insufficients,
			&timeoutVsInsufficients,
			&stalemates,
			&agrees,
			&fiftyMoveRules,
			&total,
		); err != nil {
			fmt.Printf("Error parsing draw stats query result: %s\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response[timeClass] = DrawStats{
			NumRepetitions:            repetitions,
			NumInsufficients:          insufficients,
			NumTimeoutVsInsufficients: timeoutVsInsufficients,
			NumStalemates:             stalemates,
			NumAgrees:                 agrees,
			Num50Rules:                fiftyMoveRules,
			Total:                     total,
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding draw stats query result: %s\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
