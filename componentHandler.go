package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func routeComponents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	split := strings.Split(i.MessageComponentData().CustomID, "-")
	m, ok := activeMenus[split[0]]
	if !ok {
		return
	}
	ctx := &cmdContext{
		s:         s,
		ID:        i.ID,
		ChannelID: i.ChannelID,
		GuildID:   i.GuildID,

		isComponent:   true,
		isInteraction: true,
		Menu:          m,
		ComponentData: i.MessageComponentData(),
		Interaction:   i.Interaction,
		Message:       i.Interaction.Message,
	}
	if i.Member != nil {
		ctx.Author = i.Member.User
	} else {
		ctx.Author = i.User
	}

	m.MenuID = split[0]

	m.Call(ctx)
}
