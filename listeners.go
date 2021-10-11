package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pmylund/go-cache"
)

type Listener struct {
	Cache    *cache.Cache
	Callback func(*listenerContext)
}

type listenerContext struct {
	s *discordgo.Session

	ID        string
	ChannelID string
	GuildID   string

	Author *discordgo.User

	Data interface{}

	Message *discordgo.Message
}

func isChannelInCache(id string) bool {
	_, found := listenersCache.Get(id)
	return found
}

func getDataFromCache(id string) (interface{}, time.Time, bool, func(*listenerContext)) {
	val, found := listenersCache.Get(id)
	if !found {
		return val, time.Time{}, found, nil
	}
	d, expire, found := val.(Listener).Cache.GetWithExpiration(id)
	return d, expire, found, val.(Listener).Callback
}

func addDataToCache(id string, cache *cache.Cache, callback func(*listenerContext), data interface{}) {
	listenersCache.Set(id, Listener{
		Cache:    cache,
		Callback: callback,
	}, 0)
	cache.Set(id, data, 0)
}

func (l *listenerContext) reply(p replyParams) (st *discordgo.Message, err error) {
	ctx := &cmdContext{
		s:         l.s,
		ID:        l.ID,
		ChannelID: l.ChannelID,
		GuildID:   l.GuildID,
		Author:    l.Author,
		Message:   l.Message,
	}
	return ctx.reply(p)
}
