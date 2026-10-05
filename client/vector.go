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
	"errors"
	"fmt"

	"google.golang.org/grpc/status"

	runtimev1pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

// DistanceMetric is the similarity metric of a vector collection or query.
type DistanceMetric int32

const (
	// DistanceMetricUnspecified selects the component's default metric on
	// collection creation, and the collection's metric on queries.
	DistanceMetricUnspecified DistanceMetric = DistanceMetric(runtimev1pb.DistanceMetric_DISTANCE_METRIC_UNSPECIFIED)
	// DistanceMetricCosine is cosine similarity; higher scores are closer.
	DistanceMetricCosine DistanceMetric = DistanceMetric(runtimev1pb.DistanceMetric_DISTANCE_METRIC_COSINE)
	// DistanceMetricDotProduct is dot-product similarity; higher scores are
	// closer.
	DistanceMetricDotProduct DistanceMetric = DistanceMetric(runtimev1pb.DistanceMetric_DISTANCE_METRIC_DOT_PRODUCT)
	// DistanceMetricEuclidean is Euclidean distance; lower scores are closer.
	DistanceMetricEuclidean DistanceMetric = DistanceMetric(runtimev1pb.DistanceMetric_DISTANCE_METRIC_EUCLIDEAN)
)

// String returns the name of the metric.
func (m DistanceMetric) String() string {
	return runtimev1pb.DistanceMetric(m).String()
}

// VectorCollection describes a vector collection.
type VectorCollection struct {
	Name string
	// RecordCount is approximate; providers that cannot supply it return 0.
	RecordCount uint64
	Dimensions  uint32
	// Metric is the effective metric of the collection.
	Metric DistanceMetric
	// Properties are component-specific collection properties.
	Properties map[string]string
}

// VectorRecord is a dense vector with its associated data.
type VectorRecord struct {
	ID string
	// Values is the dense vector. Its length must equal the collection's
	// dimensions.
	Values []float32
	// Payload is opaque caller data stored with the record. It is not
	// filterable.
	Payload []byte
	// Metadata holds structured attributes addressed by the filter DSL.
	Metadata map[string]any
}

func (r *VectorRecord) toProto() (*runtimev1pb.VectorRecord, error) {
	metadata, err := toStructProto(r.Metadata)
	if err != nil {
		return nil, fmt.Errorf("invalid metadata for vector record %q: %w", r.ID, err)
	}
	return &runtimev1pb.VectorRecord{
		Id:       r.ID,
		Values:   r.Values,
		Payload:  r.Payload,
		Metadata: metadata,
	}, nil
}

func fromVectorRecordProto(r *runtimev1pb.VectorRecord) *VectorRecord {
	if r == nil {
		return nil
	}
	return &VectorRecord{
		ID:       r.GetId(),
		Values:   r.GetValues(),
		Payload:  r.GetPayload(),
		Metadata: fromStructProto(r.GetMetadata()),
	}
}

// VectorQuery is a similarity query against a vector collection. Exactly one
// of Vector or ByID must be set.
type VectorQuery struct {
	// Vector is the dense query vector.
	Vector []float32
	// ByID queries using the stored vector of an existing record.
	ByID string
	TopK uint32
	// Filter uses the portable filter DSL over VectorRecord.Metadata.
	Filter         map[string]any
	IncludeValues  bool
	IncludePayload bool
	// Metric determines how scores and ScoreThreshold are interpreted.
	// DistanceMetricUnspecified uses the collection's metric.
	Metric DistanceMetric
	// ScoreThreshold is an inclusive cutoff: matches must score at least this
	// value for cosine and dot-product, or at most this value for Euclidean.
	ScoreThreshold *float64
	Metadata       map[string]string
}

func (q *VectorQuery) toProto(storeName, collection string) (*runtimev1pb.QueryVectorsRequestAlpha1, error) {
	req := &runtimev1pb.QueryVectorsRequestAlpha1{
		StoreName:      storeName,
		Collection:     collection,
		TopK:           q.TopK,
		IncludeValues:  q.IncludeValues,
		IncludePayload: q.IncludePayload,
		Metric:         runtimev1pb.DistanceMetric(q.Metric),
		ScoreThreshold: q.ScoreThreshold,
		Metadata:       q.Metadata,
	}

	switch {
	case q.Vector != nil && q.ByID != "":
		return nil, errors.New("query Vector and ByID are mutually exclusive")
	case q.Vector != nil:
		req.Query = &runtimev1pb.QueryVectorsRequestAlpha1_Vector{
			Vector: &runtimev1pb.VectorRecord{Values: q.Vector},
		}
	case q.ByID != "":
		req.Query = &runtimev1pb.QueryVectorsRequestAlpha1_ById{ById: q.ByID}
	default:
		return nil, errors.New("query requires either Vector or ByID")
	}

	filter, err := toStructProto(q.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}
	req.Filter = filter

	return req, nil
}

