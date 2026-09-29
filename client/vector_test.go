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
	"context"
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

type fakeCollection struct {
	dimensions uint32
	metric     pb.DistanceMetric
	records    map[string]*pb.VectorRecord
}

// fakeVector is the in-memory vector store backing testDaprServer.
var fakeVector = struct {
	sync.Mutex
	collections map[string]*fakeCollection
	lastOptions *pb.IndexingOptionsAlpha1
	lastQuery   *pb.QueryVectorsRequestAlpha1
	lastBatch   *pb.BatchQueryVectorsRequestAlpha1
}{
	collections: make(map[string]*fakeCollection),
}

func (s *testDaprServer) CreateCollectionAlpha1(_ context.Context, in *pb.CreateCollectionRequestAlpha1) (*emptypb.Empty, error) {
	fakeVector.Lock()
	defer fakeVector.Unlock()
	key := fakeKey(in.GetStoreName(), in.GetCollection())
	if _, ok := fakeVector.collections[key]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "collection %s already exists", in.GetCollection())
	}
	metric := in.GetMetric()
	if metric == pb.DistanceMetric_DISTANCE_METRIC_UNSPECIFIED {
		metric = pb.DistanceMetric_DISTANCE_METRIC_COSINE
	}
	fakeVector.collections[key] = &fakeCollection{
		dimensions: in.GetDimensions(),
		metric:     metric,
		records:    make(map[string]*pb.VectorRecord),
	}
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) GetCollectionAlpha1(_ context.Context, in *pb.GetCollectionRequestAlpha1) (*pb.GetCollectionResponseAlpha1, error) {
	fakeVector.Lock()
	defer fakeVector.Unlock()
	c, ok := fakeVector.collections[fakeKey(in.GetStoreName(), in.GetCollection())]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "collection %s not found", in.GetCollection())
	}
	return &pb.GetCollectionResponseAlpha1{
		Collection:  in.GetCollection(),
		RecordCount: uint64(len(c.records)),
		Dimensions:  c.dimensions,
		Metric:      c.metric,
		Properties:  in.GetMetadata(),
	}, nil
}

func (s *testDaprServer) ListCollectionsAlpha1(_ context.Context, in *pb.ListCollectionsRequestAlpha1) (*pb.ListCollectionsResponseAlpha1, error) {
	fakeVector.Lock()
	defer fakeVector.Unlock()
	var names []string
	for key := range fakeVector.collections {
		if store, name, _ := strings.Cut(key, "||"); store == in.GetStoreName() {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return &pb.ListCollectionsResponseAlpha1{Collections: names}, nil
}

func (s *testDaprServer) DeleteCollectionAlpha1(_ context.Context, in *pb.DeleteCollectionRequestAlpha1) (*emptypb.Empty, error) {
	fakeVector.Lock()
	defer fakeVector.Unlock()
	delete(fakeVector.collections, fakeKey(in.GetStoreName(), in.GetCollection()))
	return &emptypb.Empty{}, nil
}

func (s *testDaprServer) UpsertVectorsAlpha1(_ context.Context, in *pb.UpsertVectorsRequestAlpha1) (*pb.UpsertVectorsResponseAlpha1, error) {
	fakeVector.Lock()
	defer fakeVector.Unlock()
	fakeVector.lastOptions = in.GetOptions()
	c, ok := fakeVector.collections[fakeKey(in.GetStoreName(), in.GetCollection())]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "collection %s not found", in.GetCollection())
	}
	resp := &pb.UpsertVectorsResponseAlpha1{Ack: ackFor(in.GetOptions())}
	for _, r := range in.GetRecords() {
		if uint32(len(r.GetValues())) != c.dimensions {
			resp.FailedItems = append(resp.FailedItems, &pb.FailedItem{
				Id:    r.GetId(),
				Error: status.New(codes.InvalidArgument, "dimension mismatch").Proto(),
			})
			continue
		}
		c.records[r.GetId()] = proto.Clone(r).(*pb.VectorRecord)
	}
	return resp, nil
}

func (s *testDaprServer) DeleteVectorsAlpha1(_ context.Context, in *pb.DeleteVectorsRequestAlpha1) (*pb.DeleteVectorsResponseAlpha1, error) {
	fakeVector.Lock()
	defer fakeVector.Unlock()
	fakeVector.lastOptions = in.GetOptions()
	if c, ok := fakeVector.collections[fakeKey(in.GetStoreName(), in.GetCollection())]; ok {
		for _, id := range in.GetIds() {
			delete(c.records, id)
		}
	}
	return &pb.DeleteVectorsResponseAlpha1{Ack: ackFor(in.GetOptions())}, nil
}

