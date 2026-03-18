package sync

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	imaplib "github.com/emersion/go-imap/v2"
	"github.com/kocdeniz/mailmole/internal/imap"
)

// StatusKind classifies each update sent to the UI.
type StatusKind int

const (
	StatusAccountStart StatusKind = iota
	StatusAccountDone
	StatusAccountError
	StatusFolderStart
	StatusFolderDone
	StatusMessageCopied
	StatusMessageSkipped
	StatusRetrying
	StatusReportPlaced
	StatusStateSaving
	StatusStateSaved
	StatusMigrationDone
)

// StatusUpdateMsg is posted from the migration goroutine to Bubble Tea.
type StatusUpdateMsg struct {
	Kind            StatusKind
	Account         string
	Folder          string
	Copied          int
	Total           int
	MovedBytesDelta int64
	SkippedDelta    int
	Err             error
	RetryAfterS     int
	Stats           *AccountStats
}

// EngineConfig controls transfer speed and worker parallelism.
type EngineConfig struct {
	MessageDelay  time.Duration
	FolderWorkers int
	StateFile     string
	UseCheckpoint bool
}

func defaultEngineConfig() EngineConfig {
	delay := 0 * time.Millisecond
	if raw := os.Getenv("MAILMOLE_MESSAGE_DELAY_MS"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			delay = time.Duration(v) * time.Millisecond
		}
	}

	workers := 3
	if raw := os.Getenv("MAILMOLE_FOLDER_WORKERS"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			workers = v
		}
	}

	if workers > 3 {
		workers = 3
	}

	return EngineConfig{
		MessageDelay:  delay,
		FolderWorkers: workers,
		StateFile:     resolveStateFile(),
		UseCheckpoint: false,
	}
}

// RunMigration launches the migration engine with default settings.
func RunMigration(pairs []AccountPair) (tea.Cmd, <-chan StatusUpdateMsg) {
	return RunMigrationWithConfig(pairs, defaultEngineConfig())
}

// RunMigrationWithCheckpoint enables per-account persistence to state file.
func RunMigrationWithCheckpoint(pairs []AccountPair, stateFile string) (tea.Cmd, <-chan StatusUpdateMsg) {
	cfg := defaultEngineConfig()
	cfg.UseCheckpoint = true
	if stateFile != "" {
		cfg.StateFile = stateFile
	}
	return RunMigrationWithConfig(pairs, cfg)
}

// RunMigrationWithConfig launches migration and returns first read cmd + channel.
func RunMigrationWithConfig(pairs []AccountPair, cfg EngineConfig) (tea.Cmd, <-chan StatusUpdateMsg) {
	ch := make(chan StatusUpdateMsg, 256)
	go runEngine(pairs, cfg, ch)
	return drainCmd(ch), ch
}

func drainCmd(ch <-chan StatusUpdateMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return StatusUpdateMsg{Kind: StatusMigrationDone}
		}
		return msg
	}
}

func WaitForNext(ch <-chan StatusUpdateMsg) tea.Cmd { return drainCmd(ch) }

func runEngine(pairs []AccountPair, cfg EngineConfig, ch chan<- StatusUpdateMsg) {
	defer close(ch)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, pair := range pairs {
		label := pair.SrcCfg.Username
		sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountStart, Account: label})

		stats, err := migrateAccount(pair, label, cfg, ch, r)
		if err != nil {
			sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountError, Account: label, Err: err})
			if stats != nil {
				sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountDone, Account: label, Stats: stats})
			}
			continue
		}

		if cfg.UseCheckpoint {
			sendStatus(ch, StatusUpdateMsg{Kind: StatusStateSaving, Account: label})
			if err := MarkCompleted(cfg.StateFile, pair); err != nil {
				sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountError, Account: label, Err: fmt.Errorf("state save: %w", err)})
			} else {
				sendStatus(ch, StatusUpdateMsg{Kind: StatusStateSaved, Account: label})
			}
		}

		sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountDone, Account: label, Stats: stats})
	}
}

