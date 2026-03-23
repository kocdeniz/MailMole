// Package imap wraps go-imap/v2 for MailMole's connection, folder, and
// message-transfer needs. Every exported function that touches the network is
// designed to be called inside a tea.Cmd. CGO_ENABLED=0 safe — pure Go TLS.
package imap

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	imaplib "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// isIPAddress returns true when host is a bare IPv4 or IPv6 address.
// TLS certificates are almost never issued to raw IPs, so we skip
// verification in that case rather than refusing the connection.
func isIPAddress(host string) bool {
	return net.ParseIP(host) != nil
}

// ---- Config ------------------------------------------------------------------

// Config holds credentials and address for one IMAP endpoint.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	TLS      bool
}

func (c Config) Addr() string { return fmt.Sprintf("%s:%d", c.Host, c.Port) }

// ---- Client ------------------------------------------------------------------

// Client is a connected, authenticated IMAP session.
// Callers must call Close when done.
type Client struct {
	inner *imapclient.Client
	Cfg   Config
}

// TransferOutcome describes the result of attempting one message transfer.
type TransferOutcome struct {
	Migrated         bool
	SkippedDuplicate bool
	SizeBytes        int64
}

// MessageMeta holds lightweight header metadata for a message.
type MessageMeta struct {
	UID       imaplib.UID
	MessageID string
	SizeBytes int64
}

// IsRetryableError returns true when an IMAP operation should be retried.
// It covers transient network failures and common server-throttling messages.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	msg := strings.ToLower(err.Error())
	transient := []string{
		"timeout",
		"connection reset",
		"broken pipe",
		"temporarily unavailable",
		"server busy",
		"rate limit",
		"throttl",
		"too many requests",
		"try again",
	}
	for _, s := range transient {
		if strings.Contains(msg, s) {
			return true
		}
	}

	return false
}

// IsConnectionLostError returns true for hard disconnect symptoms.
func IsConnectionLostError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "closed network connection") || strings.Contains(msg, "connection reset") || strings.Contains(msg, "broken pipe") || strings.Contains(msg, "timeout")
}

// Connect dials and authenticates. Pure-Go TLS (no CGO).
func Connect(cfg Config) (*Client, error) {
	var (
		raw net.Conn
		err error
	)
	if cfg.TLS {
		if isIPAddress(cfg.Host) && os.Getenv("MAILMOLE_ALLOW_INSECURE_IP_TLS") != "1" {
			return nil, fmt.Errorf("tls verification for IP %s requires MAILMOLE_ALLOW_INSECURE_IP_TLS=1", cfg.Host)
		}
		tlsCfg := &tls.Config{
			ServerName: cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
		// Certificates are not issued to IP addresses; skip verification
		// when the caller provided a bare IP so the connection still works.
		if isIPAddress(cfg.Host) {
			tlsCfg.InsecureSkipVerify = true //nolint:gosec
		}
		raw, err = tls.Dial("tcp", cfg.Addr(), tlsCfg)
	} else {
		raw, err = net.DialTimeout("tcp", cfg.Addr(), 15*time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", cfg.Addr(), err)
	}
	c := imapclient.New(raw, &imapclient.Options{})
	if err := c.Login(cfg.Username, cfg.Password).Wait(); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("login %s@%s: %w", cfg.Username, cfg.Host, err)
	}
	return &Client{inner: c, Cfg: cfg}, nil
}

// Close logs out and tears down the underlying connection.
func (cl *Client) Close() {
	_ = cl.inner.Logout().Wait()
	_ = cl.inner.Close()
}

// Reconnect force-closes the old socket and opens a fresh authenticated session.
func (cl *Client) Reconnect() error {
	_ = cl.inner.Close()
	nc, err := Connect(cl.Cfg)
	if err != nil {
		return err
	}
	cl.inner = nc.inner
	return nil
}

// StartKeepAlive sends NOOP every 30 seconds to keep long-lived sessions active.
// The goroutine exits when stop is closed.
func (cl *Client) StartKeepAlive(stop <-chan struct{}, onError func(error)) {
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-stop:
				return
			case <-t.C:
				if err := cl.inner.Noop().Wait(); err != nil && onError != nil {
					onError(fmt.Errorf("NOOP %s: %w", cl.Cfg.Username, err))
				}
			}
		}
	}()
}

// ---- Folder operations -------------------------------------------------------

