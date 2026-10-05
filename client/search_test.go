/*
Copyright 2026 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

// fakeSearch is the in-memory search store backing testDaprServer.
var fakeSearch = struct {
	sync.Mutex
	indexes     map[string]map[string]*pb.SearchDocument
	lastOptions *pb.IndexingOptionsAlpha1
	lastSearch  *pb.SearchRequestAlpha1
}{
	indexes: make(map[string]map[string]*pb.SearchDocument),
}

func fakeKey(storeName, name string) string {
	return storeName + "||" + name
}

func ackFor(opts *pb.IndexingOptionsAlpha1) pb.IndexAck {
	if opts.GetMode() == pb.IndexingMode_INDEXING_MODE_WAIT_FOR_COMPLETION {
		return pb.IndexAck_INDEX_ACK_COMPLETED
	}
	return pb.IndexAck_INDEX_ACK_QUEUED
}

func (s *testDaprServer) CreateIndexAlpha1(_ context.Context, in *pb.CreateIndexRequestAlpha1) (*emptypb.Empty, error) {
	fakeSearch.Lock()
	defer fakeSearch.Unlock()
	key := fakeKey(in.GetStoreName(), in.GetIndex())
	if _, ok := fakeSearch.indexes[key]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "index %s already exists", in.GetIndex())
	}
	fakeSearch.indexes[key] = make(map[string]*pb.SearchDocument)
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) GetIndexAlpha1(_ context.Context, in *pb.GetIndexRequestAlpha1) (*pb.GetIndexResponseAlpha1, error) {
	fakeSearch.Lock()
	defer fakeSearch.Unlock()
	docs, ok := fakeSearch.indexes[fakeKey(in.GetStoreName(), in.GetIndex())]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "index %s not found", in.GetIndex())
	}
	return &pb.GetIndexResponseAlpha1{
		Index:         in.GetIndex(),
		DocumentCount: uint64(len(docs)),
		Properties:    in.GetMetadata(),
	}, nil
}

func (s *testDaprServer) ListIndexesAlpha1(_ context.Context, in *pb.ListIndexesRequestAlpha1) (*pb.ListIndexesResponseAlpha1, error) {
	fakeSearch.Lock()
	defer fakeSearch.Unlock()
	var names []string
	for key := range fakeSearch.indexes {
		if store, name, _ := strings.Cut(key, "||"); store == in.GetStoreName() {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return &pb.ListIndexesResponseAlpha1{Indexes: names}, nil
}

func (s *testDaprServer) DeleteIndexAlpha1(_ context.Context, in *pb.DeleteIndexRequestAlpha1) (*emptypb.Empty, error) {
	fakeSearch.Lock()
	defer fakeSearch.Unlock()
	delete(fakeSearch.indexes, fakeKey(in.GetStoreName(), in.GetIndex()))
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) IndexDocumentsAlpha1(_ context.Context, in *pb.IndexDocumentsRequestAlpha1) (*pb.IndexDocumentsResponseAlpha1, error) {
	fakeSearch.Lock()
	defer fakeSearch.Unlock()
	fakeSearch.lastOptions = in.GetOptions()
	docs, ok := fakeSearch.indexes[fakeKey(in.GetStoreName(), in.GetIndex())]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "index %s not found", in.GetIndex())
	}
	resp := &pb.IndexDocumentsResponseAlpha1{Ack: ackFor(in.GetOptions())}
	for _, d := range in.GetDocuments() {
		var obj map[string]any
		if err := json.Unmarshal(d.GetContent(), &obj); err != nil {
			resp.FailedItems = append(resp.FailedItems, &pb.FailedItem{
				Id:    d.GetId(),
				Error: status.New(codes.InvalidArgument, "content is not a JSON object").Proto(),
			})
			continue
		}
		docs[d.GetId()] = proto.Clone(d).(*pb.SearchDocument)
	}
	return resp, nil
}

func (s *testDaprServer) GetDocumentsAlpha1(_ context.Context, in *pb.GetDocumentsRequestAlpha1) (*pb.GetDocumentsResponseAlpha1, error) {
	fakeSearch.Lock()
	defer fakeSearch.Unlock()
	docs := fakeSearch.indexes[fakeKey(in.GetStoreName(), in.GetIndex())]
	resp := &pb.GetDocumentsResponseAlpha1{}
	for _, id := range in.GetIds() {
		d, ok := docs[id]
		if !ok {
			continue
		}
		d = proto.Clone(d).(*pb.SearchDocument)
		if !in.GetIncludeContent() {
			d.Content = nil
		}
		resp.Documents = append(resp.Documents, d)
	}
	return resp, nil
}

func (s *testDaprServer) DeleteDocumentsAlpha1(_ context.Context, in *pb.DeleteDocumentsRequestAlpha1) (*pb.DeleteDocumentsResponseAlpha1, error) {
	fakeSearch.Lock()
	defer fakeSearch.Unlock()
	fakeSearch.lastOptions = in.GetOptions()
	docs := fakeSearch.indexes[fakeKey(in.GetStoreName(), in.GetIndex())]
	for _, id := range in.GetIds() {
		delete(docs, id)
	}
	return &pb.DeleteDocumentsResponseAlpha1{Ack: ackFor(in.GetOptions())}, nil
}

func (s *testDaprServer) SearchAlpha1(_ context.Context, in *pb.SearchRequestAlpha1) (*pb.SearchResponseAlpha1, error) {
	fakeSearch.Lock()
	defer fakeSearch.Unlock()
	fakeSearch.lastSearch = proto.Clone(in).(*pb.SearchRequestAlpha1)
	docs := fakeSearch.indexes[fakeKey(in.GetStoreName(), in.GetIndex())]

	ids := make([]string, 0, len(docs))
	for id, d := range docs {
		if bytes.Contains(d.GetContent(), []byte(in.GetText())) {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)

	total := uint64(len(ids))
	resp := &pb.SearchResponseAlpha1{
		TotalHits:         &total,
		TotalHitsRelation: pb.TotalHitsRelation_TOTAL_HITS_RELATION_EXACT,
	}
	if top := int(in.GetTopK()); top > 0 && len(ids) > top {
		ids = ids[:top]
		resp.ContinuationToken = "next"
	}
	for _, id := range ids {
		d := proto.Clone(docs[id]).(*pb.SearchDocument)
		if !in.GetIncludeContent() {
			d.Content = nil
		}
		resp.Hits = append(resp.Hits, &pb.SearchHit{
			Document:   d,
			Score:      1,
			Highlights: map[string]string{"title": "<em>" + in.GetText() + "</em>"},
		})
	}
	return resp, nil
}

func TestSearchAlpha1(t *testing.T) {
	ctx := t.Context()
	const store = "search-store"
	const index = "products"

	// TestMain runs the suite once per client transport; start from a clean store.
	fakeSearch.Lock()
	fakeSearch.indexes = make(map[string]map[string]*pb.SearchDocument)
	fakeSearch.Unlock()

	t.Run("validation", func(t *testing.T) {
		require.EqualError(t, testClient.CreateIndexAlpha1(ctx, "", index, nil), "storeName is empty")
		require.EqualError(t, testClient.CreateIndexAlpha1(ctx, store, "", nil), "index is empty")
		_, err := testClient.ListIndexesAlpha1(ctx, "", nil)
		require.EqualError(t, err, "storeName is empty")
		_, err = testClient.SearchAlpha1(ctx, store, index, nil)
		require.EqualError(t, err, "query is nil")
		_, err = testClient.SearchAlpha1(ctx, store, index, &SearchQuery{Text: "a", Native: map[string]any{"q": "a"}})
		require.EqualError(t, err, "query Text and Native are mutually exclusive")
		_, err = testClient.SearchAlpha1(ctx, store, index, &SearchQuery{Filter: map[string]any{"bad": make(chan int)}})
		require.ErrorContains(t, err, "invalid filter")
		_, err = testClient.IndexDocumentsAlpha1(ctx, store, index, []*SearchDocument{nil}, nil)
		require.EqualError(t, err, "document is nil")
	})

	t.Run("index lifecycle", func(t *testing.T) {
		require.NoError(t, testClient.CreateIndexAlpha1(ctx, store, index, nil))
		require.NoError(t, testClient.CreateIndexAlpha1(ctx, store, "other", nil))

		err := testClient.CreateIndexAlpha1(ctx, store, index, nil)
		require.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))

		indexes, err := testClient.ListIndexesAlpha1(ctx, store, nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"other", index}, indexes)

		require.NoError(t, testClient.DeleteIndexAlpha1(ctx, store, "other", nil))
		indexes, err = testClient.ListIndexesAlpha1(ctx, store, nil)
		require.NoError(t, err)
		assert.Equal(t, []string{index}, indexes)

		_, err = testClient.GetIndexAlpha1(ctx, store, "missing", nil)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("index documents", func(t *testing.T) {
		resp, err := testClient.IndexDocumentsAlpha1(ctx, store, index, []*SearchDocument{
			{ID: "1", Content: []byte(`{"title":"wireless headphones"}`), Metadata: map[string]string{"src": "a"}},
			{ID: "2", Content: []byte(`{"title":"wired headphones"}`)},
			{ID: "3", Content: []byte(`{"title":"wireless mouse"}`)},
			{ID: "bad", Content: []byte(`not json`)},
		}, nil, WithWaitForIndexCompletion(5*time.Second, IndexingWaitTimeoutContinueAsync))
		require.NoError(t, err)
		assert.Equal(t, IndexAckCompleted, resp.Ack)
		require.Len(t, resp.FailedItems, 1)
		assert.Equal(t, "bad", resp.FailedItems[0].ID)
		assert.Equal(t, codes.InvalidArgument, status.Code(resp.FailedItems[0].Err))

		fakeSearch.Lock()
		opts := fakeSearch.lastOptions
		fakeSearch.Unlock()
		assert.Equal(t, pb.IndexingMode_INDEXING_MODE_WAIT_FOR_COMPLETION, opts.GetMode())
		assert.Equal(t, 5*time.Second, opts.GetWaitTimeout().AsDuration())
		assert.Equal(t, pb.IndexingWaitTimeoutAction_INDEXING_WAIT_TIMEOUT_ACTION_CONTINUE_ASYNC, opts.GetOnWaitTimeout())

		idx, err := testClient.GetIndexAlpha1(ctx, store, index, nil)
		require.NoError(t, err)
		assert.Equal(t, index, idx.Name)
		assert.Equal(t, uint64(3), idx.DocumentCount)
	})

	t.Run("index documents with default options", func(t *testing.T) {
		resp, err := testClient.IndexDocumentsAlpha1(ctx, store, index, []*SearchDocument{
			{ID: "4", Content: []byte(`{"title":"speaker"}`)},
		}, nil)
		require.NoError(t, err)
		assert.Equal(t, IndexAckQueued, resp.Ack)
		assert.Empty(t, resp.FailedItems)

		fakeSearch.Lock()
		opts := fakeSearch.lastOptions
		fakeSearch.Unlock()
		assert.Nil(t, opts)
	})

	t.Run("get documents", func(t *testing.T) {
		docs, err := testClient.GetDocumentsAlpha1(ctx, store, index, []string{"2", "missing", "1"}, true, nil)
		require.NoError(t, err)
		require.Len(t, docs, 2)
		assert.Equal(t, "2", docs[0].ID)
		assert.Equal(t, "1", docs[1].ID)
		assert.JSONEq(t, `{"title":"wireless headphones"}`, string(docs[1].Content))
		assert.Equal(t, map[string]string{"src": "a"}, docs[1].Metadata)

		docs, err = testClient.GetDocumentsAlpha1(ctx, store, index, []string{"1"}, false, nil)
		require.NoError(t, err)
		require.Len(t, docs, 1)
		assert.Empty(t, docs[0].Content)
	})

	t.Run("search", func(t *testing.T) {
		resp, err := testClient.SearchAlpha1(ctx, store, index, &SearchQuery{
			Text:            "wireless",
			Filter:          map[string]any{"price": map[string]any{"$lt": 200}},
			TopK:            1,
			IncludeContent:  true,
			SearchFields:    []string{"title"},
			ReturnFields:    []string{"title"},
			HighlightFields: []string{"title"},
			Sort:            []SortClause{{Field: "price", Order: SortOrderDesc}},
			Metadata:        map[string]string{"k": "v"},
		})
		require.NoError(t, err)
		require.Len(t, resp.Hits, 1)
		assert.Equal(t, "1", resp.Hits[0].Document.ID)
		assert.NotEmpty(t, resp.Hits[0].Document.Content)
		assert.InDelta(t, 1.0, resp.Hits[0].Score, 0)
		assert.Equal(t, map[string]string{"title": "<em>wireless</em>"}, resp.Hits[0].Highlights)
		require.NotNil(t, resp.TotalHits)
		assert.Equal(t, uint64(2), *resp.TotalHits)
		assert.Equal(t, TotalHitsRelationExact, resp.TotalHitsRelation)
		assert.Equal(t, "next", resp.ContinuationToken)

		fakeSearch.Lock()
		req := fakeSearch.lastSearch
		fakeSearch.Unlock()
		assert.Equal(t, "wireless", req.GetText())
		assert.Equal(t, map[string]any{"price": map[string]any{"$lt": float64(200)}}, req.GetFilter().AsMap())
		assert.Equal(t, []string{"title"}, req.GetSearchFields())
		assert.Equal(t, []string{"title"}, req.GetReturnFields())
		assert.Equal(t, []string{"title"}, req.GetHighlightFields())
		require.Len(t, req.GetSort(), 1)
		assert.Equal(t, "price", req.GetSort()[0].GetField())
		assert.Equal(t, pb.SortOrder_SORT_ORDER_DESC, req.GetSort()[0].GetOrder())
		assert.Equal(t, map[string]string{"k": "v"}, req.GetMetadata())
	})

	t.Run("search native query", func(t *testing.T) {
		_, err := testClient.SearchAlpha1(ctx, store, index, &SearchQuery{
			Native:            map[string]any{"q": "wireless"},
			ContinuationToken: "next",
		})
		require.NoError(t, err)

		fakeSearch.Lock()
		req := fakeSearch.lastSearch
		fakeSearch.Unlock()
		assert.Equal(t, map[string]any{"q": "wireless"}, req.GetNative().AsMap())
		assert.Equal(t, "next", req.GetContinuationToken())
		assert.Nil(t, req.GetFilter())
	})

	t.Run("delete documents", func(t *testing.T) {
		ack, err := testClient.DeleteDocumentsAlpha1(ctx, store, index, []string{"1", "2"}, nil, WithReturnOnIndexAcceptance())
		require.NoError(t, err)
		assert.Equal(t, IndexAckQueued, ack)

		fakeSearch.Lock()
		opts := fakeSearch.lastOptions
		fakeSearch.Unlock()
		assert.Equal(t, pb.IndexingMode_INDEXING_MODE_RETURN_ON_ACCEPTANCE, opts.GetMode())
		assert.Nil(t, opts.GetWaitTimeout())

		docs, err := testClient.GetDocumentsAlpha1(ctx, store, index, []string{"1", "2", "3"}, false, nil)
		require.NoError(t, err)
		require.Len(t, docs, 1)
		assert.Equal(t, "3", docs[0].ID)
	})
}

func TestIndexingOptions(t *testing.T) {
	assert.Nil(t, toIndexingOptionsProto(nil))

	opts := toIndexingOptionsProto([]IndexingOption{
		WithWaitForIndexCompletion(time.Second, IndexingWaitTimeoutFailRequest),
		nil,
	})
	assert.Equal(t, pb.IndexingMode_INDEXING_MODE_WAIT_FOR_COMPLETION, opts.GetMode())
	assert.Equal(t, time.Second, opts.GetWaitTimeout().AsDuration())
	assert.Equal(t, pb.IndexingWaitTimeoutAction_INDEXING_WAIT_TIMEOUT_ACTION_FAIL_REQUEST, opts.GetOnWaitTimeout())

	// The last option wins and clears wait-only fields.
	opts = toIndexingOptionsProto([]IndexingOption{
		WithWaitForIndexCompletion(time.Second, IndexingWaitTimeoutFailRequest),
		WithReturnOnIndexAcceptance(),
	})
	assert.Equal(t, pb.IndexingMode_INDEXING_MODE_RETURN_ON_ACCEPTANCE, opts.GetMode())
	assert.Nil(t, opts.GetWaitTimeout())
	assert.Equal(t, pb.IndexingWaitTimeoutAction_INDEXING_WAIT_TIMEOUT_ACTION_UNSPECIFIED, opts.GetOnWaitTimeout())

	assert.Equal(t, "INDEX_ACK_COMPLETED", IndexAckCompleted.String())
}