func migrateAccount(pair AccountPair, label string, cfg EngineConfig, ch chan<- StatusUpdateMsg, r *rand.Rand) (*AccountStats, error) {
	stats := &AccountStats{AccountEmail: pair.SrcCfg.Username, DestinationEmail: pair.DstCfg.Username}
	started := time.Now()

	// Control connections for folder listing + final report append.
	srcCtl, err := connectWithRetry(pair.SrcCfg, label, "", "connect source", ch, r, nil)
	if err != nil {
		stats.Duration = time.Since(started)
		stats.FolderErrors = append(stats.FolderErrors, fmt.Sprintf("source connection: %v", err))
		return stats, err
	}
	defer srcCtl.Close()

	dstCtl, err := connectWithRetry(pair.DstCfg, label, "", "connect destination", ch, r, nil)
	if err != nil {
		stats.Duration = time.Since(started)
		stats.FolderErrors = append(stats.FolderErrors, fmt.Sprintf("destination connection: %v", err))
		return stats, err
	}
	defer dstCtl.Close()

	var folders []string
	err = retryIMAP(label, "", "list folders", ch, r, nil, nil, func() error {
		f, e := srcCtl.ListFolders()
		if e != nil {
			return e
		}
		folders = f
		return nil
	})
	if err != nil {
		stats.Duration = time.Since(started)
		stats.FolderErrors = append(stats.FolderErrors, fmt.Sprintf("list folders: %v", err))
		_ = trySendReport(dstCtl, stats, ch, label)
		return stats, err
	}

	workers := cfg.FolderWorkers
	if workers > len(folders) {
		workers = len(folders)
	}
	if workers < 1 {
		workers = 1
	}

	var (
		mu          sync.Mutex
		adaptiveGap = cfg.MessageDelay
		jobs        = make(chan string)
		wg          sync.WaitGroup
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			srcW, err := connectWithRetry(pair.SrcCfg, label, "", "worker connect source", ch, r, func() {
				increaseDelayOnRateLimit(&mu, &adaptiveGap)
			})
			if err != nil {
				sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountError, Account: label, Err: err})
				return
			}
			defer srcW.Close()

			dstW, err := connectWithRetry(pair.DstCfg, label, "", "worker connect destination", ch, r, func() {
				increaseDelayOnRateLimit(&mu, &adaptiveGap)
			})
			if err != nil {
				sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountError, Account: label, Err: err})
				return
			}
			defer dstW.Close()

			stopKeepAlive := make(chan struct{})
			defer close(stopKeepAlive)
			srcW.StartKeepAlive(stopKeepAlive, func(err error) {
				sendStatus(ch, StatusUpdateMsg{Kind: StatusRetrying, Account: label, Err: err, RetryAfterS: 30})
			})
			dstW.StartKeepAlive(stopKeepAlive, func(err error) {
				sendStatus(ch, StatusUpdateMsg{Kind: StatusRetrying, Account: label, Err: err, RetryAfterS: 30})
			})

			for folder := range jobs {
				if err := migrateFolder(srcW, dstW, folder, label, stats, ch, r, &mu, &adaptiveGap); err != nil {
					mu.Lock()
					stats.FolderErrors = append(stats.FolderErrors, fmt.Sprintf("%s: %v", folder, err))
					mu.Unlock()
					sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountError, Account: label, Folder: folder, Err: err})
				}
			}
		}()
	}

	for _, folder := range folders {
		jobs <- folder
	}
	close(jobs)
	wg.Wait()

	stats.Duration = time.Since(started)
	_ = trySendReport(dstCtl, stats, ch, label)
	return stats, nil
}

func connectWithRetry(cfg imap.Config, account, folder, op string, ch chan<- StatusUpdateMsg, r *rand.Rand, onRateLimit func()) (*imap.Client, error) {
	var out *imap.Client
	err := retryIMAP(account, folder, op, ch, r, onRateLimit, nil, func() error {
		c, err := imap.Connect(cfg)
		if err != nil {
			return err
		}
		out = c
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return out, nil
}

func trySendReport(dst *imap.Client, stats *AccountStats, ch chan<- StatusUpdateMsg, account string) error {
	if dst == nil {
		return fmt.Errorf("destination unavailable")
	}
	if err := SendMigrationReport(dst, *stats); err != nil {
		sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountError, Account: account, Err: fmt.Errorf("report: %w", err)})
		return err
	}
	sendStatus(ch, StatusUpdateMsg{Kind: StatusReportPlaced, Account: account})
	return nil
}

