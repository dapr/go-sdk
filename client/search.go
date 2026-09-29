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
	"time"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"

	runtimev1pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
)

// IndexAck describes how far a search or vector write progressed before the
// sidecar returned.
type IndexAck int32

const (
	// IndexAckUnspecified is never returned by a successful write.
	IndexAckUnspecified IndexAck = IndexAck(runtimev1pb.IndexAck_INDEX_ACK_UNSPECIFIED)
	// IndexAckQueued means the provider accepted the write for asynchronous
	// processing. It does not indicate that any item was written successfully.
	IndexAckQueued IndexAck = IndexAck(runtimev1pb.IndexAck_INDEX_ACK_QUEUED)
	// IndexAckCompleted means the provider completed the write. FailedItems
	// contains every item-specific failure.
	IndexAckCompleted IndexAck = IndexAck(runtimev1pb.IndexAck_INDEX_ACK_COMPLETED)
)

// String returns the name of the acknowledgement.
func (a IndexAck) String() string {
	return runtimev1pb.IndexAck(a).String()
}

// IndexingWaitTimeoutAction selects what happens when a write that waits for
// completion exceeds its wait timeout.
type IndexingWaitTimeoutAction int32

const (
	// IndexingWaitTimeoutContinueAsync returns IndexAckQueued and lets the
	// durably queued provider task continue.
	IndexingWaitTimeoutContinueAsync IndexingWaitTimeoutAction = IndexingWaitTimeoutAction(runtimev1pb.IndexingWaitTimeoutAction_INDEXING_WAIT_TIMEOUT_ACTION_CONTINUE_ASYNC)
	// IndexingWaitTimeoutFailRequest fails the request with DEADLINE_EXCEEDED.
	// Provider-side work is not guaranteed to be cancelled.
	IndexingWaitTimeoutFailRequest IndexingWaitTimeoutAction = IndexingWaitTimeoutAction(runtimev1pb.IndexingWaitTimeoutAction_INDEXING_WAIT_TIMEOUT_ACTION_FAIL_REQUEST)
)

// IndexingOption configures how a search or vector write is acknowledged.
type IndexingOption func(*runtimev1pb.IndexingOptionsAlpha1)

// WithReturnOnIndexAcceptance returns as soon as the provider durably accepts
// the write for background processing. This is also the behaviour when no
// IndexingOption is supplied.
func WithReturnOnIndexAcceptance() IndexingOption {
	return func(o *runtimev1pb.IndexingOptionsAlpha1) {
		o.Mode = runtimev1pb.IndexingMode_INDEXING_MODE_RETURN_ON_ACCEPTANCE
		o.WaitTimeout = nil
		o.OnWaitTimeout = runtimev1pb.IndexingWaitTimeoutAction_INDEXING_WAIT_TIMEOUT_ACTION_UNSPECIFIED
	}
}

// WithWaitForIndexCompletion waits up to timeout for the provider to finish
// the write. onTimeout selects the outcome when the wait expires. The timeout
// must be positive and shorter than any deadline set on the request context.
func WithWaitForIndexCompletion(timeout time.Duration, onTimeout IndexingWaitTimeoutAction) IndexingOption {
	return func(o *runtimev1pb.IndexingOptionsAlpha1) {
		o.Mode = runtimev1pb.IndexingMode_INDEXING_MODE_WAIT_FOR_COMPLETION
		o.WaitTimeout = durationpb.New(timeout)
		o.OnWaitTimeout = runtimev1pb.IndexingWaitTimeoutAction(onTimeout)
	}
}

func toIndexingOptionsProto(opts []IndexingOption) *runtimev1pb.IndexingOptionsAlpha1 {
	if len(opts) == 0 {
		return nil
	}
	out := &runtimev1pb.IndexingOptionsAlpha1{}
	for _, opt := range opts {
		if opt != nil {
			opt(out)
		}
	}
	return out
}

// FailedItem is an item-specific failure reported by a search or vector write.
type FailedItem struct {
	ID string
	// Err is a gRPC status error; use status.Code(Err) to obtain the code.
	Err error
}

// IndexingResponse is the result of a document indexing or vector upsert
// request.
type IndexingResponse struct {
	Ack IndexAck
	// FailedItems holds item-specific failures known at the acknowledgement
	// boundary.
	FailedItems []FailedItem
}

