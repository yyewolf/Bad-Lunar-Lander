package main

import "math/rand"

func ping(ctx *cmdContext) {
	ctx.reply(replyParams{
		Content:   "Pong!",
		Ephemeral: true,
	})
}

const (
	void = iota
	floor
	player
)

func genGrid() [20][20]int {
	grid := [20][20]int{}
	playerspawn := false
	for y, line := range grid {
		for x := range line {
			if y == 0 && !playerspawn {
				rng := rand.Float32()
				if rng < 0.2 {
					grid[y][x] = player
					playerspawn = true
				} else {
					grid[y][x] = void
				}
			} else if y < 15 {
				grid[y][x] = void
			} else if y == 15 {
				rng := rand.Float32()
				if rng < 0.2 {
					grid[y][x] = floor
				} else {
					grid[y][x] = void
				}
			} else if y < 19 {
				rng := rand.Float32()
				if grid[y-1][x] == floor {
					grid[y][x] = floor
				} else if rng < 0.5 {
					if x-1 > 0 {
						if grid[y][x-1] == floor {
							grid[y][x] = floor
							continue
						}
					}
					if x+1 < 20 {
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
	return grid
}