func (s *testDaprServer) GetVectorsAlpha1(_ context.Context, in *pb.GetVectorsRequestAlpha1) (*pb.GetVectorsResponseAlpha1, error) {
	fakeVector.Lock()
	defer fakeVector.Unlock()
	resp := &pb.GetVectorsResponseAlpha1{}
	c, ok := fakeVector.collections[fakeKey(in.GetStoreName(), in.GetCollection())]
	if !ok {
		return resp, nil
	}
	for _, id := range in.GetIds() {
		r, ok := c.records[id]
		if !ok {
			continue
		}
		r = proto.Clone(r).(*pb.VectorRecord)
		if !in.GetIncludeValues() {
			r.Values = nil
		}
		resp.Records = append(resp.Records, r)
	}
	return resp, nil
}

// fakeQuery returns every record in the collection, ordered by ID, with a
// score of 1. A by_id query for an unknown record fails with NotFound.
func fakeQuery(c *fakeCollection, in *pb.QueryVectorsRequestAlpha1) (*pb.QueryVectorsResponseAlpha1, error) {
	if id := in.GetById(); id != "" {
		if _, ok := c.records[id]; !ok {
			return nil, status.Errorf(codes.NotFound, "record %s not found", id)
		}
	}
	ids := make([]string, 0, len(c.records))
	for id := range c.records {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	if top := int(in.GetTopK()); top > 0 && len(ids) > top {
		ids = ids[:top]
	}
	metric := in.GetMetric()
	if metric == pb.DistanceMetric_DISTANCE_METRIC_UNSPECIFIED {
		metric = c.metric
	}
	resp := &pb.QueryVectorsResponseAlpha1{Metric: metric}
	for _, id := range ids {
		r := proto.Clone(c.records[id]).(*pb.VectorRecord)
		if !in.GetIncludeValues() {
			r.Values = nil
		}
		if !in.GetIncludePayload() {
			r.Payload = nil
		}
		resp.Matches = append(resp.Matches, &pb.VectorMatch{Record: r, Score: 1})
	}
	return resp, nil
}

func (s *testDaprServer) QueryVectorsAlpha1(_ context.Context, in *pb.QueryVectorsRequestAlpha1) (*pb.QueryVectorsResponseAlpha1, error) {
	fakeVector.Lock()
	defer fakeVector.Unlock()
	fakeVector.lastQuery = proto.Clone(in).(*pb.QueryVectorsRequestAlpha1)
	c, ok := fakeVector.collections[fakeKey(in.GetStoreName(), in.GetCollection())]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "collection %s not found", in.GetCollection())
	}
	return fakeQuery(c, in)
}

func (s *testDaprServer) BatchQueryVectorsAlpha1(_ context.Context, in *pb.BatchQueryVectorsRequestAlpha1) (*pb.BatchQueryVectorsResponseAlpha1, error) {
	fakeVector.Lock()
	defer fakeVector.Unlock()
	fakeVector.lastBatch = proto.Clone(in).(*pb.BatchQueryVectorsRequestAlpha1)
	c, ok := fakeVector.collections[fakeKey(in.GetStoreName(), in.GetCollection())]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "collection %s not found", in.GetCollection())
	}
	resp := &pb.BatchQueryVectorsResponseAlpha1{}
	for _, q := range in.GetQueries() {
		r, err := fakeQuery(c, q)
		if err != nil {
			resp.Results = append(resp.Results, &pb.BatchQueryResultAlpha1{
				Result: &pb.BatchQueryResultAlpha1_Error{Error: status.Convert(err).Proto()},
			})
			continue
		}
		resp.Results = append(resp.Results, &pb.BatchQueryResultAlpha1{
			Result: &pb.BatchQueryResultAlpha1_Response{Response: r},
		})
	}
	return resp, nil
}

