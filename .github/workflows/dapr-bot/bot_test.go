package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v55/github"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBot = &Bot{
	ctx:         context.Background(),
	issueClient: &testClient{},
}

type testClient struct {
	issue        *github.Issue
	issueComment *github.IssueComment
	resp         *github.Response
}

func (tc *testClient) CreateComment(ctx context.Context, org, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	return tc.issueComment, tc.resp, nil
}

func (tc *testClient) AddAssignees(ctx context.Context, org, repo string, number int, assignees []string) (*github.Issue, *github.Response, error) {
	return tc.issue, tc.resp, nil
}

func TestNewBot(t *testing.T) {
	t.Run("create a bot test", func(t *testing.T) {
		bot := NewBot(github.NewClient(nil))
		assert.NotNil(t, bot)
	})
}

func TestHandleEvent(t *testing.T) {
	t.Run("handle valid event", func(t *testing.T) {
		tc := testClient{
			resp: &github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
		}
		testBot.issueClient = &tc
		ctx := context.Background()
		var testEventCopy Event
		errC := copier.CopyWithOption(&testEventCopy, &testEvent, copier.Option{DeepCopy: true})
		if errC != nil {
			t.Error(errC)
		}
		testEventCopy.IssueCommentEvent.Comment.Body = github.String("/assign")
		res, err := testBot.HandleEvent(ctx, testEventCopy)
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("handle valid (longer body) event", func(t *testing.T) {
		tc := testClient{
			resp: &github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
		}
		testBot.issueClient = &tc
		ctx := context.Background()
		var testEventCopy Event
		errC := copier.CopyWithOption(&testEventCopy, &testEvent, copier.Option{DeepCopy: true})
		if errC != nil {
			t.Error(errC)
		}
		testEventCopy.IssueCommentEvent.Comment.Body = github.String("/assign \r \ntest body")
		res, err := testBot.HandleEvent(ctx, testEventCopy)
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("handle unable to assign", func(t *testing.T) {
		tc := testClient{
			resp: &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}},
		}
		testBot.issueClient = &tc
		ctx := context.Background()
		var testEventCopy Event
		errC := copier.CopyWithOption(&testEventCopy, &testEvent, copier.Option{DeepCopy: true})
		if errC != nil {
			t.Error(errC)
		}
		testEventCopy.IssueCommentEvent.Comment.Body = github.String("/assign")
		res, err := testBot.HandleEvent(ctx, testEventCopy)
		require.Error(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("handle no event", func(t *testing.T) {
		tc := testClient{}
		testBot.issueClient = &tc
		ctx := context.Background()
		var testEventCopy Event
		errC := copier.CopyWithOption(&testEventCopy, &testEvent, copier.Option{DeepCopy: true})
		if errC != nil {
			t.Error(errC)
		}
		testEventCopy.IssueCommentEvent.Comment.Body = github.String("assign")
		res, err := testBot.HandleEvent(ctx, testEventCopy)
		require.NoError(t, err)
		assert.Equal(t, "no command found", res)
	})
}

func TestCreateIssueComment(t *testing.T) {
	t.Run("failure to create issue comment", func(t *testing.T) {
		tc := testClient{
			resp: &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}},
		}
		testBot.issueClient = &tc
		err := testBot.CreateIssueComment("test", testEvent)
		require.Error(t, err)
	})

	t.Run("create issue comment", func(t *testing.T) {
		tc := testClient{
			resp: &github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
		}
		testBot.issueClient = &tc
		err := testBot.CreateIssueComment("test", testEvent)
		require.NoError(t, err)
	})

	t.Run("create issue comment with empty body", func(t *testing.T) {
		tc := testClient{
			resp: &github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
		}
		testBot.issueClient = &tc
		err := testBot.CreateIssueComment("", testEvent)
		require.Error(t, err)
	})
}

func TestAssignIssueToCommenter(t *testing.T) {
	t.Run("failure to assign issue to commenter", func(t *testing.T) {
		tc := testClient{
			resp: &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}},
		}
		testBot.issueClient = &tc
		assignee, err := testBot.AssignIssueToCommenter(testEvent)
		require.Error(t, err)
		assert.Empty(t, assignee)
	})

	t.Run("successfully assign issue to commenter", func(t *testing.T) {
		tc := testClient{
			resp: &github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
		}
		testBot.issueClient = &tc
		var testEventCopy Event
		errC := copier.CopyWithOption(&testEventCopy, &testEvent, copier.Option{DeepCopy: true})
		if errC != nil {
			t.Error(errC)
		}
		testEventCopy.IssueCommentEvent.Issue.Assignees = []*github.User{}
		assignee, err := testBot.AssignIssueToCommenter(testEventCopy)
		require.NoError(t, err)
		assert.Equal(t, "testCommentLogin", assignee)
	})

	t.Run("attempt to assign a closed issue", func(t *testing.T) {
		tc := testClient{}
		testBot.issueClient = &tc
		var testEventCopy Event
		errC := copier.CopyWithOption(&testEventCopy, &testEvent, copier.Option{DeepCopy: true})
		if errC != nil {
			t.Error(errC)
		}
		testEventCopy.IssueCommentEvent.Issue.State = github.String("closed")
		assignee, err := testBot.AssignIssueToCommenter(testEventCopy)
		require.Error(t, err)
		assert.Empty(t, assignee)
	})

	t.Run("issue already assigned to user", func(t *testing.T) {
		tc := testClient{}
		testBot.issueClient = &tc
		var testEventCopy Event
		errC := copier.CopyWithOption(&testEventCopy, &testEvent, copier.Option{DeepCopy: true})
		if errC != nil {
			t.Error(errC)
		}
		testEventCopy.IssueCommentEvent.Issue.Assignees = []*github.User{{Login: github.String("testCommentLogin")}}
		assignee, err := testBot.AssignIssueToCommenter(testEventCopy)
		require.Error(t, err)
		assert.Empty(t, assignee)
	})

	t.Run("issue already assigned to another user", func(t *testing.T) {
		tc := testClient{
			resp: &github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
		}
		testBot.issueClient = &tc
		var testEventCopy Event
		errC := copier.CopyWithOption(&testEventCopy, &testEvent, copier.Option{DeepCopy: true})
		if errC != nil {
			t.Error(errC)
		}
		testEventCopy.IssueCommentEvent.Issue.Assignees = []*github.User{{Login: github.String("testCommentLogin2")}}
		assignee, err := testBot.AssignIssueToCommenter(testEventCopy)
		require.Error(t, err)
		assert.Empty(t, assignee)
	})
}