// VectorMatch is a single vector query result.
type VectorMatch struct {
	Record *VectorRecord
	// Score is the unnormalized value of the effective metric.
	Score float64
}

// VectorQueryResponse is the result of a vector query.
type VectorQueryResponse struct {
	Matches []VectorMatch
	// Metric is the effective metric used for scores.
	Metric DistanceMetric
}

func fromQueryVectorsResponseProto(resp *runtimev1pb.QueryVectorsResponseAlpha1) *VectorQueryResponse {
	out := &VectorQueryResponse{
		Metric: DistanceMetric(resp.GetMetric()),
	}
	if len(resp.GetMatches()) > 0 {
		out.Matches = make([]VectorMatch, len(resp.GetMatches()))
		for i, m := range resp.GetMatches() {
			out.Matches[i] = VectorMatch{
				Record: fromVectorRecordProto(m.GetRecord()),
				Score:  m.GetScore(),
			}
		}
	}
	return out
}

// BatchQueryResult is the outcome of one query in a batch. Exactly one of
// Response or Err is set.
type BatchQueryResult struct {
	Response *VectorQueryResponse
	// Err is a gRPC status error; use status.Code(Err) to obtain the code.
	Err error
}

func validateVectorTarget(storeName, collection string) error {
	if storeName == "" {
		return errors.New("storeName is empty")
	}
	if collection == "" {
		return errors.New("collection is empty")
	}
	return nil
}

// CreateCollectionAlpha1 creates a vector collection holding vectors of the
// given dimensions. DistanceMetricUnspecified selects the component's default
// metric.
func (c *GRPCClient) CreateCollectionAlpha1(ctx context.Context, storeName, collection string, dimensions uint32, metric DistanceMetric, meta map[string]string) error {
	if err := validateVectorTarget(storeName, collection); err != nil {
		return err
	}
	if dimensions == 0 {
		return errors.New("dimensions must be greater than zero")
	}
	_, err := c.protoClient.CreateCollectionAlpha1(ctx, &runtimev1pb.CreateCollectionRequestAlpha1{
		StoreName:  storeName,
		Collection: collection,
		Dimensions: dimensions,
		Metric:     runtimev1pb.DistanceMetric(metric),
		Metadata:   meta,
	})
	if err != nil {
		return fmt.Errorf("error creating collection %s: %w", collection, err)
	}
	return nil
}

