package main

import (
	"encoding/json"
	"errors"

	"github.com/google/go-github/v55/github"
)

type Event struct {
	Type              string                    `json:"type"`
	Path              string                    `json:"path"`
	IssueCommentEvent *github.IssueCommentEvent `json:"issue_comment_event"`
}

func ProcessEvent(eventType string, eventPath string, data []byte) (e Event, err error) {
	var issueCommentEvent *github.IssueCommentEvent
	if eventPath == "" {
		return Event{}, errors.New("invalid event path")
	}
	switch eventType {
	case "issue_comment":
		err = json.Unmarshal(data, &issueCommentEvent)
		if err != nil {
			return Event{}, err
		}
	}
	e = Event{
		Type:              eventType,
		Path:              eventPath,
		IssueCommentEvent: issueCommentEvent,
	}
	return
}

func (e *Event) GetIssueAssignees() []string {
	assignees := make([]string, 0)
	for _, assignee := range e.IssueCommentEvent.Issue.Assignees {
		assignees = append(assignees, assignee.GetLogin())
	}
	return assignees
}

func (e *Event) GetIssueNumber() int {
	return e.IssueCommentEvent.Issue.GetNumber()
}

func (e *Event) GetIssueOrg() string {
	return e.IssueCommentEvent.Repo.Owner.GetLogin()
}

func (e *Event) GetIssueRepo() string {
	return e.IssueCommentEvent.Repo.GetName()
}

func (e *Event) GetIssueState() string {
	return e.IssueCommentEvent.Issue.GetState()
}

func (e *Event) GetIssueUser() string {
	return e.IssueCommentEvent.Comment.User.GetLogin()
}