func TestVectorAlpha1(t *testing.T) {
	ctx := t.Context()
	const store = "vector-store"
	const collection = "manuals"

	// TestMain runs the suite once per client transport; start from a clean store.
	fakeVector.Lock()
	fakeVector.collections = make(map[string]*fakeCollection)
	fakeVector.Unlock()

	t.Run("validation", func(t *testing.T) {
		require.EqualError(t, testClient.CreateCollectionAlpha1(ctx, "", collection, 4, DistanceMetricCosine, nil), "storeName is empty")
		require.EqualError(t, testClient.CreateCollectionAlpha1(ctx, store, "", 4, DistanceMetricCosine, nil), "collection is empty")
		require.EqualError(t, testClient.CreateCollectionAlpha1(ctx, store, collection, 0, DistanceMetricCosine, nil), "dimensions must be greater than zero")
		_, err := testClient.ListCollectionsAlpha1(ctx, "", nil)
		require.EqualError(t, err, "storeName is empty")
		_, err = testClient.QueryVectorsAlpha1(ctx, store, collection, nil)
		require.EqualError(t, err, "query is nil")
		_, err = testClient.QueryVectorsAlpha1(ctx, store, collection, &VectorQuery{})
		require.EqualError(t, err, "query requires either Vector or ByID")
		_, err = testClient.QueryVectorsAlpha1(ctx, store, collection, &VectorQuery{Vector: []float32{1}, ByID: "a"})
		require.EqualError(t, err, "query Vector and ByID are mutually exclusive")
		_, err = testClient.QueryVectorsAlpha1(ctx, store, collection, &VectorQuery{ByID: "a", Filter: map[string]any{"bad": make(chan int)}})
		require.ErrorContains(t, err, "invalid filter")
		_, err = testClient.BatchQueryVectorsAlpha1(ctx, store, collection, []*VectorQuery{{ByID: "a"}, nil}, nil)
		require.EqualError(t, err, "query 1 is nil")
		_, err = testClient.UpsertVectorsAlpha1(ctx, store, collection, []*VectorRecord{nil}, nil)
		require.EqualError(t, err, "vector record is nil")
		_, err = testClient.UpsertVectorsAlpha1(ctx, store, collection, []*VectorRecord{{ID: "x", Metadata: map[string]any{"bad": make(chan int)}}}, nil)
		require.ErrorContains(t, err, `invalid metadata for vector record "x"`)
	})

	t.Run("collection lifecycle", func(t *testing.T) {
		require.NoError(t, testClient.CreateCollectionAlpha1(ctx, store, collection, 4, DistanceMetricUnspecified, nil))
		require.NoError(t, testClient.CreateCollectionAlpha1(ctx, store, "other", 2, DistanceMetricEuclidean, nil))

		err := testClient.CreateCollectionAlpha1(ctx, store, collection, 4, DistanceMetricCosine, nil)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))

		cols, err := testClient.ListCollectionsAlpha1(ctx, store, nil)
		require.NoError(t, err)
		assert.Equal(t, []string{collection, "other"}, cols)

		col, err := testClient.GetCollectionAlpha1(ctx, store, "other", map[string]string{"p": "v"})
		require.NoError(t, err)
		assert.Equal(t, &VectorCollection{
			Name:       "other",
			Dimensions: 2,
			Metric:     DistanceMetricEuclidean,
			Properties: map[string]string{"p": "v"},
		}, col)

		require.NoError(t, testClient.DeleteCollectionAlpha1(ctx, store, "other", nil))
		_, err = testClient.GetCollectionAlpha1(ctx, store, "other", nil)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("upsert vectors", func(t *testing.T) {
		resp, err := testClient.UpsertVectorsAlpha1(ctx, store, collection, []*VectorRecord{
			{ID: "a", Values: []float32{1, 0, 0, 0}, Payload: []byte("pa"), Metadata: map[string]any{"tenant": "acme", "revision": 3}},
			{ID: "b", Values: []float32{0, 1, 0, 0}, Payload: []byte("pb")},
			{ID: "short", Values: []float32{1}},
		}, nil, WithWaitForIndexCompletion(time.Second, IndexingWaitTimeoutFailRequest))
		require.NoError(t, err)
		assert.Equal(t, IndexAckCompleted, resp.Ack)
		require.Len(t, resp.FailedItems, 1)
		assert.Equal(t, "short", resp.FailedItems[0].ID)
		assert.Equal(t, codes.InvalidArgument, status.Code(resp.FailedItems[0].Err))

		fakeVector.Lock()
		opts := fakeVector.lastOptions
		fakeVector.Unlock()
		assert.Equal(t, pb.IndexingWaitTimeoutAction_INDEXING_WAIT_TIMEOUT_ACTION_FAIL_REQUEST, opts.GetOnWaitTimeout())

		col, err := testClient.GetCollectionAlpha1(ctx, store, collection, nil)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), col.RecordCount)
		assert.Equal(t, uint32(4), col.Dimensions)
		assert.Equal(t, DistanceMetricCosine, col.Metric)
	})

	t.Run("get vectors", func(t *testing.T) {
		recs, err := testClient.GetVectorsAlpha1(ctx, store, collection, []string{"b", "missing", "a"}, true, nil)
		require.NoError(t, err)
		require.Len(t, recs, 2)
		assert.Equal(t, "b", recs[0].ID)
		assert.Nil(t, recs[0].Metadata)
		assert.Equal(t, &VectorRecord{
			ID:       "a",
			Values:   []float32{1, 0, 0, 0},
			Payload:  []byte("pa"),
			Metadata: map[string]any{"tenant": "acme", "revision": float64(3)},
		}, recs[1])

		recs, err = testClient.GetVectorsAlpha1(ctx, store, collection, []string{"a"}, false, nil)
		require.NoError(t, err)
		require.Len(t, recs, 1)
		assert.Empty(t, recs[0].Values)
	})

	t.Run("query vectors", func(t *testing.T) {
		threshold := 0.5
		resp, err := testClient.QueryVectorsAlpha1(ctx, store, collection, &VectorQuery{
			Vector:         []float32{1, 0, 0, 0},
			TopK:           1,
			Filter:         map[string]any{"tenant": "acme"},
			IncludeValues:  true,
			IncludePayload: true,
			ScoreThreshold: &threshold,
			Metadata:       map[string]string{"k": "v"},
		})
		require.NoError(t, err)
		assert.Equal(t, DistanceMetricCosine, resp.Metric)
		require.Len(t, resp.Matches, 1)
		assert.Equal(t, "a", resp.Matches[0].Record.ID)
		assert.Equal(t, []byte("pa"), resp.Matches[0].Record.Payload)
		assert.Equal(t, []float32{1, 0, 0, 0}, resp.Matches[0].Record.Values)
		assert.InDelta(t, 1.0, resp.Matches[0].Score, 0)

		fakeVector.Lock()
		req := fakeVector.lastQuery
		fakeVector.Unlock()
		assert.Equal(t, []float32{1, 0, 0, 0}, req.GetVector().GetValues())
		assert.Equal(t, map[string]any{"tenant": "acme"}, req.GetFilter().AsMap())
		require.NotNil(t, req.ScoreThreshold)
		assert.InDelta(t, 0.5, req.GetScoreThreshold(), 0)
		assert.Equal(t, map[string]string{"k": "v"}, req.GetMetadata())

		resp, err = testClient.QueryVectorsAlpha1(ctx, store, collection, &VectorQuery{
			ByID:   "b",
			Metric: DistanceMetricDotProduct,
		})
		require.NoError(t, err)
		assert.Equal(t, DistanceMetricDotProduct, resp.Metric)
		require.Len(t, resp.Matches, 2)
		assert.Nil(t, resp.Matches[0].Record.Payload)
		assert.Nil(t, resp.Matches[0].Record.Values)

		fakeVector.Lock()
		req = fakeVector.lastQuery
		fakeVector.Unlock()
		assert.Equal(t, "b", req.GetById())
		assert.Nil(t, req.ScoreThreshold)
	})

	t.Run("batch query vectors", func(t *testing.T) {
		results, err := testClient.BatchQueryVectorsAlpha1(ctx, store, collection, []*VectorQuery{
			{Vector: []float32{0, 1, 0, 0}, TopK: 1},
			{ByID: "missing"},
		}, map[string]string{"k": "v"})
		require.NoError(t, err)
		require.Len(t, results, 2)

		require.NoError(t, results[0].Err)
		require.Len(t, results[0].Response.Matches, 1)
		assert.Equal(t, "a", results[0].Response.Matches[0].Record.ID)

		assert.Nil(t, results[1].Response)
		assert.Equal(t, codes.NotFound, status.Code(results[1].Err))

		fakeVector.Lock()
		req := fakeVector.lastBatch
		fakeVector.Unlock()
		assert.Equal(t, map[string]string{"k": "v"}, req.GetMetadata())
		require.Len(t, req.GetQueries(), 2)
		assert.Equal(t, "missing", req.GetQueries()[1].GetById())
	})

	t.Run("delete vectors", func(t *testing.T) {
		ack, err := testClient.DeleteVectorsAlpha1(ctx, store, collection, []string{"a"}, nil)
		require.NoError(t, err)
		assert.Equal(t, IndexAckQueued, ack)

		recs, err := testClient.GetVectorsAlpha1(ctx, store, collection, []string{"a", "b"}, false, nil)
		require.NoError(t, err)
		require.Len(t, recs, 1)
		assert.Equal(t, "b", recs[0].ID)
	})

	assert.Equal(t, "DISTANCE_METRIC_EUCLIDEAN", DistanceMetricEuclidean.String())
}
