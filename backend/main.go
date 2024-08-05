package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"gopkg.in/freeeve/pgn.v1"

	_ "github.com/mattn/go-sqlite3"
)

const insertBatchSize = 5000

type ArchivesData struct {
	Archives []string `json:"archives"`
}

type GamePlayer struct {
	Id       string `json:"uuid"`
	Url      string `json:"@id"`
	Username string `json:"username"`
	Result   string `json:"result"`
	Rating   uint16 `json:"rating"`
}

type Game struct {
	Id          string     `json:"uuid"`
	Url         string     `json:"url"`
	Pgn         string     `json:"pgn"`
	TimeControl string     `json:"time_control"`
	EndTime     uint32     `json:"end_time"`
	IsRated     bool       `json:"rated"`
	TimeClass   string     `json:"time_class"`
	WhitePlayer GamePlayer `json:"white"`
	BlackPlayer GamePlayer `json:"black"`
}

type Archive struct {
	Games []Game `json:"games"`
}

func listArchives(user string) []string {
	url := fmt.Sprintf("http://api.chess.com/pub/player/%s/games/archives", user)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data ArchivesData
	if err := json.Unmarshal(body, &data); err != nil {
		panic(err)
	}

	return data.Archives
}

func getArchive(url string) []Game {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data Archive
	if err := json.Unmarshal(body, &data); err != nil {
		panic(err)
	}

	return data.Games
}

func createTables(db *sql.DB) {
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

func insertGame(tx *sql.Tx, stmt *sql.Stmt, game Game) {
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
		fmt.Printf("Insert game error\n")
		panic(err)
	}
}

func insertFens(tx *sql.Tx, stmt *sql.Stmt, pgnStr string, gameId string) {
	ps := pgn.NewPGNScanner(strings.NewReader(pgnStr))
	for ps.Next() {
		game, err := ps.Scan()
		if err != nil {
			fmt.Printf("Pgn error\n")
			fmt.Println(err)
			continue
		}

		b := pgn.NewBoard()
		for _, move := range game.Moves {
			b.MakeMove(move)
			fenParts := strings.Split(b.String(), " ")
			fenStr := strings.Join(fenParts[:len(fenParts)-2], " ")

			hash := sha256.New()
			hash.Write([]byte(fenStr))
			hash.Write([]byte(gameId))

			if _, err := tx.Stmt(stmt).Exec(hash.Sum(nil), fenStr, gameId); err != nil {
				fmt.Println("Fen insert error")
				panic(err)
			}
		}
	}
}

func insertUserData(db *sql.DB, allGames []Game) {
	start := time.Now()

	tx, err := db.Begin()
	if err != nil {
		panic(err)
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
		panic(err)
	}
	defer gameInsertStmt.Close()
	fenInsertStmt, err := db.Prepare("INSERT OR IGNORE INTO positions (id, fen, game_id) VALUES(?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer fenInsertStmt.Close()

	for i, game := range allGames {
		fmt.Printf("%d / %d\n", i+1, len(allGames))
		insertGame(tx, gameInsertStmt, game)
		insertFens(tx, fenInsertStmt, game.Pgn, game.Id)

		if i%insertBatchSize == 0 {
			if err := tx.Commit(); err != nil {
				panic(err)
			}

			tx, err = db.Begin()
			if err != nil {
				panic(err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}

	createPositionsIndex := `
	CREATE INDEX IF NOT EXISTS fen_idx ON positions(fen)
	`
	if _, err := db.Exec(createPositionsIndex); err != nil {
		fmt.Println("Error creating positions index")
		panic(err)
	}

	duration := time.Since(start)
	fmt.Printf("Inserted user data in %v\n", duration)
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	createTables(db)

	user := os.Getenv("USERNAME")
	archives := listArchives(user)

	allGames := []Game{}
	for _, archive := range archives {
		games := getArchive(archive)
		allGames = append(allGames, games...)
	}
	fmt.Printf("# games: %d\n", len(allGames))

	insertUserData(db, allGames)

	// rows, err := db.Query(`
	// SELECT fen, count(*)
	// FROM positions
	// GROUP BY fen
	// HAVING count(*) > 1
	// ORDER BY count(*)
	// `)
	// if err != nil {
	// 	fmt.Printf("Query error\n")
	// 	panic(err)
	// }
	// defer rows.Close()

	// for i := 0; rows.Next() && i < 20; i++ {
	// 	var fen string
	// 	var count int
	// 	if err := rows.Scan(&fen, &count); err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Printf("Fen: %s, Count: %d\n", fen, count)
	// }
}
