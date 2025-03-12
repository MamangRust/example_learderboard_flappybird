package main

import (
	"log"
	"net/http"
	"sort"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Player struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

var players = make(map[*websocket.Conn]Player)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		var player Player
		err := conn.ReadJSON(&player)
		if err != nil {
			log.Println("Read error:", err)
			delete(players, conn)
			break
		}

		players[conn] = player
		broadcastRanking(player)
	}
}

func broadcastRanking(currentPlayer Player) {
	ranking := getRanking(currentPlayer)
	for conn := range players {
		err := conn.WriteJSON(ranking)
		if err != nil {
			log.Println("Write error:", err)
			conn.Close()
			delete(players, conn)
		}
	}
}

func getRanking(player Player) map[string]interface{} {
	var playerList []Player
	for _, p := range players {
		playerList = append(playerList, p)
	}

	sort.Slice(playerList, func(i, j int) bool {
		return playerList[i].Score > playerList[j].Score
	})

	return map[string]interface{}{
		"rank":        getPlayerRank(playerList, player),
		"total":       len(playerList),
		"leaderboard": playerList,
	}
}

func getPlayerRank(playerList []Player, player Player) int {
	for i, p := range playerList {
		if p.Name == player.Name {
			return i + 1
		}
	}
	return -1
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
