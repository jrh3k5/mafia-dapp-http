package main

import (
	"math/rand"
	"time"

	"github.com/jrh3k5/mafia-dapp-http/server"
)

func main() {
	// initialize random seed for shuffling player assignments
	rand.Seed(time.Now().UnixNano())

	server.NewServer().Run("0.0.0.0:3000")
}
