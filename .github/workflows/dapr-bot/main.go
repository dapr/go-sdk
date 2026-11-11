package main

import (
	"context"
	"log"
	"os"

	"github.com/google/go-github/v55/github"
)

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
	data, err := os.ReadFile(eventPath) //nolint:gosec // eventPath is provided by the GitHub Actions runner
	if err != nil {
		log.Fatalf("failed to read event: %v", err)
	}
	event, err := ProcessEvent(eventType, eventPath, data)
	if err != nil {
		log.Fatalf("failed to process event: %v", err)
	}
	log.Printf("processing event: %s", event.Type) //nolint:gosec // event.Type is the GitHub Actions event name from GITHUB_EVENT_NAME

	res, err := bot.HandleEvent(ctx, event)
	if err != nil {
		log.Fatalf("failed to handle event: %v", err)
	}
	log.Println(res) //nolint:gosec // res is a bot-generated status string, not user-controlled
}
