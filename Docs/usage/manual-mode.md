# Manual Mode

Manual mode is used for migrating a single account pair (source → destination).

## When to Use

- Migrating one user's email account
- Testing migration before bulk operation
- One-time migrations

## How It Works

1. Enter source IMAP server credentials
2. Enter destination IMAP server credentials
3. MailMole connects and lists all folders
4. Review folders and press `s` to start migration
5. Watch real-time progress

## Form Fields

### Source Server
| Field | Description | Example |
|-------|-------------|---------|
| Host | IMAP server address | `mail.gmail.com` |
| Port | IMAP port (993 for SSL) | `993` |
| SSL/TLS | Enable secure connection | Check/uncheck |
| Username | Full email address | `user@gmail.com` |
| Password | Account password or app password | `xxxx xxxx xxxx xxxx` |

### Destination Server
Same fields as source server.

## Using Provider Presets

Press `F1` on the Host field to see preset providers:

```
1. Gmail / Google Workspace
2. Outlook / Office 365
3. Yahoo Mail
4. Yandex Mail
5. ProtonMail (Bridge)
6. Zoho Mail
7. iCloud Mail
8. FastMail
9. Custom (manual entry)
```

Selecting a provider auto-fills:
- Host address
- Correct port (993 for SSL, 143 for TLS)
- SSL/TLS checkbox

## What Gets Migrated

- All folders from source to destination
- Message content (full email)
- Message flags:
  - `\Seen` - Read/Unread status
  - `\Answered` - Answered status
  - `\Flagged` - Starred/Flagged status
  - `\Draft` - Draft status

## What Doesn't Get Migrated

- Contacts
- Calendars
- Filters/Rules
- Signatures

## Tips

- Test with a small folder first before migrating everything
- Ensure both servers allow IMAP access
- Use app passwords for Gmail/Outlook if 2FA is enabled
- Check destination server's storage limits before migration