// ListFolders returns the names of all mailboxes visible to the account.
func (cl *Client) ListFolders() ([]string, error) {
	items, err := cl.inner.List("", "*", &imaplib.ListOptions{
		ReturnStatus: &imaplib.StatusOptions{NumMessages: true},
	}).Collect()
	if err != nil {
		return nil, fmt.Errorf("LIST: %w", err)
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Mailbox)
	}
	return names, nil
}

// FolderStatus returns the number of messages in a named mailbox.
func (cl *Client) FolderStatus(name string) (uint32, error) {
	data, err := cl.inner.Status(name, &imaplib.StatusOptions{
		NumMessages: true,
	}).Wait()
	if err != nil {
		return 0, fmt.Errorf("STATUS %s: %w", name, err)
	}
	if data.NumMessages == nil {
		return 0, nil
	}
	return *data.NumMessages, nil
}

// EnsureFolder creates the mailbox on dst if it does not already exist.
// Idempotent: already-exists responses from the server are silently ignored.
func (cl *Client) EnsureFolder(name string) error {
	err := cl.inner.Create(name, nil).Wait()
	if err != nil && !isAlreadyExists(err) {
		return fmt.Errorf("CREATE %s: %w", name, err)
	}
	return nil
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	msg := bytes.ToLower([]byte(err.Error()))
	for _, marker := range [][]byte{
		[]byte("alreadyexists"),
		[]byte("already exists"),
		[]byte("mailbox exists"),
	} {
		if bytes.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// ---- Message transfer --------------------------------------------------------

// FetchUIDs selects a mailbox (read-only) and returns all message UIDs.
func (cl *Client) FetchUIDs(folder string) ([]imaplib.UID, error) {
	_, err := cl.inner.Select(folder, &imaplib.SelectOptions{ReadOnly: true}).Wait()
	if err != nil {
		return nil, fmt.Errorf("SELECT %s: %w", folder, err)
	}
	searchData, err := cl.inner.UIDSearch(&imaplib.SearchCriteria{}, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("UID SEARCH %s: %w", folder, err)
	}
	return searchData.AllUIDs(), nil
}

// FetchMessageMetaBatch fetches UID + Envelope.MessageID + RFC822Size for a
// batch of UIDs in a single IMAP FETCH command.
func (cl *Client) FetchMessageMetaBatch(folder string, uids []imaplib.UID) ([]MessageMeta, error) {
	if len(uids) == 0 {
		return nil, nil
	}

	_, err := cl.inner.Select(folder, &imaplib.SelectOptions{ReadOnly: true}).Wait()
	if err != nil {
		return nil, fmt.Errorf("SELECT %s: %w", folder, err)
	}

	set := imaplib.UIDSetNum(uids...)
	bufs, err := cl.inner.Fetch(set, &imaplib.FetchOptions{
		UID:        true,
		Envelope:   true,
		RFC822Size: true,
	}).Collect()
	if err != nil {
		return nil, fmt.Errorf("FETCH meta %s: %w", folder, err)
	}

	out := make([]MessageMeta, 0, len(bufs))
	for _, b := range bufs {
		id := ""
		if b.Envelope != nil {
			id = strings.TrimSpace(b.Envelope.MessageID)
		}
		out = append(out, MessageMeta{
			UID:       b.UID,
			MessageID: normalizeMessageID(id),
			SizeBytes: b.RFC822Size,
		})
	}

	return out, nil
}

// FetchAllMessageIDCache fetches all message IDs from destination folder and
// stores them in memory for O(1) duplicate checks.
func (cl *Client) FetchAllMessageIDCache(folder string) (map[string]bool, error) {
	uids, err := cl.FetchUIDs(folder)
	if err != nil {
		return nil, err
	}
	cache := make(map[string]bool, len(uids))
	if len(uids) == 0 {
		return cache, nil
	}

	const batchSize = 100
	for i := 0; i < len(uids); i += batchSize {
		end := i + batchSize
		if end > len(uids) {
			end = len(uids)
		}
		meta, err := cl.FetchMessageMetaBatch(folder, uids[i:end])
		if err != nil {
			return nil, err
		}
		for _, m := range meta {
			if m.MessageID != "" {
				cache[m.MessageID] = true
			}
		}
	}

	return cache, nil
}

// HasMessageID checks whether destination mailbox already contains a message
// with this Message-ID. Used as a basic duplicate-prevention mechanism.
func (cl *Client) HasMessageID(folder, messageID string) (bool, error) {
	if strings.TrimSpace(messageID) == "" {
		return false, nil
	}

	_, err := cl.inner.Select(folder, &imaplib.SelectOptions{ReadOnly: true}).Wait()
	if err != nil {
		return false, fmt.Errorf("SELECT %s: %w", folder, err)
	}

	criteria := &imaplib.SearchCriteria{
		Header: []imaplib.SearchCriteriaHeaderField{{
			Key:   "Message-ID",
			Value: messageID,
		}},
	}

	data, err := cl.inner.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return false, fmt.Errorf("UID SEARCH Message-ID in %s: %w", folder, err)
	}
	return len(data.AllUIDs()) > 0, nil
}

// AppendRawMessage appends a fully formed RFC-5322 message to mailbox.
func (cl *Client) AppendRawMessage(mailbox string, raw []byte) error {
	appendCmd := cl.inner.Append(mailbox, int64(len(raw)), &imaplib.AppendOptions{})
	if _, err := io.Copy(appendCmd, bytes.NewReader(raw)); err != nil {
		return fmt.Errorf("APPEND stream %s: %w", mailbox, err)
	}
	if err := appendCmd.Close(); err != nil {
		return fmt.Errorf("APPEND close %s: %w", mailbox, err)
	}
	if _, err := appendCmd.Wait(); err != nil {
		return fmt.Errorf("APPEND wait %s: %w", mailbox, err)
	}
	return nil
}

// TransferMessage fetches the raw RFC-5322 body of one message by UID from
// this client and APPENDs it to dstFolder on dst.
// An io.Pipe keeps peak RAM usage bounded to one message at a time.
func (cl *Client) TransferMessage(uid imaplib.UID, dst *Client, dstFolder string) (TransferOutcome, error) {
	// UID FETCH <uid> BODY[]
	uidSet := imaplib.UIDSetNum(uid)
	fetchOpts := &imaplib.FetchOptions{
		BodySection: []*imaplib.FetchItemBodySection{{}}, // empty section == whole message
	}
	fetchCmd := cl.inner.Fetch(uidSet, fetchOpts)
	defer fetchCmd.Close()

	msgData := fetchCmd.Next()
	if msgData == nil {
		return TransferOutcome{}, fmt.Errorf("FETCH uid %d: no message data", uid)
	}

	// Walk the fetch items to find the body section literal
	var bodyReader io.Reader
	for {
		item := msgData.Next()
		if item == nil {
			break
		}
		if bs, ok := item.(imapclient.FetchItemDataBodySection); ok {
			bodyReader = bs.Literal
			break
		}
	}
	if bodyReader == nil {
		return TransferOutcome{}, fmt.Errorf("FETCH uid %d: no body section in response", uid)
	}

	// Buffer the message body — required to know the exact size for APPEND
	var buf bytes.Buffer
	bw := bufio.NewWriterSize(&buf, 64*1024)
	if _, err := io.Copy(bw, bodyReader); err != nil {
		return TransferOutcome{}, fmt.Errorf("FETCH uid %d: read body: %w", uid, err)
	}
	if err := bw.Flush(); err != nil {
		return TransferOutcome{}, fmt.Errorf("FETCH uid %d: flush: %w", uid, err)
	}

	// Drain any remaining fetch data (flags, etc.)
	for fetchCmd.Next() != nil {
	}

	// APPEND to destination
	size := int64(buf.Len())
	appendCmd := dst.inner.Append(dstFolder, size, &imaplib.AppendOptions{})
	if _, err := io.Copy(appendCmd, bytes.NewReader(buf.Bytes())); err != nil {
		return TransferOutcome{}, fmt.Errorf("APPEND stream %s: %w", dstFolder, err)
	}
	if err := appendCmd.Close(); err != nil {
		return TransferOutcome{}, fmt.Errorf("APPEND close %s: %w", dstFolder, err)
	}
	if _, err := appendCmd.Wait(); err != nil {
		return TransferOutcome{}, fmt.Errorf("APPEND wait %s: %w", dstFolder, err)
	}
	return TransferOutcome{Migrated: true, SizeBytes: size}, nil
}

func normalizeMessageID(v string) string {
	v = strings.TrimSpace(v)
	v = strings.Trim(v, "<>")
	return strings.ToLower(v)
}

func extractMessageID(raw []byte) string {
	headersEnd := bytes.Index(raw, []byte("\r\n\r\n"))
	if headersEnd < 0 {
		headersEnd = bytes.Index(raw, []byte("\n\n"))
	}
	if headersEnd < 0 {
		headersEnd = len(raw)
	}
	headers := string(raw[:headersEnd])

	for _, line := range strings.Split(headers, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "message-id:") {
			v := strings.TrimSpace(line[len("message-id:"):])
			v = strings.Trim(v, "<>")
			return v
		}
	}

	return ""
}