func fromFailedItemsProto(items []*runtimev1pb.FailedItem) []FailedItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]FailedItem, len(items))
	for i, item := range items {
		out[i] = FailedItem{
			ID:  item.GetId(),
			Err: status.ErrorProto(item.GetError()),
		}
	}
	return out
}

// toStructProto converts a JSON-like map into a protobuf Struct. A nil map
// yields a nil Struct.
func toStructProto(m map[string]any) (*structpb.Struct, error) {
	if m == nil {
		return nil, nil
	}
	return structpb.NewStruct(m)
}

// fromStructProto converts a protobuf Struct into a JSON-like map. A nil
// Struct yields a nil map.
func fromStructProto(s *structpb.Struct) map[string]any {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

// SearchDocument is a single index-ready document.
type SearchDocument struct {
	// ID is the caller-supplied document identifier, unique within an index.
	ID string
	// Content is a UTF-8 encoded JSON object.
	Content []byte
	// Metadata is opaque caller metadata stored with the document. It is not
	// indexed or filterable.
	Metadata map[string]string
}

func (d *SearchDocument) toProto() *runtimev1pb.SearchDocument {
	return &runtimev1pb.SearchDocument{
		Id:       d.ID,
		Content:  d.Content,
		Metadata: d.Metadata,
	}
}

func fromSearchDocumentProto(d *runtimev1pb.SearchDocument) *SearchDocument {
	if d == nil {
		return nil
	}
	return &SearchDocument{
		ID:       d.GetId(),
		Content:  d.GetContent(),
		Metadata: d.GetMetadata(),
	}
}

// SearchIndex describes a search index.
type SearchIndex struct {
	Name string
	// DocumentCount is approximate; providers that cannot supply it return 0.
	DocumentCount uint64
	// Properties are component-specific index properties.
	Properties map[string]string
}

// SortOrder is the direction of a SortClause.
type SortOrder int32

const (
	SortOrderUnspecified SortOrder = SortOrder(runtimev1pb.SortOrder_SORT_ORDER_UNSPECIFIED)
	SortOrderAsc         SortOrder = SortOrder(runtimev1pb.SortOrder_SORT_ORDER_ASC)
	SortOrderDesc        SortOrder = SortOrder(runtimev1pb.SortOrder_SORT_ORDER_DESC)
)

// SortClause orders search hits by a document field.
type SortClause struct {
	Field string
	Order SortOrder
}

// SearchQuery is a query against a search index.
type SearchQuery struct {
	// Text is a lexical query. Mutually exclusive with Native.
	Text string
	// Native is a backend-native query. Mutually exclusive with Text.
	Native map[string]any
	// Filter uses the portable filter DSL ($eq, $ne, $gt, $gte, $lt, $lte,
	// $in, $nin, $exists, $and, $or, $not) over dotted field paths.
	Filter map[string]any
	// TopK is the maximum number of hits to return in this page.
	TopK uint32
	// ContinuationToken is the token returned by the previous page. Empty for
	// the first page.
	ContinuationToken string
	ReturnFields      []string
	IncludeContent    bool
	SearchFields      []string
	Sort              []SortClause
	HighlightFields   []string
	Metadata          map[string]string
}

// TotalHitsRelation describes the accuracy of SearchResponse.TotalHits.
type TotalHitsRelation int32

const (
	TotalHitsRelationUnspecified TotalHitsRelation = TotalHitsRelation(runtimev1pb.TotalHitsRelation_TOTAL_HITS_RELATION_UNSPECIFIED)
	TotalHitsRelationExact       TotalHitsRelation = TotalHitsRelation(runtimev1pb.TotalHitsRelation_TOTAL_HITS_RELATION_EXACT)
	TotalHitsRelationLowerBound  TotalHitsRelation = TotalHitsRelation(runtimev1pb.TotalHitsRelation_TOTAL_HITS_RELATION_LOWER_BOUND)
	TotalHitsRelationEstimate    TotalHitsRelation = TotalHitsRelation(runtimev1pb.TotalHitsRelation_TOTAL_HITS_RELATION_ESTIMATE)
)

// SearchHit is a single search result.
type SearchHit struct {
	Document *SearchDocument
	// Score is an unnormalized provider-specific relevance score; higher is
	// more relevant.
	Score float64
	// Highlights are highlighted fragments keyed by field name.
	Highlights map[string]string
}

// SearchResponse is a page of search results.
type SearchResponse struct {
	Hits []SearchHit
	// TotalHits is a best-effort total, nil when the provider cannot supply one.
	TotalHits         *uint64
	TotalHitsRelation TotalHitsRelation
	// ContinuationToken fetches the next page. Empty when there are no more
	// results.
	ContinuationToken string
}

func validateSearchTarget(storeName, index string) error {
	if storeName == "" {
		return errors.New("storeName is empty")
	}
	if index == "" {
		return errors.New("index is empty")
	}
	return nil
}

// CreateIndexAlpha1 creates a search index.
func (c *GRPCClient) CreateIndexAlpha1(ctx context.Context, storeName, index string, meta map[string]string) error {
	if err := validateSearchTarget(storeName, index); err != nil {
		return err
	}
	_, err := c.protoClient.CreateIndexAlpha1(ctx, &runtimev1pb.CreateIndexRequestAlpha1{
		StoreName: storeName,
		Index:     index,
		Metadata:  meta,
	})
	if err != nil {
		return fmt.Errorf("error creating index %s: %w", index, err)
	}
	return nil
}

// GetIndexAlpha1 returns a search index.
func (c *GRPCClient) GetIndexAlpha1(ctx context.Context, storeName, index string, meta map[string]string) (*SearchIndex, error) {
	if err := validateSearchTarget(storeName, index); err != nil {
		return nil, err
	}
	resp, err := c.protoClient.GetIndexAlpha1(ctx, &runtimev1pb.GetIndexRequestAlpha1{
		StoreName: storeName,
		Index:     index,
		Metadata:  meta,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting index %s: %w", index, err)
	}
	return &SearchIndex{
		Name:          resp.GetIndex(),
		DocumentCount: resp.GetDocumentCount(),
		Properties:    resp.GetProperties(),
	}, nil
}

// ListIndexesAlpha1 lists the indexes in a search store.
func (c *GRPCClient) ListIndexesAlpha1(ctx context.Context, storeName string, meta map[string]string) ([]string, error) {
	if storeName == "" {
		return nil, errors.New("storeName is empty")
	}
	resp, err := c.protoClient.ListIndexesAlpha1(ctx, &runtimev1pb.ListIndexesRequestAlpha1{
		StoreName: storeName,
		Metadata:  meta,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing indexes: %w", err)
	}
	return resp.GetIndexes(), nil
}

// DeleteIndexAlpha1 deletes a search index.
func (c *GRPCClient) DeleteIndexAlpha1(ctx context.Context, storeName, index string, meta map[string]string) error {
	if err := validateSearchTarget(storeName, index); err != nil {
		return err
	}
	_, err := c.protoClient.DeleteIndexAlpha1(ctx, &runtimev1pb.DeleteIndexRequestAlpha1{
		StoreName: storeName,
		Index:     index,
		Metadata:  meta,
	})
	if err != nil {
		return fmt.Errorf("error deleting index %s: %w", index, err)
	}
	return nil
}

// IndexDocumentsAlpha1 upserts documents into a search index, keyed by ID.
func (c *GRPCClient) IndexDocumentsAlpha1(ctx context.Context, storeName, index string, documents []*SearchDocument, meta map[string]string, opts ...IndexingOption) (*IndexingResponse, error) {
	if err := validateSearchTarget(storeName, index); err != nil {
		return nil, err
	}
	docs := make([]*runtimev1pb.SearchDocument, 0, len(documents))
	for _, d := range documents {
		if d == nil {
			return nil, errors.New("document is nil")
		}
		docs = append(docs, d.toProto())
	}
	resp, err := c.protoClient.IndexDocumentsAlpha1(ctx, &runtimev1pb.IndexDocumentsRequestAlpha1{
		StoreName: storeName,
		Index:     index,
		Documents: docs,
		Metadata:  meta,
		Options:   toIndexingOptionsProto(opts),
	})
	if err != nil {
		return nil, fmt.Errorf("error indexing documents: %w", err)
	}
	return &IndexingResponse{
		Ack:         IndexAck(resp.GetAck()),
		FailedItems: fromFailedItemsProto(resp.GetFailedItems()),
	}, nil
}

// GetDocumentsAlpha1 returns the documents with the given IDs in request
// order. IDs that do not exist are omitted.
func (c *GRPCClient) GetDocumentsAlpha1(ctx context.Context, storeName, index string, ids []string, includeContent bool, meta map[string]string) ([]*SearchDocument, error) {
	if err := validateSearchTarget(storeName, index); err != nil {
		return nil, err
	}
	resp, err := c.protoClient.GetDocumentsAlpha1(ctx, &runtimev1pb.GetDocumentsRequestAlpha1{
		StoreName:      storeName,
		Index:          index,
		Ids:            ids,
		IncludeContent: includeContent,
		Metadata:       meta,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting documents: %w", err)
	}
	out := make([]*SearchDocument, len(resp.GetDocuments()))
	for i, d := range resp.GetDocuments() {
		out[i] = fromSearchDocumentProto(d)
	}
	return out, nil
}

// DeleteDocumentsAlpha1 deletes the documents with the given IDs.
func (c *GRPCClient) DeleteDocumentsAlpha1(ctx context.Context, storeName, index string, ids []string, meta map[string]string, opts ...IndexingOption) (IndexAck, error) {
	if err := validateSearchTarget(storeName, index); err != nil {
		return IndexAckUnspecified, err
	}
	resp, err := c.protoClient.DeleteDocumentsAlpha1(ctx, &runtimev1pb.DeleteDocumentsRequestAlpha1{
		StoreName: storeName,
		Index:     index,
		Ids:       ids,
		Metadata:  meta,
		Options:   toIndexingOptionsProto(opts),
	})
	if err != nil {
		return IndexAckUnspecified, fmt.Errorf("error deleting documents: %w", err)
	}
	return IndexAck(resp.GetAck()), nil
}

// SearchAlpha1 searches an index.
func (c *GRPCClient) SearchAlpha1(ctx context.Context, storeName, index string, query *SearchQuery) (*SearchResponse, error) {
	if err := validateSearchTarget(storeName, index); err != nil {
		return nil, err
	}
	if query == nil {
		return nil, errors.New("query is nil")
	}

	req := &runtimev1pb.SearchRequestAlpha1{
		StoreName:         storeName,
		Index:             index,
		TopK:              query.TopK,
		ContinuationToken: query.ContinuationToken,
		ReturnFields:      query.ReturnFields,
		IncludeContent:    query.IncludeContent,
		SearchFields:      query.SearchFields,
		HighlightFields:   query.HighlightFields,
		Metadata:          query.Metadata,
	}

	switch {
	case query.Text != "" && query.Native != nil:
		return nil, errors.New("query Text and Native are mutually exclusive")
	case query.Native != nil:
		native, err := toStructProto(query.Native)
		if err != nil {
			return nil, fmt.Errorf("invalid native query: %w", err)
		}
		req.Query = &runtimev1pb.SearchRequestAlpha1_Native{Native: native}
	default:
		req.Query = &runtimev1pb.SearchRequestAlpha1_Text{Text: query.Text}
	}

	filter, err := toStructProto(query.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}
	req.Filter = filter

	if len(query.Sort) > 0 {
		req.Sort = make([]*runtimev1pb.SortClause, len(query.Sort))
		for i, s := range query.Sort {
			req.Sort[i] = &runtimev1pb.SortClause{
				Field: s.Field,
				Order: runtimev1pb.SortOrder(s.Order),
			}
		}
	}

	resp, err := c.protoClient.SearchAlpha1(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error searching index %s: %w", index, err)
	}

	out := &SearchResponse{
		TotalHits:         resp.TotalHits,
		TotalHitsRelation: TotalHitsRelation(resp.GetTotalHitsRelation()),
		ContinuationToken: resp.GetContinuationToken(),
	}
	if len(resp.GetHits()) > 0 {
		out.Hits = make([]SearchHit, len(resp.GetHits()))
		for i, h := range resp.GetHits() {
			out.Hits[i] = SearchHit{
				Document:   fromSearchDocumentProto(h.GetDocument()),
				Score:      h.GetScore(),
				Highlights: h.GetHighlights(),
			}
		}
	}
	return out, nil
}
