package main

import (
	"fmt"
	"math"
	"time"
)

func unixFormatter(timestamp int64) string {
	t := timestamp - time.Now().Unix()
	H := int(float64(t) / float64(3600))
	M := int(math.Mod(float64(t)/float64(60), 60))
	S := int(math.Mod(float64(t), 60))

	return fmt.Sprintf("**%02d hours %02d minutes %02d seconds**", H, M, S)
}

func unixNanoFormatter(timestamp int64) string {
	t := timestamp - time.Now().UnixNano()
	t = t / 1e9
	H := int(float64(t) / float64(3600))
	M := int(math.Mod(float64(t)/float64(60), 60))
	S := int(math.Mod(float64(t), 60))

	return fmt.Sprintf("**%02d hours %02d minutes %02d seconds**", H, M, S)
}
