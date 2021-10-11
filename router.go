package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type cmdContext struct {
	s *discordgo.Session

	ID        string
	ChannelID string
	GuildID   string

	Arguments []*CommandArg

	Author        *discordgo.User
	isInteraction bool
	isComponent   bool
	Menu          *Menus

	ComponentData discordgo.MessageComponentInteractionData
	Interaction   *discordgo.Interaction
	Message       *discordgo.Message
}

type Choice struct {
	Name  string
	Value string
}

type CommandArg struct {
	Name  string
	Value interface{}
}

type Arg struct {
	Name        string
	Description string
	Size        int
	Required    bool
	Choices     []*Choice
	Type        discordgo.ApplicationCommandOptionType
}

type cmdAlias []string

func (a cmdAlias) Has(alias string) bool {
	for _, str := range a {
		if str == alias {
			return true
		}
	}
	return false
}

type Command struct {
	Name        string
	Description string
	Type        discordgo.ApplicationCommandType
	Aliases     cmdAlias
	Menu        menuName

	Args        []Arg
	SubCommands []*Command

	Call func(*cmdContext)
}

type router struct {
	Prefix         string
	ListenerPrefix string
	//RateLimit in milliseconds
	RateLimit int

	Commands []*Command
}

type Menus struct {
	MenuID        string
	Source        *discordgo.MessageEmbed
	SourceContext *cmdContext

	Call func(*cmdContext)
}

func (r *router) findTopCommand(name string) *Command {
	for _, cmd := range r.Commands {
		if cmd.Name == name || cmd.Aliases.Has(name) {
			return cmd
		}
	}
	return nil
}

func (c *Command) findDeepestLink(args []string) (*Command, []string) {
	if len(c.SubCommands) == 0 {
		return c, args
	} else {
		if len(args) == 0 {
			return c, args
		}
		for _, sub := range c.SubCommands {
			if args[0] == sub.Name {
				test, args := sub.findDeepestLink(args[1:])
				if test != nil {
					return test, args
				}
			}
		}
		return nil, args
	}
}

func slicer(data *discordgo.ApplicationCommandInteractionDataOption, args []string) ([]string, []*discordgo.ApplicationCommandInteractionDataOption) {
	// Tout pareil que interactionToSlice
	args = append(args, data.Name)
	if len(data.Options) == 0 {
		return args, []*discordgo.ApplicationCommandInteractionDataOption{}
	}
	if len(data.Options) > 1 {
		return args, data.Options
	}
	if data.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return args, data.Options
	}
	return slicer(data.Options[0], args)
}

func interactionToSlice(data *discordgo.ApplicationCommandInteractionData) ([]string, []*discordgo.ApplicationCommandInteractionDataOption) {
	// Initialise les arguments de la fonction
	args := []string{data.Name}
	// Cas où la commande n'a pas d'option
	if len(data.Options) == 0 {
		return args, []*discordgo.ApplicationCommandInteractionDataOption{}
	}
	// Cas où la commande a plus d'une option (arguments)
	if len(data.Options) > 1 {
		return args, data.Options
	}
	// Cas où la commande a une option (sous commande)
	if data.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return args, data.Options
	}
	// On va dans la sous-commande
	return slicer(data.Options[0], args)
}

func routeMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	/*
		Is Command or for Listener :
	*/
	if !strings.HasPrefix(m.Content, commandRouter.Prefix) {
		if !strings.HasPrefix(m.Content, commandRouter.ListenerPrefix) {
			return
		}
		/*
			Is for Listener :
		*/
		data, _, found, callback := getDataFromCache(m.ChannelID)
		if !found {
			return
		}
		/*
			Create context :
		*/
		ctx := &listenerContext{
			s:         s,
			ID:        m.ID,
			GuildID:   m.GuildID,
			ChannelID: m.ChannelID,
			Author:    m.Author,
			Message:   m.Message,
			Data:      data,
		}
		callback(ctx)
		return
	}
	/*
		Create context :
	*/
	ctx := &cmdContext{
		s:         s,
		ID:        m.ID,
		GuildID:   m.GuildID,
		ChannelID: m.ChannelID,
		Author:    m.Author,
		Message:   m.Message,
	}
	/*
		Rate limits :
	*/
	rateLimited := checkUser(ctx.Author.ID)
	if rateLimited {
		s.MessageReactionAdd(ctx.ChannelID, ctx.ID, "⌛")
		return
	}
	/*
		Find command & args :
	*/
	m.Content = m.ContentWithMentionsReplaced()
	m.Content = strings.TrimSpace(m.Content)
	m.Content = strings.TrimPrefix(m.Content, commandRouter.Prefix)
	splt := strings.Split(m.Content, " ")
	if len(splt) == 0 {
		return
	}
	topCmd := splt[0]
	cmd := commandRouter.findTopCommand(topCmd)
	if cmd == nil {
		r := regexp.MustCompile("[^a-zA-Z]+")
		topCmd := r.ReplaceAllString(topCmd, "")
		if topCmd != "" {
			ctx.reply(replyParams{
				Content: fmt.Sprintf("%s, cette commande n'existe pas (`%s`).", ctx.Author.Mention(), topCmd),
			})
		} else {
			ctx.reply(replyParams{
				Content: fmt.Sprintf("%s, cette commande n'existe pas encore.", ctx.Author.Mention()),
			})
		}
		return
	}
	deepestLink, argsLeft := cmd.findDeepestLink(splt[1:])

	var realArgs []*CommandArg
	for _, cmdArg := range deepestLink.Args {
		i := 0
		if i >= len(argsLeft) {
			break
		}
		if cmdArg.Size > 1 {
			current := &CommandArg{
				Name:  cmdArg.Name,
				Value: "",
			}
			for j := i; j < i+cmdArg.Size && j < len(argsLeft); j++ {
				if j != i+cmdArg.Size-1 {
					current.Value = current.Value.(string) + argsLeft[j] + " "
				} else {
					current.Value = current.Value.(string) + argsLeft[j]
				}
			}
			realArgs = append(realArgs, current)
		} else {
			current := &CommandArg{
				Name:  cmdArg.Name,
				Value: argsLeft[i],
			}
			realArgs = append(realArgs, current)
			i++
		}
	}

	ctx.Arguments = realArgs

	deepestLink.Call(ctx)
}

func routeInteraction(s *discordgo.Session, interaction *discordgo.InteractionCreate) {
	/*
		Verify type :
	*/
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}
	/*
		Create context :
	*/
	ctx := &cmdContext{
		s:         s,
		ID:        interaction.ID,
		GuildID:   interaction.GuildID,
		ChannelID: interaction.ChannelID,

		Interaction:   interaction.Interaction,
		isInteraction: true,
	}

	if interaction.Member != nil {
		ctx.Author = interaction.Member.User
	} else {
		ctx.Author = interaction.User
	}

	/*
		Find command & args :
	*/
	data := interaction.ApplicationCommandData()

	splt, parsedArgs := interactionToSlice(&data)
	if len(splt) == 0 {
		return
	}
	topCmd := splt[0]
	cmd := commandRouter.findTopCommand(topCmd)
	if cmd == nil {
		return
	}
	deepestLink, _ := cmd.findDeepestLink(splt[1:])

	var realArgs []*CommandArg
	for _, arg := range parsedArgs {
		for _, cmdArg := range deepestLink.Args {
			if arg.Name == cmdArg.Name {
				realArgs = append(realArgs, &CommandArg{
					Name:  arg.Name,
					Value: arg.Value,
				})
			}
		}
	}

	ctx.Arguments = realArgs
	deepestLink.Call(ctx)
}
