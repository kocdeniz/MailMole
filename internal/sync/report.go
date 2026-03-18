package sync

import (
	"fmt"
	"strings"
	"time"

	"github.com/kocdeniz/mailmole/internal/imap"
)

// AccountStats stores post-migration summary metrics for one account.
type AccountStats struct {
	AccountEmail      string
	DestinationEmail  string
	MigratedMessages  int
	MigratedBytes     int64
	SkippedDuplicates int
	FolderErrors      []string
	Duration          time.Duration
}

// CompletedWithErrors indicates partial success with one or more issues.
func (s AccountStats) CompletedWithErrors() bool {
	return len(s.FolderErrors) > 0
}

// SendMigrationReport writes a summary email to destination INBOX via APPEND.
func SendMigrationReport(dst *imap.Client, stats AccountStats) error {
	status := "Completed"
	if stats.CompletedWithErrors() {
		status = "Completed with Errors"
	}

	dataSize := formatMBGB(stats.MigratedBytes)
	dur := stats.Duration.Round(time.Second)
	if dur < 0 {
		dur = 0
	}

	subject := fmt.Sprintf("[MAILMOLE] Migration Report - %s", stats.DestinationEmail)
	body := buildReportBody(status, stats, dataSize, dur)
	raw := buildPlainTextEmail(stats.DestinationEmail, subject, body)

	return dst.AppendRawMessage("INBOX", raw)
}

func buildReportBody(status string, stats AccountStats, size string, dur time.Duration) string {
	lines := []string{
		"Hello,",
		"",
		"Your MailMole migration report is ready.",
		"",
		fmt.Sprintf("Migration Status: %s", status),
		fmt.Sprintf("Account: %s", stats.AccountEmail),
		fmt.Sprintf("Total Mails Moved: %d", stats.MigratedMessages),
		fmt.Sprintf("Total Data Size: %s", size),
		fmt.Sprintf("Skipped Duplicates: %d", stats.SkippedDuplicates),
		fmt.Sprintf("Duration: %s", dur.String()),
		"Tools Used: MailMole (Go-based Migration Tool)",
	}

	if len(stats.FolderErrors) > 0 {
		lines = append(lines, "", "Folder Errors:")
		for _, e := range stats.FolderErrors {
			lines = append(lines, "- "+e)
		}
	}

	lines = append(lines, "", "Regards,", "MailMole")
	return strings.Join(lines, "\r\n")
}

func buildPlainTextEmail(to, subject, body string) []byte {
	now := time.Now().Format(time.RFC1123Z)
	headers := []string{
		"From: MailMole Report <noreply@mailmole.local>",
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		fmt.Sprintf("Date: %s", now),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
		"",
	}
	return []byte(strings.Join(headers, "\r\n"))
}

func formatMBGB(sizeBytes int64) string {
	mb := float64(sizeBytes) / (1024 * 1024)
	if mb >= 1024 {
		return fmt.Sprintf("%.2f GB", mb/1024)
	}
	return fmt.Sprintf("%.2f MB", mb)
}
