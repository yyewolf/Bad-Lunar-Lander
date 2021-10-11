package main

import (
	"sort"

	"github.com/bwmarrin/discordgo"
)

type menuName string

const (
	GeneralMenu   menuName = "Générales"
	UtilitiesMenu menuName = "Utilitaires"
)

func menuEmoji(name string) string {
	switch name {
	case string(GeneralMenu):
		return "🖥️"
	case string(UtilitiesMenu):
		return "🧰"
	}
	return ""
}

var helpMenus map[string][]*Command
var menuEmbed map[string]*discordgo.MessageEmbed

func makeEmbed() {
	menuEmbed = make(map[string]*discordgo.MessageEmbed)
	for menuName, cmds := range helpMenus {
		embed := &discordgo.MessageEmbed{
			Title: menuName + " :",
			Color: botColor,
		}
		for i, cmd := range cmds {
			if i != 0 {
				embed.Description += "\n"
			}
			embed.Description += "`" + commandRouter.Prefix + cmd.Name + " : " + cmd.Description + "`"
		}
		menuEmbed[menuName] = embed
	}
}

func helpComponent(menuID string, defaultMenu string) []discordgo.MessageComponent {
	var opts []discordgo.SelectMenuOption

	var keys []string
	for k := range menuEmbed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		opt := discordgo.SelectMenuOption{
			Label: name,
			Value: name,
			Emoji: discordgo.ComponentEmoji{
				Name: menuEmoji(name),
			},
		}
		if name == defaultMenu {
			opt.Default = true
		}
		opts = append(opts, opt)
	}

	cmp := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.SelectMenu{
					CustomID: menuID,
					Options:  opts,
				},
			},
		},
	}

	return cmp
}

func help(ctx *cmdContext) {
	menu := string(GeneralMenu)
	menuID := ctx.ID
	if ctx.isComponent {
		ctx.s.InteractionRespond(ctx.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		if ctx.Author.ID != ctx.Menu.SourceContext.Author.ID {
			return
		}
		menu = ctx.ComponentData.Values[0]
		menuID = ctx.Menu.MenuID
	}
	complex := &discordgo.MessageSend{
		Embed:      menuEmbed[menu],
		Components: helpComponent(menuID, menu),
	}
	complex.Embed.Footer = defaultFooter

	if ctx.isComponent {
		// Keep old context if a button is pressed
		ctx.Menu.SourceContext.ID = ctx.Message.ID
		ctx = ctx.Menu.SourceContext
		ctx.isComponent = true
	} else {
		activeMenus[ctx.ID] = &Menus{
			MenuID:        ctx.ID,
			SourceContext: ctx,
			Call:          help,
		}
	}

	ctx.reply(replyParams{
		Content: complex,
		ID:      ctx.ID,
		Edit:    ctx.isComponent,
	})
}
