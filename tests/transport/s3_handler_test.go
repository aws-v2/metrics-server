package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"metrics-gateway/application"
	"metrics-gateway/internal/transport"
)

type mockS3Service struct {
	ingestErr      error
	getByBucketRes []application.S3Metric
	getByBucketErr error
	listRes        []application.S3Metric
	listErr        error
}

func (m *mockS3Service) Ingest(ctx context.Context, req application.S3IngestRequest) error {
	return m.ingestErr
}

func (m *mockS3Service) GetByBucket(ctx context.Context, bucketID string) ([]application.S3Metric, error) {
	return m.getByBucketRes, m.getByBucketErr
}

func (m *mockS3Service) List(ctx context.Context) ([]application.S3Metric, error) {
	return m.listRes, m.listErr
}

func TestS3Handler_Ingest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		payload        any
		ingestErr      error
		expectedStatus int
	}{
		{
			name: "Valid Ingest",
			payload: application.S3IngestRequest{
				BucketID: "test-bucket",
				OwnerID:  "user-1",
			},
			ingestErr:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing BucketID",
			payload: application.S3IngestRequest{
				OwnerID: "user-1",
			},
			ingestErr:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			payload:        "not a json",
			ingestErr:      nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Service{ingestErr: tt.ingestErr}
			handler := transport.NewS3Handler(logger, mock)

			var body []byte
			if s, ok := tt.payload.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.payload)
			}

			req := httptest.NewRequest("POST", "/s3/ingest", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			handler.Ingest(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("%s: handler returned wrong status code: got %v want %v", tt.name, status, tt.expectedStatus)
			}
		})
	}
}

func TestS3Handler_GetByBucket(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mock := &mockS3Service{
		getByBucketRes: []application.S3Metric{{BucketID: "test-bucket"}},
	}
	handler := transport.NewS3Handler(logger, mock)

	req := httptest.NewRequest("GET", "/s3/test-bucket", nil)
	// In Go 1.22+, http.ServeMux handles path parameters. 
	// To test R.PathValue, we need to make sure the request is routed or manually set the value if using httptest.
	req.SetPathValue("bucketId", "test-bucket")
	
	rr := httptest.NewRecorder()

	handler.GetByBucket(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var res []application.S3Metric
	json.NewDecoder(rr.Body).Decode(&res)
	if len(res) != 1 || res[0].BucketID != "test-bucket" {
		t.Errorf("handler returned unexpected body: %v", rr.Body.String())
	}
}