func migrateFolder(src, dst *imap.Client, folder, account string, stats *AccountStats, ch chan<- StatusUpdateMsg, r *rand.Rand, mu *sync.Mutex, adaptiveGap *time.Duration) error {
	reconnectBoth := func() error {
		if err := src.Reconnect(); err != nil {
			return err
		}
		if err := dst.Reconnect(); err != nil {
			return err
		}
		return nil
	}

	if err := retryIMAP(account, folder, "ensure folder", ch, r, func() {
		increaseDelayOnRateLimit(mu, adaptiveGap)
	}, reconnectBoth, func() error {
		return dst.EnsureFolder(folder)
	}); err != nil {
		return fmt.Errorf("ensure folder: %w", err)
	}

	var uids []imaplib.UID
	if err := retryIMAP(account, folder, "fetch UIDs", ch, r, func() {
		increaseDelayOnRateLimit(mu, adaptiveGap)
	}, reconnectBoth, func() error {
		v, err := src.FetchUIDs(folder)
		if err != nil {
			return err
		}
		uids = v
		return nil
	}); err != nil {
		return fmt.Errorf("fetch UIDs: %w", err)
	}

	total := len(uids)
	if total == 0 {
		return nil
	}

	// Build destination Message-ID cache once per folder (O(1) speedup).
	var dstCache map[string]bool
	if err := retryIMAP(account, folder, "build destination cache", ch, r, func() {
		increaseDelayOnRateLimit(mu, adaptiveGap)
	}, reconnectBoth, func() error {
		v, err := dst.FetchAllMessageIDCache(folder)
		if err != nil {
			return err
		}
		dstCache = v
		return nil
	}); err != nil {
		return fmt.Errorf("destination cache: %w", err)
	}

	sendStatus(ch, StatusUpdateMsg{Kind: StatusFolderStart, Account: account, Folder: folder, Copied: 0, Total: total})

	const batchSize = 50
	copied := 0

	for i := 0; i < len(uids); i += batchSize {
		end := i + batchSize
		if end > len(uids) {
			end = len(uids)
		}
		chunk := uids[i:end]

		var meta []imap.MessageMeta
		err := retryIMAP(account, folder, "fetch source headers batch", ch, r, func() {
			increaseDelayOnRateLimit(mu, adaptiveGap)
		}, reconnectBoth, func() error {
			m, e := src.FetchMessageMetaBatch(folder, chunk)
			if e != nil {
				return e
			}
			meta = m
			return nil
		})
		if err != nil {
			return fmt.Errorf("fetch source metadata: %w", err)
		}

		for _, m := range meta {
			if m.MessageID != "" && dstCache[m.MessageID] {
				mu.Lock()
				stats.SkippedDuplicates++
				mu.Unlock()
				sendStatus(ch, StatusUpdateMsg{Kind: StatusMessageSkipped, Account: account, Folder: folder, SkippedDelta: 1})
				continue
			}

			var outcome imap.TransferOutcome
			err := retryIMAP(account, folder, fmt.Sprintf("transfer uid %d", m.UID), ch, r, func() {
				increaseDelayOnRateLimit(mu, adaptiveGap)
			}, reconnectBoth, func() error {
				o, e := src.TransferMessage(m.UID, dst, folder)
				if e != nil {
					return e
				}
				outcome = o
				return nil
			})
			if err != nil {
				sendStatus(ch, StatusUpdateMsg{Kind: StatusAccountError, Account: account, Folder: folder, Err: fmt.Errorf("uid %d: %w", m.UID, err)})
				time.Sleep(currentDelay(mu, adaptiveGap))
				continue
			}

			if outcome.Migrated {
				copied++
				mu.Lock()
				stats.MigratedMessages++
				stats.MigratedBytes += outcome.SizeBytes
				if m.MessageID != "" {
					dstCache[m.MessageID] = true
				}
				mu.Unlock()
				sendStatus(ch, StatusUpdateMsg{Kind: StatusMessageCopied, Account: account, Folder: folder, Copied: copied, Total: total, MovedBytesDelta: outcome.SizeBytes})
			}

			time.Sleep(currentDelay(mu, adaptiveGap))
		}
	}

	sendStatus(ch, StatusUpdateMsg{Kind: StatusFolderDone, Account: account, Folder: folder, Copied: copied, Total: total})

	// Free per-folder cache memory before returning.
	for k := range dstCache {
		delete(dstCache, k)
	}

	return nil
}

// retryIMAP retries transient errors with backoff 2s, 5s, 10s plus jitter.
func retryIMAP(account, folder, op string, ch chan<- StatusUpdateMsg, r *rand.Rand, onRateLimit func(), onNetworkRetry func() error, fn func() error) error {
	backoff := []time.Duration{2 * time.Second, 5 * time.Second, 10 * time.Second}

	err := fn()
	if err == nil {
		return nil
	}

	for i, wait := range backoff {
		if !imap.IsRetryableError(err) {
			return err
		}
		if isRateLimitError(err) && onRateLimit != nil {
			onRateLimit()
		}
		if imap.IsConnectionLostError(err) && onNetworkRetry != nil {
			if recErr := onNetworkRetry(); recErr != nil {
				return fmt.Errorf("reconnect failed: %w", recErr)
			}
		}

		jitter := time.Duration(r.Intn(500)) * time.Millisecond
		delay := wait + jitter
		sendStatus(ch, StatusUpdateMsg{Kind: StatusRetrying, Account: account, Folder: folder, Err: fmt.Errorf("%s: %w", op, err), RetryAfterS: int(delay.Round(time.Second).Seconds())})

		time.Sleep(delay)
		err = fn()
		if err == nil {
			return nil
		}
		if i == len(backoff)-1 {
			return err
		}
	}

	return err
}

func isRateLimitError(err error) bool {
	msg := strings.ToLower(err.Error())
	for _, s := range []string{"rate limit", "throttl", "too many requests", "too many login", "slow down"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

func increaseDelayOnRateLimit(mu *sync.Mutex, d *time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	if *d < 250*time.Millisecond {
		*d += 20 * time.Millisecond
	}
}

func currentDelay(mu *sync.Mutex, d *time.Duration) time.Duration {
	mu.Lock()
	v := *d
	mu.Unlock()
	if v < 0 {
		return 0
	}
	return v
}

func sendStatus(ch chan<- StatusUpdateMsg, msg StatusUpdateMsg) { ch <- msg }
