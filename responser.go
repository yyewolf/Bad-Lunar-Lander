package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type replyParams struct {
	Content     interface{}
	Components  []discordgo.MessageComponent
	Interaction *discordgo.Interaction
	ID          string
	ChannelID   string
	GuildID     string

	DM        bool
	Edit      bool
	Delete    bool
	FollowUp  bool
	Ephemeral bool
}

func (c *cmdContext) reply(p replyParams) (st *discordgo.Message, err error) {
	if p.DM {
		channel, err := c.s.UserChannelCreate(c.Author.ID)
		if err != nil {
			p.ChannelID = c.ChannelID
			p.Content = "Sorry <@!" + c.Author.ID + ">, but I cannot contact you through DMs, check your privacy settings!"
		} else {
			p.ChannelID = channel.ID
		}
	}

	if p.ChannelID == "" {
		p.ChannelID = c.ChannelID
	}
	if p.GuildID == "" {
		p.GuildID = c.GuildID
	}

	if c.isInteraction {
		return c.replyInteraction(p)
	}
	return c.replyClassic(p)
}

func (c *cmdContext) replyClassic(p replyParams) (st *discordgo.Message, err error) {
	if p.Delete {
		err = c.s.ChannelMessageDelete(p.ChannelID, p.ID)
		return
	}
	switch p.Content.(type) {
	case string:
		if p.Edit {
			return c.s.ChannelMessageEdit(p.ChannelID, p.ID, fmt.Sprint(p.Content))
		}
		if len(p.Components) == 0 {
			return c.s.ChannelMessageSend(p.ChannelID, fmt.Sprint(p.Content))
		} else {
			return c.s.ChannelMessageSendComplex(p.ChannelID, &discordgo.MessageSend{
				Content:    fmt.Sprint(p.Content),
				Components: p.Components,
			})
		}
	case *discordgo.MessageEmbed:
		if p.Edit {
			v := &discordgo.MessageEdit{
				Embed:      p.Content.(*discordgo.MessageEmbed),
				Components: p.Components,

				ID:      p.ID,
				Channel: p.ChannelID,
			}
			return c.s.ChannelMessageEditComplex(v)
		}
		v := &discordgo.MessageSend{
			Embed:      p.Content.(*discordgo.MessageEmbed),
			Components: p.Components,
		}
		return c.s.ChannelMessageSendComplex(p.ChannelID, v)
	case *discordgo.MessageSend:
		if p.Edit {
			complex := p.Content.(*discordgo.MessageSend)
			v := &discordgo.MessageEdit{
				Content:    &complex.Content,
				Embed:      complex.Embed,
				Components: complex.Components,

				ID:      p.ID,
				Channel: p.ChannelID,
			}
			return c.s.ChannelMessageEditComplex(v)
		}
		return c.s.ChannelMessageSendComplex(p.ChannelID, p.Content.(*discordgo.MessageSend))
	case *discordgo.MessageEdit:
		return c.s.ChannelMessageEditComplex(p.Content.(*discordgo.MessageEdit))
	default:
		fmt.Println("unknown")
	}
	return
}

func (c *cmdContext) replyInteraction(p replyParams) (st *discordgo.Message, err error) {
	var flags uint64
	if p.Ephemeral {
		flags = 1 << 6
	}
	if p.Delete {
		err = c.s.InteractionResponseDelete(p.ChannelID, c.Interaction)
		return
	}
	switch p.Content.(type) {
	case string:
		if !p.FollowUp {
			if p.Edit {
				return c.s.InteractionResponseEdit(appID, c.Interaction, &discordgo.WebhookEdit{
					Content:    p.Content.(string),
					Components: p.Components,
				})
			}
			err = c.s.InteractionRespond(c.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:      flags,
					Content:    p.Content.(string),
					Components: p.Components,
				},
			})
			return
		} else {
			if p.Edit {
				return c.s.FollowupMessageEdit(appID, c.Interaction, p.ID, &discordgo.WebhookEdit{
					Content:    p.Content.(string),
					Components: p.Components,
				})
			}
			return c.s.FollowupMessageCreate(appID, c.Interaction, true, &discordgo.WebhookParams{
				Content:    p.Content.(string),
				Components: p.Components,
				Flags:      flags,
			})
		}
	case *discordgo.MessageEmbed:
		if !p.FollowUp {
			if p.Edit {
				return c.s.InteractionResponseEdit(appID, c.Interaction, &discordgo.WebhookEdit{
					Embeds: []*discordgo.MessageEmbed{
						p.Content.(*discordgo.MessageEmbed),
					},
					Components: p.Components,
				})
			}
			err = c.s.InteractionRespond(c.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: flags,
					Embeds: []*discordgo.MessageEmbed{
						p.Content.(*discordgo.MessageEmbed),
					},
					Components: p.Components,
				},
			})
			return
		} else {
			if p.Edit {
				return c.s.FollowupMessageEdit(appID, c.Interaction, p.ID, &discordgo.WebhookEdit{
					Embeds: []*discordgo.MessageEmbed{
						p.Content.(*discordgo.MessageEmbed),
					},
					Components: p.Components,
				})
			}
			return c.s.FollowupMessageCreate(appID, c.Interaction, true, &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{
					p.Content.(*discordgo.MessageEmbed),
				},
				Components: p.Components,
				Flags:      flags,
			})
		}
	case *discordgo.MessageSend:
		complex := p.Content.(*discordgo.MessageSend)
		if !p.FollowUp {
			if p.Edit {
				return c.s.InteractionResponseEdit(appID, c.Interaction, &discordgo.WebhookEdit{
					Content: complex.Content,
					Embeds: []*discordgo.MessageEmbed{
						complex.Embed,
					},
					Components: complex.Components,
				})
			}
			err = c.s.InteractionRespond(c.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   flags,
					Content: complex.Content,
					Embeds: []*discordgo.MessageEmbed{
						complex.Embed,
					},
					Components: complex.Components,
				},
			})
			return
		} else {
			if p.Edit {
				return c.s.FollowupMessageEdit(appID, c.Interaction, p.ID, &discordgo.WebhookEdit{
					Content: complex.Content,
					Embeds: []*discordgo.MessageEmbed{
						complex.Embed,
					},
					Components: complex.Components,
				})
			}
			return c.s.FollowupMessageCreate(appID, c.Interaction, true, &discordgo.WebhookParams{
				Content: complex.Content,
				Embeds: []*discordgo.MessageEmbed{
					complex.Embed,
				},
				Components: complex.Components,
				Flags:      flags,
			})
		}
	case *discordgo.MessageEdit:
		complex := p.Content.(*discordgo.MessageEdit)
		if !p.FollowUp {
			return c.s.InteractionResponseEdit(appID, c.Interaction, &discordgo.WebhookEdit{
				Content: *complex.Content,
				Embeds: []*discordgo.MessageEmbed{
					complex.Embed,
				},
				Components: complex.Components,
			})
		} else {
			return c.s.FollowupMessageEdit(appID, c.Interaction, p.ID, &discordgo.WebhookEdit{
				Content: *complex.Content,
				Embeds: []*discordgo.MessageEmbed{
					complex.Embed,
				},
				Components: complex.Components,
			})
		}
	default:
		fmt.Println("unknown")
	}
	return
}
