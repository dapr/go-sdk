package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v55/github"
)

var (
	errCommentBodyEmpty     = errors.New("comment body is empty")
	errIssueClosed          = errors.New("issue is closed")
	errIssueAlreadyAssigned = errors.New("issue is already assigned")
	errUnauthorizedClient   = errors.New("possibly unauthorized client issue")
)

type issueInterface interface {
	CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	AddAssignees(ctx context.Context, owner string, repo string, number int, assignees []string) (*github.Issue, *github.Response, error)
}

type Bot struct {
	ctx         context.Context
	issueClient issueInterface
}

func NewBot(ghClient *github.Client) *Bot {
	return &Bot{
		ctx:         context.Background(),
		issueClient: ghClient.Issues,
	}
}

func (b *Bot) HandleEvent(ctx context.Context, event Event) (res string, err error) {
	commentBody := event.IssueCommentEvent.Comment.GetBody()

	// split the comment after any potential new lines
	newline := strings.Split(strings.ReplaceAll(commentBody, "\r\n", "\n"), "\n")[0]

	command := strings.Split(newline, " ")[0]

	if command[0] != '/' {
		return "no command found", err
	}

	switch command {
	case "/assign":
		assignee, err := b.AssignIssueToCommenter(event)
		res = "👍 Issue assigned to " + assignee
		if err == nil {
			err = b.CreateIssueComment("🚀 Issue assigned to you @"+assignee, event)
		} else {
			err = b.CreateIssueComment("⚠️ Unable to assign issue", event)
		}
		if err != nil {
			return fmt.Sprintf("failed to comment on issue: %v", err), err
		}
	}
	return
}

func (b *Bot) CreateIssueComment(body string, event Event) error {
	if body == "" {
		return errCommentBodyEmpty
	}
	ctx := context.Background()
	comment := &github.IssueComment{
		Body: github.String(body),
	}
	_, response, err := b.issueClient.CreateComment(ctx, event.GetIssueOrg(), event.GetIssueRepo(), event.GetIssueNumber(), comment)
	if err != nil || response.StatusCode == http.StatusNotFound {
		return fmt.Errorf("failed to create comment: %v%v", err, response.StatusCode)
	}
	return nil
}

func (b *Bot) AssignIssueToCommenter(event Event) (string, error) {
	if event.GetIssueState() == "closed" {
		return "", errIssueClosed
	}

	if len(event.GetIssueAssignees()) > 0 {
		return "", errIssueAlreadyAssigned
	}

	ctx := context.Background()
	_, response, err := b.issueClient.AddAssignees(ctx, event.GetIssueOrg(), event.GetIssueRepo(), event.GetIssueNumber(), []string{event.GetIssueUser()})
	if response.StatusCode == http.StatusNotFound {
		return "", errUnauthorizedClient
	}
	return event.GetIssueUser(), err
}
