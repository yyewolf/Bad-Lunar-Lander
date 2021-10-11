package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func choiceConvert(c []*Choice) (r []*discordgo.ApplicationCommandOptionChoice) {
	for _, choice := range c {
		r = append(r, &discordgo.ApplicationCommandOptionChoice{
			Name:  choice.Name,
			Value: choice.Value,
		})
	}
	return
}

func (c *Command) makeOption() (opts []*discordgo.ApplicationCommandOption) {
	if len(c.SubCommands) == 0 {
		for _, arg := range c.Args {
			opt := &discordgo.ApplicationCommandOption{
				Name:        arg.Name,
				Description: arg.Description,
				Type:        arg.Type,
				Required:    arg.Required,
				Choices:     choiceConvert(arg.Choices),
			}
			opts = append(opts, opt)
		}
	} else {
		for _, sub := range c.SubCommands {
			opt := &discordgo.ApplicationCommandOption{
				Name:        sub.Name,
				Description: sub.Description,
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options:     sub.makeOption(),
			}
			opts = append(opts, opt)
		}
	}
	return
}

func (c *Command) make() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name,
		Type:        c.Type,
		Description: c.Description,
		Options:     c.makeOption(),
	}
}

func (r *router) getSlashCommands() (out []*discordgo.ApplicationCommand) {
	for _, cmd := range r.Commands {
		if cmd.Menu != "" {
			out = append(out, cmd.make())
		}
	}
	return
}

func (r *router) loadSlashCommands(s *discordgo.Session) {
	cmds := r.getSlashCommands()

	for _, cmd := range cmds {
		_, err := s.ApplicationCommandCreate(appID, "", cmd)
		if err != nil {
			fmt.Printf("Cannot create '%v' : %v\n", cmd.Name, err.Error())
		} else {
			fmt.Printf("Created '%v' \n", cmd.Name)
		}
	}

	dcmds, _ := s.ApplicationCommands(appID, "")
	for _, dcmd := range dcmds {
		remove := true
		for _, botcmd := range cmds {
			if botcmd.Name == dcmd.Name && botcmd.Type == dcmd.Type {
				remove = false
				break
			}
		}
		if remove {
			s.ApplicationCommandDelete(appID, "", dcmd.ID)
			fmt.Printf("Removed '%v' \n", dcmd.Name)
		}
	}
}
