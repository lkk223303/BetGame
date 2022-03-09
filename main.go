package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

//////////// 抽獎遊戲 ///////////////////
type User struct {
	Id      string `json:"Id"`      // player ID
	Balance int    `json:"balance"` // money player have
}

type UserBet struct { 
	Id     string `json:Id`  
	Round  int    `json:Round` // 局數  
	Amount int    `json:Amount`  // 下注金額
}

const (
	RoundSecond    = 60               // Each round time
	DefaultBalance = 1000             // Everyone gets 1000 first
	UserMember     = "game"           // 儲存所有使用者的Balance  Redis:`Sorted-Set`	SCORE -> USER
	BetThisRound   = "bet_this_round" // 儲存目前局的下注狀況		 	Redis:`Sorted-Set`  SCORE -> USER
)

var Round := 0
var startTimeThisRound = time.Time
var RC *redis.Client


func init(){
	RC.newClient()
	
	// initialize all redis
	RC.Del(UserMember)
	RC.Del(BetThisGame)

	go GameServer()
}

func main() {

	router:= gin.Default()

	router.RedirectFixedPath = true
	router.GET("/bet/:user",GetUserBalance) // 玩家註冊（不須密碼，填入帳號即可）`user`區分大小寫
	router.GET("/bet/:user/:amount",Bet)  // 玩家對目前的局面進行下注，`amount`金額

	router.GET("/prize",GetCurrentPrize) // 此局目前的獎金池
	router.GET("/bet",GetUserBets) // 此局所有玩家目前的下注

	router.Run(":80")

	fmt.Println("Hello 每位玩家註冊後預設有1000塊。可以投注任意正整數金額,投越多錢、中獎機率越高。
	伺服器每分鐘會抽一位幸運者出來，並把這局所有的錢給予這名幸運者。")
	c := NewClient()
	test(c)

	
}

func newClient() *redis.Client { //redis.Client instance and return the addr.
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  //use default DB
	})

	pong, err := client.Ping(client.Context()).Result()
	log.Println(pong)
	if err != nil{
		log.Fatalln(err)
	}

	return client
}


func GameServer()  {
	rand.Seed(time.Now().UTC().UnixNano())
	
	ticker := time.NewTicker(RoundSecond * time.Second) // 每過RoundSecond (60) 秒，執行一次以下迴圈

	go func ()  {
		for {
			Round++
			startTimeThisRound = time.Now()
			log.Println(startTimeThisRound.Format("2022-02-08 18:18:18"),"\t round : ",Round,"start")
		}
		_ = <-ticker.C


		var prizePool = getCurrentPrize()
		var userBets = getUserBets()
		if len(userBets) == 0{
			log.Println("Round", Round , "沒有任何玩家下注")
			continue
		}


		// 抽獎選贏家
		winNum := rand.Intn(prizePool + 1)
		var winner string
		for _,userBet := range userBets{
			winNum = winNum - userBet.Amount

			if winNum <=0{
				winner = userBet.Id
				break
			}
		}  

		log.Println("獎金池:", prizePool, "\t 得主:", winner)

		// 發獎金給得主
		RC.ZIncrBy(UserMember, float64(prizePool),winner)
		
		
	}
}






///////////// Redis Test (用來測試 redis 是否連線成功)///////////////
// func NewClient() *redis.Client { //redis.Client instance and return the addr.
// 	client := redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "", // no password set
// 		DB:       0,  //use default DB
// 	})

// 	pong, err := client.Ping(client.Context()).Result()
// 	fmt.Println(pong, err)

// 	return client
// }

// func test(client *redis.Client) { // manupulate redis.Client
// 	err := client.Set(client.Context(), "key", "value", 0).Err() // => SET key value 0 means never expire.
// 	if err != nil {
// 		panic(err)
// 	}

// 	val, err := client.Get(client.Context(), "key").Result() // =>Get key
// 	if err != nil {
// 		panic(err)
// 	} else {
// 		fmt.Println("key", val) //key value
// 	}

// 	val2, err := client.Get(client.Context(), "key2").Result() // => GET key2
// 	if err == redis.Nil {
// 		fmt.Println("key2 does not exist")
// 	} else if err != nil {
// 		panic(err)
// 	} else {
// 		fmt.Println("key2", val2) /// key2 does not exist
// 	}

// }



