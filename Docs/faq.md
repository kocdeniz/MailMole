# Frequently Asked Questions

Common questions about MailMole.

## General

### What is MailMole?

MailMole is a terminal-based IMAP-to-IMAP email migration tool. It allows you to transfer emails between any two mail servers that support IMAP protocol.

### What does MailMole support?

- Migrating between any IMAP-compatible mail servers
- Gmail, Outlook, Yahoo, Yandex, iCloud, and custom servers
- Bulk migration from CSV file
- Preserving message flags (read/unread, starred, etc.)
- Resume interrupted migrations
- Real-time progress tracking

### What doesn't MailMole support?

- POP3 mailboxes
- Migrating contacts or calendars
- Migrating filters or rules
- OAuth2 authentication (planned)

## Security

### Are my passwords saved?

**No.** Passwords are entered directly in the terminal and are:
- Never saved to disk
- Never sent to any external server
- Only used for the current IMAP connection

### Is the connection secure?

Yes. MailMole:
- Supports SSL/TLS encryption
- Uses hostname verification for TLS
- Falls back to certificate verification for IP addresses

### What about the log file?

The log file (`mailmole.log`) contains:
- Connection timestamps
- Migration progress
- Error messages

It does NOT contain:
- Passwords
- Email content
- Attachments

## Migration

### Will duplicate emails be created?

**No.** MailMole uses Message-ID to detect duplicates:
- Before transferring, it checks if the message exists
- Existing messages are skipped
- Safe to run multiple times

### What happens if migration is interrupted?

MailMole supports resume:
- Completed accounts are saved to `migration_state.json`
- Select "Resume Previous" to continue
- Already completed accounts are skipped automatically

### Can I migrate only specific folders?

Currently, all folders are migrated. Folder filtering is planned for future versions.

### Will my read/unread status be preserved?

**Yes.** MailMole preserves:
- `\Seen` - Read/Unread status
- `\Answered` - Answered status
- `\Flagged` - Starred/Flagged status
- `\Draft` - Draft status

### How fast is migration?

Speed depends on:
- Network bandwidth
- Server response time
- Message size

Typical speeds: 30-100 emails per second

## Technical

### What are the system requirements?

- Linux, macOS, or Windows
- Go 1.24+ (to build from source)
- IMAP access to both mail servers
- Network connectivity

### Can I run this on a server?

**Yes.** MailMole:
- Has a CLI interface (no GUI needed)
- Creates minimal resource usage
- Can run in background
- Supports resume for long migrations

### Does it use a database?

**No.** MailMole uses simple JSON files:
- `migration_state.json` - tracks completed migrations
- No database installation required

### Can I run multiple instances?

Not recommended. Use separate directories for different migrations to avoid state file conflicts.

## Troubleshooting

### Gmail says "Application-specific password required"

Gmail requires app passwords for accounts with 2-Step Verification. See [Gmail Setup Guide](providers/gmail.md).

### Outlook authentication fails

Microsoft may require app passwords. See [Outlook Setup Guide](providers/outlook.md).

### Connection times out

- Check internet connection
- Try again in a few minutes
- Server might be rate-limiting

### Migration is slow

- Normal for large attachments
- Server may have rate limits
- Consider running during off-peak hours

## Contributing

### How can I contribute?

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

### Where is the source code?

[github.com/kocdeniz/MailMole](https://github.com/kocdeniz/MailMole)

### How do I report bugs?

Create an issue on GitHub with:
- Error message
- Steps to reproduce
- Mail server types
