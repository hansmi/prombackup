package api

import (
	"context"
	"io"
	"time"
)

// SnapshotOptions are the options available for requesting a new snapshot.
type SnapshotOptions struct {
	// Whether to skip data present in the head block.
	SkipHead bool `json:"skip_head"`
}

// SnapshotResult reports the name of a newly created TSDB snapshot.
type SnapshotResult struct {
	// Snapshot name.
	Name string `json:"name"`
}

// DownloadOptions are the options available when requesting the download of
// a snapshot archive.
type DownloadOptions struct {
	// Snapshot name.
	SnapshotName string

	// Requested archive format.
	Format ArchiveFormat

	// Function returning a writer for storing the body returned by the server.
	BodyWriter func(DownloadResult) (io.Writer, error)
}

// DownloadResult contains information about a requested snapshot download.
type DownloadResult struct {
	// Unique download ID, used to request the status.
	ID string `json:"id"`

	// Body content type (e.g. "application/x-tar").
	ContentType string

	// Parameters for body content type (e.g. "charset").
	ContentTypeParams map[string]string

	// The preferred filename as reported by the server.
	Filename string
}

// DownloadStatusOptions are the options available when requesting status
// information about an ongoing or recent download.
type DownloadStatusOptions struct {
	// Unique download ID.
	ID string `json:"id"`
}

// DownloadStatus contains information about an ongoing or recent download.
type DownloadStatus struct {
	// Unique download ID.
	ID string `json:"id"`

	// Requested snapshot name.
	SnapshotName string `json:"snapshot_name"`

	// Finished is non-nil if the server consider the download finished.
	Finished *DownloadStatusFinished `json:"finished"`
}

// DownloadStatusFinished contains information about a finished download.
type DownloadStatusFinished struct {
	// Success is true if the server encountered no error while generating the
	// snapshot archive.
	Success bool `json:"success"`

	// ErrorText contains a descriptive error message after the server
	// encountered an error.
	ErrorText *string `json:"error_text"`

	// Sha256Hex is the result of the SHA256 algorithm over the downloaded
	// archive.
	Sha256Hex string `json:"sha256_hex"`
}

// PruneOptions are the options available when requesting pruning of snapshots.
type PruneOptions struct {
	// Keep all snapshots within this time interval.
	KeepWithin time.Duration `json:"keep_within"`
}

// PruneResult may be used in the future.
type PruneResult struct {
}

type Interface interface {
	Snapshot(context.Context, SnapshotOptions) (*SnapshotResult, error)
	Download(context.Context, DownloadOptions) (*DownloadResult, error)
	DownloadStatus(context.Context, DownloadStatusOptions) (*DownloadStatus, error)
	Prune(context.Context, PruneOptions) (*PruneResult, error)
}
