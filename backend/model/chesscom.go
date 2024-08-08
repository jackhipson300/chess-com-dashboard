package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
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
		fmt.Println("Error requesting archives")
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error parsing archives list body")
		panic(err)
	}

	var data ArchivesData
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Error parsing archives list json")
		panic(err)
	}

	return data.Archives
}

func getArchive(url string, wg *sync.WaitGroup, ch chan<- []Game) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error requesting archive")
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error parsing archive body")
		panic(err)
	}

	var data Archive
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Error parsing archive json")
		panic(err)
	}

	ch <- data.Games
}

func GetAllGames(user string) []Game {
	fmt.Println("Requesting list of archives...")
	archives := listArchives(user)
	var wg sync.WaitGroup

	fmt.Println("Requesting individual archives...")
	games_ch := make(chan []Game, len(archives))
	for _, archive := range archives {
		wg.Add(1)
		go getArchive(archive, &wg, games_ch)
	}

	wg.Wait()
	close(games_ch)

	fmt.Println("Receiving individual archives...")
	allGames := []Game{}
	for games := range games_ch {
		allGames = append(allGames, games...)
	}

	return allGames
}
