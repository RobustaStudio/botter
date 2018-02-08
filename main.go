package main

import "fmt"
import "log"
import "flag"

var (
	HTTP_ADDR = flag.String("http", ":80", "the http address to listen on")
	BOT_FILE  = flag.String("bot", "", "chatbot html script")
	REDIS_DSN = flag.String("redis", "redis://localhost:6379/10", "redis dsn")
)

func main() {
	fmt.Println("Initializing the chatbot server")
	flag.Parse()
	redis, err := InitRedisClient(*REDIS_DSN)
	if err != nil {
		log.Fatal("[REDIS]", err)
	}
	tree, err := CompileFile(*BOT_FILE)
	if err != nil {
		log.Fatal("[SYNTEX]", err)
	}
	log.Fatal("[API]", RunAPIServer(NewManager(redis, tree)))
}
