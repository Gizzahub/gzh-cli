package sync

import (
	"context"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// VersionSyncStatus represents the synchronization status between a version manager and package manager
type VersionSyncStatus struct {
	VersionManager    string   `json:"version_manager"`
	PackageManager    string   `json:"package_manager"`
	VMVersion         string   `json:"vm_version"`
	PMVersion         string   `json:"pm_version"`
	ExpectedPMVersion string   `json:"expected_pm_version"`
	InSync            bool     `json:"in_sync"`
	SyncAction        string   `json:"sync_action"`
	Issues            []string `json:"issues,omitempty"`
}

// SyncReport provides a comprehensive view of all version manager synchronization
type SyncReport struct {
	Platform       string              `json:"platform"`
	TotalPairs     int                 `json:"total_pairs"`
	InSyncCount    int                 `json:"in_sync_count"`
	OutOfSyncCount int                 `json:"out_of_sync_count"`
	SyncStatuses   []VersionSyncStatus `json:"sync_statuses"`
	Timestamp      time.Time           `json:"timestamp"`
}

// SyncPolicy defines how synchronization should be performed
type SyncPolicy struct {
	Strategy      string `json:"strategy"` // "vm_priority", "pm_priority", "latest"
	AutoFix       bool   `json:"auto_fix"`
	BackupEnabled bool   `json:"backup_enabled"`
	PromptUser    bool   `json:"prompt_user"`
}

// VersionSynchronizer interface defines operations for version synchronization
type VersionSynchronizer interface {
	CheckSync(ctx context.Context) (*VersionSyncStatus, error)
	Synchronize(ctx context.Context, policy SyncPolicy) error
	GetExpectedVersion(ctx context.Context, vmVersion string) (string, error)
	ValidateSync(ctx context.Context) error
	GetManagerPair() (string, string) // Returns (version_manager, package_manager)
}

// SyncManager coordinates synchronization across multiple manager pairs
type SyncManager struct {
	synchronizers map[string]VersionSynchronizer
	policy        SyncPolicy
	logger        logger.CommonLogger
}

// PolicyEngine manages synchronization policies
type PolicyEngine struct {
	defaultPolicy  SyncPolicy
	customPolicies map[string]SyncPolicy
}
