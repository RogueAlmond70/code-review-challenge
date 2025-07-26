package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "notes-service"

// Metrics for the notes-service used to observe the behaviour of the service.

var (
	CountSingleNoteRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "single_note_requests_total",
			Help:      "Counter of requests to notes-service for a single note",
		},
		[]string{"request"},
	)
	CountSingleNoteRequestErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "single_note_request_errors_total",
			Help:      "Counter of errored requests to notes-service for a single note",
		},
		[]string{"request_error"},
	)
	SingleNoteRequestDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "single_note_requests_duration_seconds",
			Help:      "Latency histogram for single note request calls",
			Buckets:   prometheus.DefBuckets,
		},
	)
	CountAllNotesRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "all_notes_requests_total",
			Help:      "Counter of requests to notes-service to return all notes",
		},
		[]string{"request"},
	)
	CountAllNotesRequestErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "all_notes_request_errors_total",
			Help:      "Counter of errored requests to notes-service to return all notes",
		},
		[]string{"request_error"},
	)
	AllNotesRequestDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "all_notes_requests_duration_seconds",
			Help:      "Latency histogram for 'all notes' request calls",
			Buckets:   prometheus.DefBuckets,
		},
	)

	CountArchivedNotesRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_archived_notes_requests_total",
			Help:      "Counter of requests to notes-service for all archived notes",
		},
		[]string{"request"},
	)
	CountArchivedNotesRequestErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_archived_notes_request_errors_total",
			Help:      "Counter of errored requests to notes-service for all archived notes",
		},
		[]string{"request_error"},
	)
	CountArchivedNotesRequestDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "archived_notes_requests_duration_seconds",
			Help:      "Latency histogram for count archived notes calls",
			Buckets:   prometheus.DefBuckets,
		},
	)
	CountUnarchivedNotesRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_unarchived_notes_requests_total",
			Help:      "Counter of requests to notes-service for all unarchived notes",
		},
		[]string{"request"},
	)
	CountUnarchivedNotesRequestErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_unarchived_notes_request_errors_total",
			Help:      "Counter of errored requests to notes-service for all unarchived notes",
		},
		[]string{"request_error"},
	)
	UnarchivedNotesRequestDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "unarchived_notes_requests_duration_seconds",
			Help:      "Latency histogram for count unarchived notes calls",
			Buckets:   prometheus.DefBuckets,
		},
	)
	CountCreateNoteRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_create_note_requests_total",
			Help:      "Counter of requests to notes-service to create a note",
		},
		[]string{"request"},
	)
	CountCreateNoteRequestErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_create_note_request_errors_total",
			Help:      "Counter of errored requests to notes-service to create note",
		},
		[]string{"request_error"},
	)
	CreateNoteRequestDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "create_note_requests_duration_seconds",
			Help:      "Latency histogram for count create note calls",
			Buckets:   prometheus.DefBuckets,
		},
	)
	CountUpdateNoteRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_update_note_requests_total",
			Help:      "Counter of requests to notes-service to update a note",
		},
		[]string{"request"},
	)
	CountUpdateNoteRequestErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_update_note_request_errors_total",
			Help:      "Counter of errored requests to notes-service to update note",
		},
		[]string{"request_error"},
	)
	UpdateNoteRequestDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "update_notes_requests_duration_seconds",
			Help:      "Latency histogram for count update note calls",
			Buckets:   prometheus.DefBuckets,
		},
	)
	CountDeleteNoteRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_delete_note_requests_total",
			Help:      "Counter of requests to notes-service to delete a note",
		},
		[]string{"request"},
	)
	CountDeleteNoteRequestErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "count_delete_note_request_errors_total",
			Help:      "Counter of errored requests to notes-service to delete note",
		},
		[]string{"request_error"},
	)
	DeleteNoteRequestDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "delete_notes_requests_duration_seconds",
			Help:      "Latency histogram for count delete note calls",
			Buckets:   prometheus.DefBuckets,
		},
	)
)
