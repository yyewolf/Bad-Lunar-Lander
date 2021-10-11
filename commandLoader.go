package main

import "github.com/bwmarrin/discordgo"

func addCmd(cmd *Command) {
	if cmd.Type == 0 {
		cmd.Type = discordgo.ChatApplicationCommand
	}
	commandRouter.Commands = append(commandRouter.Commands, cmd)
	if cmd.Menu != "" {
		helpMenus[string(cmd.Menu)] = append(helpMenus[string(cmd.Menu)], cmd)
	}
}

func loadCmd() {
	ping := &Command{
		Name:        "ping",
		Description: "Recevoir Pong!",
		Aliases:     cmdAlias{"p"},
		Menu:        GeneralMenu,
		Call:        ping,
	}
	help := &Command{
		Name:        "help",
		Description: "Voir le menu d'aide.",
		Aliases:     cmdAlias{"h"},
		Menu:        UtilitiesMenu,
		Call:        help,
	}
	addCmd(help)
	addCmd(ping)

	makeEmbed()
}
