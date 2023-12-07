package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"

	"github.com/google/go-github/v55/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testEvent = Event{
	Type: "issue_comment",
	Path: "test/test",
	IssueCommentEvent: &github.IssueCommentEvent{
		Issue: &github.Issue{
			Assignees: []*github.User{{Login: github.String("testAssignee")}},
			Number:    github.Int(123),
			State:     github.String("testState"),
		},
		Repo: &github.Repository{
			Owner: &github.User{Login: github.String("testOrg")},
			Name:  github.String("testName"),
		},
		Comment: &github.IssueComment{
			User: &github.User{Login: github.String("testCommentLogin")},
		},
	},
}

func TestProcessEvent(t *testing.T) {
	testEventData, err := json.Marshal(testEvent)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}
	t.Run("process event", func(t *testing.T) {
		event, err := ProcessEvent(testEvent.Type, testEvent.Path, testEventData)
		require.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, "test/test", event.Path)
	})

	t.Run("process event with empty path", func(t *testing.T) {
		event, err := ProcessEvent(testEvent.Type, "", testEventData)
		require.Error(t, err)
		assert.Empty(t, event)
	})

	var randomData bytes.Buffer
	encoder := gob.NewEncoder(&randomData)
	encoder.Encode("random_data")

	t.Run("process issue_comment event", func(t *testing.T) {
		event, err := ProcessEvent(testEvent.Type, testEvent.Path, testEventData)
		require.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, "issue_comment", event.Type)
	})

	t.Run("process invalid event", func(t *testing.T) {
		event, err := ProcessEvent(testEvent.Type, testEvent.Path, randomData.Bytes())
		require.Error(t, err)
		assert.Empty(t, event)
	})
}

func TestGetIssueAssignees(t *testing.T) {
	t.Run("get assignees", func(t *testing.T) {
		assignees := testEvent.GetIssueAssignees()
		assert.Len(t, assignees, 1)
		assert.Equal(t, "testAssignee", assignees[0])
	})
}

func TestGetIssueNumber(t *testing.T) {
	t.Run("get issue number", func(t *testing.T) {
		number := testEvent.GetIssueNumber()
		assert.Equal(t, 123, number)
	})
}

func TestGetIssueOrg(t *testing.T) {
	t.Run("get issue org", func(t *testing.T) {
		org := testEvent.GetIssueOrg()
		assert.Equal(t, "testOrg", org)
	})
}

func TestGetIssueRepo(t *testing.T) {
	t.Run("get issue repo", func(t *testing.T) {
		repo := testEvent.GetIssueRepo()
		assert.Equal(t, "testName", repo)
	})
}

func TestGetIssueState(t *testing.T) {
	t.Run("get issue state", func(t *testing.T) {
		state := testEvent.GetIssueState()
		assert.Equal(t, "testState", state)
	})
}

func TestGetIssueUser(t *testing.T) {
	t.Run("get issue user", func(t *testing.T) {
		user := testEvent.GetIssueUser()
		assert.Equal(t, "testCommentLogin", user)
	})
}
