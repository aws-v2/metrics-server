package transport_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"metrics-gateway/internal/transport"
)

type mockHealthService struct {
	status string
	err    error
}

func (m *mockHealthService) Check(ctx context.Context) (string, error) {
	return m.status, m.err
}

func TestHealthHandler_Health(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		status         string
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Healthy",
			status:         "ok",
			err:            nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "Unhealthy",
			status:         "unhealthy",
			err:            errors.New("db connection failure"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `{"error":"db connection failure","status":"unhealthy"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockHealthService{status: tt.status, err: tt.err}
			handler := transport.NewHealthHandler(logger, mock)

			req := httptest.NewRequest("GET", "/health", nil)
			rr := httptest.NewRecorder()

			handler.Health(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			actualBody := strings.TrimSpace(rr.Body.String())
			if actualBody != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v", actualBody, tt.expectedBody)
			}
		})
	}
}