// GetCollectionAlpha1 returns a vector collection.
func (c *GRPCClient) GetCollectionAlpha1(ctx context.Context, storeName, collection string, meta map[string]string) (*VectorCollection, error) {
	if err := validateVectorTarget(storeName, collection); err != nil {
		return nil, err
	}
	resp, err := c.protoClient.GetCollectionAlpha1(ctx, &runtimev1pb.GetCollectionRequestAlpha1{
		StoreName:  storeName,
		Collection: collection,
		Metadata:   meta,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting collection %s: %w", collection, err)
	}
	return &VectorCollection{
		Name:        resp.GetCollection(),
		RecordCount: resp.GetRecordCount(),
		Dimensions:  resp.GetDimensions(),
		Metric:      DistanceMetric(resp.GetMetric()),
		Properties:  resp.GetProperties(),
	}, nil
}

// ListCollectionsAlpha1 lists the collections in a vector store.
func (c *GRPCClient) ListCollectionsAlpha1(ctx context.Context, storeName string, meta map[string]string) ([]string, error) {
	if storeName == "" {
		return nil, errors.New("storeName is empty")
	}
	resp, err := c.protoClient.ListCollectionsAlpha1(ctx, &runtimev1pb.ListCollectionsRequestAlpha1{
		StoreName: storeName,
		Metadata:  meta,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing collections: %w", err)
	}
	return resp.GetCollections(), nil
}

// DeleteCollectionAlpha1 deletes a vector collection.
func (c *GRPCClient) DeleteCollectionAlpha1(ctx context.Context, storeName, collection string, meta map[string]string) error {
	if err := validateVectorTarget(storeName, collection); err != nil {
		return err
	}
	_, err := c.protoClient.DeleteCollectionAlpha1(ctx, &runtimev1pb.DeleteCollectionRequestAlpha1{
		StoreName:  storeName,
		Collection: collection,
		Metadata:   meta,
	})
	if err != nil {
		return fmt.Errorf("error deleting collection %s: %w", collection, err)
	}
	return nil
}

// UpsertVectorsAlpha1 upserts records into a vector collection, keyed by ID.
func (c *GRPCClient) UpsertVectorsAlpha1(ctx context.Context, storeName, collection string, records []*VectorRecord, meta map[string]string, opts ...IndexingOption) (*IndexingResponse, error) {
	if err := validateVectorTarget(storeName, collection); err != nil {
		return nil, err
	}
	recs := make([]*runtimev1pb.VectorRecord, 0, len(records))
	for _, r := range records {
		if r == nil {
			return nil, errors.New("vector record is nil")
		}
		rec, err := r.toProto()
		if err != nil {
			return nil, err
		}
		recs = append(recs, rec)
	}
	resp, err := c.protoClient.UpsertVectorsAlpha1(ctx, &runtimev1pb.UpsertVectorsRequestAlpha1{
		StoreName:  storeName,
		Collection: collection,
		Records:    recs,
		Metadata:   meta,
		Options:    toIndexingOptionsProto(opts),
	})
	if err != nil {
		return nil, fmt.Errorf("error upserting vectors: %w", err)
	}
	return &IndexingResponse{
		Ack:         IndexAck(resp.GetAck()),
		FailedItems: fromFailedItemsProto(resp.GetFailedItems()),
	}, nil
}

// DeleteVectorsAlpha1 deletes the records with the given IDs.
func (c *GRPCClient) DeleteVectorsAlpha1(ctx context.Context, storeName, collection string, ids []string, meta map[string]string, opts ...IndexingOption) (IndexAck, error) {
	if err := validateVectorTarget(storeName, collection); err != nil {
		return IndexAckUnspecified, err
	}
	resp, err := c.protoClient.DeleteVectorsAlpha1(ctx, &runtimev1pb.DeleteVectorsRequestAlpha1{
		StoreName:  storeName,
		Collection: collection,
		Ids:        ids,
		Metadata:   meta,
		Options:    toIndexingOptionsProto(opts),
	})
	if err != nil {
		return IndexAckUnspecified, fmt.Errorf("error deleting vectors: %w", err)
	}
	return IndexAck(resp.GetAck()), nil
}

// GetVectorsAlpha1 returns the records with the given IDs in request order.
// IDs that do not exist are omitted.
func (c *GRPCClient) GetVectorsAlpha1(ctx context.Context, storeName, collection string, ids []string, includeValues bool, meta map[string]string) ([]*VectorRecord, error) {
	if err := validateVectorTarget(storeName, collection); err != nil {
		return nil, err
	}
	resp, err := c.protoClient.GetVectorsAlpha1(ctx, &runtimev1pb.GetVectorsRequestAlpha1{
		StoreName:     storeName,
		Collection:    collection,
		Ids:           ids,
		IncludeValues: includeValues,
		Metadata:      meta,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting vectors: %w", err)
	}
	out := make([]*VectorRecord, len(resp.GetRecords()))
	for i, r := range resp.GetRecords() {
		out[i] = fromVectorRecordProto(r)
	}
	return out, nil
}

// QueryVectorsAlpha1 runs a similarity query against a vector collection.
func (c *GRPCClient) QueryVectorsAlpha1(ctx context.Context, storeName, collection string, query *VectorQuery) (*VectorQueryResponse, error) {
	if err := validateVectorTarget(storeName, collection); err != nil {
		return nil, err
	}
	if query == nil {
		return nil, errors.New("query is nil")
	}
	req, err := query.toProto(storeName, collection)
	if err != nil {
		return nil, err
	}
	resp, err := c.protoClient.QueryVectorsAlpha1(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error querying collection %s: %w", collection, err)
	}
	return fromQueryVectorsResponseProto(resp), nil
}

// BatchQueryVectorsAlpha1 runs several queries against the same collection.
// Each query succeeds or fails on its own; results are returned in query
// order.
func (c *GRPCClient) BatchQueryVectorsAlpha1(ctx context.Context, storeName, collection string, queries []*VectorQuery, meta map[string]string) ([]BatchQueryResult, error) {
	if err := validateVectorTarget(storeName, collection); err != nil {
		return nil, err
	}
	reqs := make([]*runtimev1pb.QueryVectorsRequestAlpha1, len(queries))
	for i, q := range queries {
		if q == nil {
			return nil, fmt.Errorf("query %d is nil", i)
		}
		req, err := q.toProto(storeName, collection)
		if err != nil {
			return nil, fmt.Errorf("invalid query %d: %w", i, err)
		}
		reqs[i] = req
	}
	resp, err := c.protoClient.BatchQueryVectorsAlpha1(ctx, &runtimev1pb.BatchQueryVectorsRequestAlpha1{
		StoreName:  storeName,
		Collection: collection,
		Queries:    reqs,
		Metadata:   meta,
	})
	if err != nil {
		return nil, fmt.Errorf("error batch querying collection %s: %w", collection, err)
	}
	out := make([]BatchQueryResult, len(resp.GetResults()))
	for i, r := range resp.GetResults() {
		switch res := r.GetResult().(type) {
		case *runtimev1pb.BatchQueryResultAlpha1_Response:
			out[i] = BatchQueryResult{Response: fromQueryVectorsResponseProto(res.Response)}
		case *runtimev1pb.BatchQueryResultAlpha1_Error:
			out[i] = BatchQueryResult{Err: status.ErrorProto(res.Error)}
		default:
			out[i] = BatchQueryResult{Err: fmt.Errorf("query %d returned no result", i)}
		}
	}
	return out, nil
}
