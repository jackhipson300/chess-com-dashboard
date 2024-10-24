package model

import (
	"backend/types"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const insertBatchSize = 5000

type InsertStatistics struct {
	NumGamesInserted        int
	NumPositionsInserted    int
	NumGameInsertErrors     int
	NumPositionInsertErrors int
}

func CreateTables(db *sql.DB) (err error) {
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

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT,
		latest_archive TEXT
	)
	`

	if _, err := db.Exec(createGamesTable); err != nil {
		return fmt.Errorf("error creating games table: %w", err)
	}
	if _, err := db.Exec(createPositionsTable); err != nil {
		return fmt.Errorf("error creating positions table: %w", err)
	}
	if _, err := db.Exec(createUsersTable); err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	return
}

func insertGame(tx *sql.Tx, gameStmt *sql.Stmt, fenStmt *sql.Stmt, game Game) (numPositionsInserted int, numPositionInsertErrors int, err error) {
	var winner interface{} = nil
	result := game.WhitePlayer.Result
	if game.WhitePlayer.Result == "win" {
		winner = game.WhitePlayer.Username
		result = game.BlackPlayer.Result
	} else if game.BlackPlayer.Result == "win" {
		winner = game.BlackPlayer.Username
		result = game.WhitePlayer.Result
	}

	_, err = tx.Stmt(gameStmt).Exec(
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
		err = fmt.Errorf("insert game error: %w", err)
		return
	}

	for _, fen := range game.Fens {
		hash := sha256.New()
		hash.Write([]byte(fen))
		hash.Write([]byte(game.Id))
		if _, err := tx.Stmt(fenStmt).Exec(hash.Sum(nil), fen, game.Id); err != nil {
			numPositionInsertErrors++
			continue
		}
		numPositionsInserted++
	}

	return
}

func InsertUserData(db *sql.DB, userId string, username string, allGames []Game, archives []string) (statistics InsertStatistics, err error) {
	numGamesInserted := 0
	numPositionsInserted := 0
	numGameInsertErrors := 0
	numPositionInsertErrors := 0

	tx, err := db.Begin()
	if err != nil {
		return statistics, fmt.Errorf("error starting initial transaction: %w", err)
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
		return statistics, fmt.Errorf("error preparing games insert: %w", err)
	}
	defer gameInsertStmt.Close()
	fenInsertStmt, err := db.Prepare("INSERT OR IGNORE INTO positions (id, fen, game_id) VALUES(?, ?, ?)")
	if err != nil {
		return statistics, fmt.Errorf("error preparing positions insert: %w", err)
	}
	defer fenInsertStmt.Close()

	for i, game := range allGames {
		fmt.Printf("%d / %d games inserted\r", i+1, len(allGames))
		if strings.Contains(game.Pgn, "[Variant \"") {
			// variants tend to break pgn parser
			continue
		}

		currNumPositionsInserted, currNumPositionInsertErrors, err := insertGame(tx, gameInsertStmt, fenInsertStmt, game)
		if err != nil {
			numGameInsertErrors++
		} else {
			numGamesInserted++
		}
		numPositionsInserted += currNumPositionsInserted
		numPositionInsertErrors += currNumPositionInsertErrors

		if i%insertBatchSize == 0 {
			if err := tx.Commit(); err != nil {
				return statistics, fmt.Errorf("error committing transaction: %w", err)
			}

			tx, err = db.Begin()
			if err != nil {
				return statistics, fmt.Errorf("error beginning transaction: %w", err)
			}
		}
	}
	fmt.Println()

	if err := tx.Commit(); err != nil {
		return statistics, fmt.Errorf("error committing final transaction: %w", err)
	}

	fmt.Println("Indexing db...")
	createPositionsIndex := `
	CREATE INDEX IF NOT EXISTS fen_idx ON positions(fen)
	`
	if _, err := db.Exec(createPositionsIndex); err != nil {
		return statistics, fmt.Errorf("error creating positions index: %w", err)
	}

	if len(archives) > 0 {
		mostRecentArchive := archives[len(archives)-1]
		insertUserStmt := "INSERT OR IGNORE INTO users (id, username, latest_archive) VALUES(?, ?, ?)"
		_, err = db.Exec(insertUserStmt, userId, username, mostRecentArchive)
		if err != nil {
			return statistics, fmt.Errorf("error inserting user entry: %w", err)
		}
	}

	return InsertStatistics{
		NumGamesInserted:        numGamesInserted,
		NumGameInsertErrors:     numGameInsertErrors,
		NumPositionsInserted:    numPositionsInserted,
		NumPositionInsertErrors: numPositionInsertErrors,
	}, nil
}

func LoadExistingDbs(dbMap *types.DBMap) (userIds []string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error loading existing dbs: %w", err)
	}

	pattern := filepath.Join(cwd, "*.db")
	existingDbs, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("error loading existing dbs: %w", err)
	}

	for _, filepath := range existingDbs {
		filepathParts := strings.Split(filepath, "/")
		filenameWithExtension := filepathParts[len(filepathParts)-1]
		dbFilename := fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL", filenameWithExtension)
		db, err := sql.Open("sqlite3", dbFilename)
		if err != nil {
			return nil, fmt.Errorf("error loading existing dbs: %w", err)
		}

		if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			return nil, fmt.Errorf("error enabling WAL: %w", err)
		}

		filenameWithoutExtension := filenameWithExtension[0 : len(filenameWithExtension)-3]
		(*dbMap)[filenameWithoutExtension] = types.NewLockedDB(db)
		userIds = append(userIds, filenameWithoutExtension)
	}

	return
}

func GetMostRecentArchive(userId string, db *sql.DB) (archive string, err error) {
	queryStr := `
	SELECT latest_archive
	FROM users
	WHERE id = $1
	`

	rows, err := db.Query(queryStr, userId)
	if err != nil {
		return
	}

	if !rows.Next() {
		err = fmt.Errorf("no user query result")
		return
	}

	err = rows.Scan(&archive)

	return
}

/*
	when loading initial dbs, set the setupstatus to pending
	when the user calls setup, check if it is pending
	if it is,
	get the latest archive in storage,
	if it is the current month or less, fetch the new archives
		note that when we do this, we should always consider the most recent stored archive as
		incomplete and replace it
	we need to be able to avoid duplicate position entries because of this
*/
