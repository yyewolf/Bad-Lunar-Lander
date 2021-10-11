package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	defines()

	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())

	grid := genGrid()
	for _, line := range grid {
		fmt.Println(line)
	}

	commandRouter = router{
		Prefix: "%",
		// ListenerPrefix: "<",
		RateLimit: 2000,
	}

	loadCmd()

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Failed creating session")
	}
	sess = s

	s.AddHandler(routeMessages)
	s.AddHandler(routeInteraction)
	s.AddHandler(routeComponents)
	s.AddHandler(botReady)
	s.Identify.Intents = discordgo.IntentsAllWithoutPrivileged + 1<<1

	log.Println("Starting the shard manager")
	err = s.Open()
	if err != nil {
		log.Fatal("Failed to start: ", err)
	}

	commandRouter.loadSlashCommands(s)

	defaultFooter = &discordgo.MessageEmbedFooter{
		IconURL: s.State.User.AvatarURL("256"),
		Text:    "Hackin'TN : La stégano c'est nul.",
	}

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	s.Close()
}

func botReady(s *discordgo.Session, evt *discordgo.Ready) {
	s.UpdateGameStatus(0, commandRouter.Prefix+"help for help (Shard : "+strconv.Itoa(s.ShardID+1)+")")
}
