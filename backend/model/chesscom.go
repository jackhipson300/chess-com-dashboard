package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"gopkg.in/freeeve/pgn.v1"
)

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

type RawGame struct {
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

type Game struct {
	RawGame
	Fens []string
}

type Archive struct {
	Games []RawGame `json:"games"`
}

func ListArchives(user string) (archives []string, err error) {
	fmt.Println("Requesting list of archives...")
	url := fmt.Sprintf("http://api.chess.com/pub/player/%s/games/archives", user)
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("error requesting archives: %w", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error parsing archives list body: %w", err)
		return
	}

	var data ArchivesData
	if err = json.Unmarshal(body, &data); err != nil {
		err = fmt.Errorf("error parsing archives list json: %w", err)
		return
	}

	archives = data.Archives
	return
}

func getArchive(url string, wg *sync.WaitGroup, ch chan<- []Game) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error requesting archive")
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error parsing archive body")
		return
	}

	var data Archive
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Error parsing archive json")
		return
	}

	var games []Game
	for _, game := range data.Games {
		games = append(games, parseGame(&game))
	}

	ch <- games
}

func parseGame(rawGame *RawGame) Game {
	var fens []string

	// variants tend to break pgn parser
	if !strings.Contains(rawGame.Pgn, "[Variant \"") {
		ps := pgn.NewPGNScanner(strings.NewReader(rawGame.Pgn))
		for ps.Next() {
			game, err := ps.Scan()
			if err != nil {
				continue
			}

			b := pgn.NewBoard()
			for _, move := range game.Moves {
				if err := b.MakeMove(move); err != nil {
					continue
				}
				fenParts := strings.Split(b.String(), " ")
				fenStr := strings.Join(fenParts[:len(fenParts)-2], " ")

				fens = append(fens, fenStr)
			}
		}
	}

	rawGame.WhitePlayer.Username = strings.ToLower(rawGame.WhitePlayer.Username)
	rawGame.BlackPlayer.Username = strings.ToLower(rawGame.BlackPlayer.Username)

	return Game{
		RawGame: *rawGame,
		Fens:    fens,
	}
}

func GetAllGames(archives []string) []Game {
	var wg sync.WaitGroup

	fmt.Println("Requesting games...")
	gamesCh := make(chan []Game, len(archives))
	for _, archive := range archives {
		wg.Add(1)
		go getArchive(archive, &wg, gamesCh)
	}

	wg.Wait()
	close(gamesCh)

	var allGames []Game
	for games := range gamesCh {
		allGames = append(allGames, games...)
	}

	return allGames
}
