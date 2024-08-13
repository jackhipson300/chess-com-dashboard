package model

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"

	"gopkg.in/freeeve/pgn.v1"
)

const insertBatchSize = 5000

type InsertStatistics struct {
	NumGamesInserted        int
	NumPositionsInserted    int
	NumGameInsertErrors     int
	NumPositionInsertErrors int
}

func CreateTables(db *sql.DB) {
	createGamesTable := `
	CREATE TABLE IF NOT EXISTS games (
		id TEXT PRIMARY KEY,
		url VARCHAR(255) NOT NULL,
		time_class VARCHAR(20) NOT NULL,
		time_control VARCHAR(25) NOT NULL,
		white_player VARCHAR(50) NOT NULL,
		black_player VARCHAR(50) NOT NULL,
		white_rating INTEGER NOT NULL,
		black_rating INTEGER NOT NULL,
		winner VARCHAR(50),
		result VARCHAR(25) NOT NULL
	)
	`

	createPositionsTable := `
	CREATE TABLE IF NOT EXISTS positions (
		id TEXT PRIMARY KEY,
		fen TEXT NOT NULL,
		game_id TEXT NOT NULL
	)
	`

	if _, err := db.Exec(createGamesTable); err != nil {
		fmt.Println("Error creating games table")
		panic(err)
	}
	if _, err := db.Exec(createPositionsTable); err != nil {
		fmt.Println("Error creating positions table")
		panic(err)
	}
}

func insertGame(tx *sql.Tx, stmt *sql.Stmt, game Game) error {
	var winner interface{} = nil
	result := game.WhitePlayer.Result
	if game.WhitePlayer.Result == "win" {
		winner = game.WhitePlayer.Username
		result = game.BlackPlayer.Result
	} else if game.BlackPlayer.Result == "win" {
		winner = game.BlackPlayer.Username
		result = game.WhitePlayer.Result
	}

	_, err := tx.Stmt(stmt).Exec(
		game.Id,
		game.Url,
		game.TimeClass,
		game.TimeControl,
		game.WhitePlayer.Username,
		game.BlackPlayer.Username,
		game.WhitePlayer.Rating,
		game.BlackPlayer.Rating,
		winner,
		result,
	)
	if err != nil {
		return fmt.Errorf("insert game error: %w", err)
	}

	return nil
}

func insertFens(tx *sql.Tx, stmt *sql.Stmt, pgnStr string, gameId string) (int, int) {
	numPositionsInserted := 0
	numPositionInsertErrors := 0

	ps := pgn.NewPGNScanner(strings.NewReader(pgnStr))
	for ps.Next() {
		game, err := ps.Scan()
		if err != nil {
			numPositionInsertErrors++
			continue
		}

		b := pgn.NewBoard()
		for _, move := range game.Moves {
			if err := b.MakeMove(move); err != nil {
				continue
			}
			fenParts := strings.Split(b.String(), " ")
			fenStr := strings.Join(fenParts[:len(fenParts)-2], " ")

			hash := sha256.New()
			hash.Write([]byte(fenStr))
			hash.Write([]byte(gameId))

			if _, err := tx.Stmt(stmt).Exec(hash.Sum(nil), fenStr, gameId); err != nil {
				numPositionInsertErrors++
				continue
			}

			numPositionsInserted++
		}
	}

	return numPositionsInserted, numPositionInsertErrors
}

func InsertUserData(db *sql.DB, allGames []Game) (InsertStatistics, error) {
	numGamesInserted := 0
	numPositionsInserted := 0
	numGameInsertErrors := 0
	numPositionInsertErrors := 0

	tx, err := db.Begin()
	if err != nil {
		return InsertStatistics{}, fmt.Errorf("error starting initial transaction: %w", err)
	}

	gameInsertStmt, err := db.Prepare(`
	INSERT OR IGNORE INTO games (
		id, 
		url, 
		time_class, 
		time_control, 
		white_player, 
		black_player, 
		white_rating,
		black_rating,
		winner,
		result
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return InsertStatistics{}, fmt.Errorf("error preparing games insert: %w", err)
	}
	defer gameInsertStmt.Close()
	fenInsertStmt, err := db.Prepare("INSERT OR IGNORE INTO positions (id, fen, game_id) VALUES(?, ?, ?)")
	if err != nil {
		return InsertStatistics{}, fmt.Errorf("error preparing positions insert: %w", err)
	}
	defer fenInsertStmt.Close()

	for i, game := range allGames {
		fmt.Printf("%d / %d games inserted\r", i+1, len(allGames))
		if strings.Contains(game.Pgn, "[Variant \"") {
			// variants tend to break pgn parser
			continue
		}

		if err := insertGame(tx, gameInsertStmt, game); err != nil {
			numGameInsertErrors++
		} else {
			numGamesInserted++
		}

		currNumPositionsInserted, currNumPositionInsertErrors := insertFens(tx, fenInsertStmt, game.Pgn, game.Id)
		numPositionsInserted += currNumPositionsInserted
		numPositionInsertErrors += currNumPositionInsertErrors

		if i%insertBatchSize == 0 {
			if err := tx.Commit(); err != nil {
				return InsertStatistics{}, fmt.Errorf("error committing transaction: %w", err)
			}

			tx, err = db.Begin()
			if err != nil {
				return InsertStatistics{}, fmt.Errorf("error beginning transaction: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return InsertStatistics{}, fmt.Errorf("error committing final transaction: %w", err)
	}

	fmt.Println("Indexing db...")
	createPositionsIndex := `
	CREATE INDEX IF NOT EXISTS fen_idx ON positions(fen)
	`
	if _, err := db.Exec(createPositionsIndex); err != nil {
		return InsertStatistics{}, fmt.Errorf("error creating positions index: %w", err)
	}

	return InsertStatistics{
    NumGamesInserted: numGamesInserted,
    NumGameInsertErrors: numGameInsertErrors,
    NumPositionsInserted: numPositionsInserted,
    NumPositionInsertErrors: numPositionInsertErrors,
	}, nil
}
