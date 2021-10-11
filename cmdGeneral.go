package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func ping(ctx *cmdContext) {
	ctx.reply(replyParams{
		Content:   "Pong!",
		Ephemeral: true,
	})
}

func letsplay(ctx *cmdContext) {
	// Start a game but doesn't handle it
	playerGame := &game{
		ID:       ctx.ID,
		PlayerID: ctx.Author.ID,
	}
	playerGame.genGrid()
	txt := playerGame.gridToText()
	msg, _ := ctx.reply(replyParams{
		Content: &discordgo.MessageSend{
			Embed: &discordgo.MessageEmbed{
				Description: txt,
			},
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							CustomID: playerGame.ID + "-G",
							Emoji: discordgo.ComponentEmoji{
								Name: "⬅️",
							},
						},
						&discordgo.Button{
							CustomID: playerGame.ID + "-D",
							Emoji: discordgo.ComponentEmoji{
								Name: "➡️",
							},
						},
					},
				},
			},
		},
	})
	playerGame.MessageID = msg.ID
	gamesCache.Set(ctx.Author.ID, playerGame, 0)
	activeMenus[ctx.ID] = &Menus{
		MenuID:        ctx.ID,
		SourceContext: ctx,
		Call:          letsmove,
	}
	playerGame.loop(ctx)
}

func letsmove(ctx *cmdContext) {
	ctx.s.InteractionRespond(ctx.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	action := strings.Split(ctx.ComponentData.CustomID, "-")[1]
	gameI, found := gamesCache.Get(ctx.Author.ID)
	if !found {
		return
	}

	g := gameI.(*game)
	switch action {
	case "G":
		g.move(ctx, -1, 0)
		break
	case "D":
		g.move(ctx, 1, 0)
		break
	}
}

func (g *game) loop(ctx *cmdContext) {
	ticker := time.NewTicker(2 * time.Second)
	g.LoopChan = make(chan int)
	go func() {
		for {
			select {
			case <-ticker.C:
				g.move(ctx, 0, 1)
			case <-g.LoopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (g *game) genGrid() {
	grid := [15][13]int{}
	playerspawn := false
	for y, line := range grid {
		for x := range line {
			if y == 0 && !playerspawn {
				rng := rand.Float32()
				if rng < 0.2 {
					grid[y][x] = void
					g.PlayerX = x
					g.PlayerY = y
					playerspawn = true
				} else {
					grid[y][x] = void
				}
			} else if y < 7 {
				grid[y][x] = void
			} else if y == 7 {
				rng := rand.Float32()
				if rng < 0.2 {
					grid[y][x] = floor
				} else {
					grid[y][x] = void
				}
			} else if y < 14 {
				rng := rand.Float32()
				if grid[y-1][x] == floor {
					grid[y][x] = floor
				} else if rng < 0.7 {
					if x-1 > 0 {
						if grid[y][x-1] == floor {
							grid[y][x] = floor
							continue
						}
					}
					if x+1 < 13 {
						if grid[y][x+1] == floor {
							grid[y][x] = floor
							continue
						}
					}
					grid[y][x] = void
				} else {
					grid[y][x] = void
				}
			} else {
				grid[y][x] = floor
			}
		}
	}
	g.Grid = grid
}

func (g *game) gridToText() string {
	text := ""
	for y, line := range g.Grid {
		for x := range line {
			if g.Finished && y == 4 {
				if x == 4 {
					text += "🇫"
				} else if x == 6 {
					text += "🇮"
				} else if x == 8 {
					text += "🇳"
				} else {
					text += emojis[g.Grid[y][x]]
				}
				continue
			}
			if g.PlayerX == x && g.PlayerY == y {
				if g.Finished && !g.Win {
					text += "💥"
				} else {
					text += emojis[player]
				}
			} else if x == g.PlayerX && y == g.PlayerY+1 && g.Grid[y][x] != floor {
				if g.Finished && !g.Win {
					text += "💥"
				} else {
					text += "<a:r_:897068111650500609>"
				}
			} else if g.Grid[y][x] == void {
				rng := rand.Float32()
				if rng < 0.158 {
					text += "⭐"
				} else {
					text += emojis[g.Grid[y][x]]
				}
			} else {
				text += emojis[g.Grid[y][x]]
			}
		}
		text += "\n"
	}
	return text
}

func (g *game) getCase(x, y int) int {
	if x > 0 && x < 13 && y > 0 && y < 15 {
		return g.Grid[y][x]
	}
	return void
}

func (g *game) verif(lastX, lastY, X, Y int) bool {
	nouvellecase := g.getCase(X, Y)
	gauche := g.getCase(lastX-1, lastY)
	if nouvellecase == void {
		return true
	}
	if lastX == X {
		if nouvellecase == floor {
			if gauche == floor {
				g.lose()
				return false
			} else {
				g.win()
				return false
			}
		}
	} else {
		if nouvellecase == floor {
			g.lose()
			return false
		}
	}
	return true
}

func (g *game) win() {
	g.Win = true
	g.Finished = true
}

func (g *game) lose() {
	g.Win = false
	g.Finished = true
}

func (g *game) move(ctx *cmdContext, X, Y int) {
	if ctx.isComponent {
		ctx.Menu.SourceContext.ID = ctx.Message.ID
		ctx = ctx.Menu.SourceContext
	}
	if g.PlayerX+X < 0 {
		g.PlayerX = 13
	}
	if g.PlayerX+X > 12 {
		g.PlayerX = -1
	}
	isValid := g.verif(g.PlayerX, g.PlayerY, g.PlayerX+X, g.PlayerY+Y)
	if isValid {
		g.PlayerY += Y
		g.PlayerX += X

		ctx.reply(replyParams{
			Content: &discordgo.MessageSend{
				Embed: &discordgo.MessageEmbed{
					Description: g.gridToText(),
				},
				Components: []discordgo.MessageComponent{
					&discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							&discordgo.Button{
								CustomID: g.ID + "-G",
								Emoji: discordgo.ComponentEmoji{
									Name: "⬅️",
								},
							},
							&discordgo.Button{
								CustomID: g.ID + "-D",
								Emoji: discordgo.ComponentEmoji{
									Name: "➡️",
								},
							},
						},
					},
				},
			},
			Edit: true,
			ID:   g.MessageID,
		})
	} else {
		ctx.reply(replyParams{
			Content: &discordgo.MessageSend{
				Embed: &discordgo.MessageEmbed{
					Description: g.gridToText(),
				},
				Components: []discordgo.MessageComponent{
					&discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							&discordgo.Button{
								CustomID: g.ID + "-G",
								Emoji: discordgo.ComponentEmoji{
									Name: "⬅️",
								},
							},
							&discordgo.Button{
								CustomID: g.ID + "-D",
								Emoji: discordgo.ComponentEmoji{
									Name: "➡️",
								},
							},
						},
					},
				},
			},
			Edit: true,
			ID:   g.MessageID,
		})
		if g.Win {
			ctx.reply(replyParams{
				Content: "GG, you win",
			})
			delete(activeMenus, ctx.ID)
			g.LoopChan <- 1
		} else {
			ctx.reply(replyParams{
				Content: "You lose",
			})
			delete(activeMenus, ctx.ID)
			g.LoopChan <- 1
		}
	}
}
