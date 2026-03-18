package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const DefaultStateFile = "migration_state.json"

var stateMu sync.Mutex

type migrationState struct {
	Version   int                         `json:"version"`
	Completed map[string]checkpointRecord `json:"completed"`
}

type checkpointRecord struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	CompletedAt string `json:"completed_at"`
}

func resolveStateFile() string {
	if v := os.Getenv("MAILMOLE_STATE_FILE"); v != "" {
		return v
	}
	return DefaultStateFile
}

func accountKey(p AccountPair) string {
	return fmt.Sprintf("%s|%s|%s|%s", p.SrcCfg.Host, p.SrcCfg.Username, p.DstCfg.Host, p.DstCfg.Username)
}

func loadMigrationState(path string) (*migrationState, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &migrationState{Version: 1, Completed: map[string]checkpointRecord{}}, nil
		}
		return nil, fmt.Errorf("read state file %s: %w", path, err)
	}

	var st migrationState
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, fmt.Errorf("parse state file %s: %w", path, err)
	}
	if st.Completed == nil {
		st.Completed = map[string]checkpointRecord{}
	}
	if st.Version == 0 {
		st.Version = 1
	}
	return &st, nil
}

func saveMigrationStateAtomic(path string, st *migrationState) error {
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir state dir: %w", err)
		}
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("write temp state file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}
	return nil
}

// FilterCompletedAccounts removes already-completed account pairs from pending
// work using the persisted migration state file.
func FilterCompletedAccounts(pairs []AccountPair) (pending []AccountPair, skipped int, statePath string, err error) {
	statePath = resolveStateFile()

	stateMu.Lock()
	defer stateMu.Unlock()

	st, err := loadMigrationState(statePath)
	if err != nil {
		return nil, 0, statePath, err
	}

	pending = make([]AccountPair, 0, len(pairs))
	for _, p := range pairs {
		if _, ok := st.Completed[accountKey(p)]; ok {
			skipped++
			continue
		}
		pending = append(pending, p)
	}

	return pending, skipped, statePath, nil
}

// MarkCompleted persists a completed account checkpoint atomically.
func MarkCompleted(statePath string, p AccountPair) error {
	stateMu.Lock()
	defer stateMu.Unlock()

	st, err := loadMigrationState(statePath)
	if err != nil {
		return err
	}

	st.Completed[accountKey(p)] = checkpointRecord{
		Source:      p.SrcCfg.Username,
		Destination: p.DstCfg.Username,
		CompletedAt: time.Now().UTC().Format(time.RFC3339),
	}

	return saveMigrationStateAtomic(statePath, st)
}
