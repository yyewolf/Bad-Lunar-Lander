package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/bwmarrin/snowflake"
	"github.com/pmylund/go-cache"
)

// DGO
var sess *discordgo.Session
var commandRouter router
var activeMenus map[string]*Menus
var defaultFooter *discordgo.MessageEmbedFooter

var rateLimitCache *cache.Cache
var listenersCache *cache.Cache
var gamesCache *cache.Cache

// UUID
var node *snowflake.Node

// Game definitions
const (
	void = iota
	floor
	player
)

var emojis = []string{"♣️", "🧀", "🛸"}

type game struct {
	ID        string
	MessageID string
	PlayerID  string
	Grid      [15][13]int
	Moves     int
	PlayerX   int
	PlayerY   int
	Finished  bool
	Win       bool
	LoopChan  chan int
}

func defines() {
	rateLimitCache = cache.New(5*time.Minute, 10*time.Minute)
	listenersCache = cache.New(5*time.Minute, 10*time.Minute)
	gamesCache = cache.New(5*time.Minute, 10*time.Minute)
	helpMenus = make(map[string][]*Command)
	activeMenus = make(map[string]*Menus)
	node, _ = snowflake.NewNode(1)
}
