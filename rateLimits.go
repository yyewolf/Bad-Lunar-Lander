package main

import (
	"time"
)

type rateLimit struct {
	lastMessage time.Time
}

func checkUser(id string) bool {
	val, err := rateLimitCache.Get(id)
	if !err {
		rl := &rateLimit{
			lastMessage: time.Now(),
		}
		rateLimitCache.Set(id, rl, 0)
		return false
	}
	if time.Now().Sub(val.(*rateLimit).lastMessage) <= time.Duration(commandRouter.RateLimit)*time.Millisecond {
		return true
	}
	rl := &rateLimit{
		lastMessage: time.Now(),
	}
	rateLimitCache.Set(id, rl, 0)
	return false
}
