package main

import (
	"context"
	"os"

	"github.com/google/go-github/v55/github"

	"github.com/dapr/kit/logger"
)

var log = logger.NewLogger("dapr.github.workflows")

func main() {
	ctx := context.Background()
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN is required")
	}

	ghClient := github.NewClient(nil).WithAuthToken(githubToken)
	bot := NewBot(ghClient)
	eventType := os.Getenv("GITHUB_EVENT_NAME")
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	data, err := os.ReadFile(eventPath)
	if err != nil {
		log.Fatalf("failed to read event: %v", err)
	}
	event, err := ProcessEvent(eventType, eventPath, data)
	if err != nil {
		log.Fatalf("failed to process event: %v", err)
	}
	log.Infof("processing event: %s", event.Type)

	res, err := bot.HandleEvent(ctx, event)
	if err != nil {
		log.Fatalf("failed to handle event: %v", err)
	}
	log.Info(res)
}
