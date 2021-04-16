package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/OpenSlides/openslides-vote-service/internal/vote"
)

func main() {
	ctx, cancel := interruptContext()
	defer cancel()

	if err := vote.Run(ctx, os.Environ(), log.Printf); err != nil {
		log.Printf("Error: %v", err)
	}
}

// interruptContext works like signal.NotifyContext
//
// In only listens on os.Interrupt. If the signal is received two times,
// os.Exit(1) is called.
func interruptContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		cancel()

		// If the signal was send for the second time, make a hard cut.
		<-sigint
		os.Exit(1)
	}()
	return ctx, cancel
}
