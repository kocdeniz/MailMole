# Getting Started

This guide will walk you through your first email migration with MailMole.

## Step 1: Launch MailMole

Run MailMole from your terminal:

```bash
./mailmole
# or
go run .
```

Press any key to see the main menu.

## Step 2: Choose Migration Mode

You'll see three options:

| Option | Description |
|--------|-------------|
| `1` Manual Entry | Migrate a single account pair |
| `2` Bulk Migration | Migrate multiple accounts from CSV file |
| `3` Resume Previous | Continue an interrupted migration |

## Step 3: Enter Credentials

### For Manual Mode

Enter your source and destination server details:

```
SOURCE SERVER
- Host: mail.source.com
- Port: 993
- [x] SSL/TLS
- Username: your@email.com
- Password: ********

DESTINATION SERVER
- Host: mail.dest.com
- Port: 993
- [x] SSL/TLS
- Username: your@email.com
- Password: ********
```

### Using Provider Presets

When on the Host field, press `F1` to see preset providers:

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

Selecting a provider automatically fills in the host, port, and SSL settings.

## Step 4: Test Connection

Before migration starts, MailMole tests all account connections:

```
[INFO] Testing connections for 3 accounts...
[TEST] Testing connection for user@example.com (1/3)...
[SUCCESS] All accounts OK.
```

Accounts with connection issues are automatically skipped.

## Step 5: Start Migration

Press `s` to start the migration. You'll see real-time progress:

```
OVERALL PROGRESS  1250 / 5000 messages  Account 2/10
[████████░░░░░░░░░░░░░░░░░░░░░░░░░░░] 25%
Speed: 45.23 mails/s  125.3 KB/s
Est. remaining: 1.4 min
```

## Step 6: View Logs

All activity is logged to `mailmole.log` in the working directory.

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `Space` | Toggle SSL/TLS |
| `Enter` | Submit / Connect |
| `F1` | Provider menu |
| `s` | Start migration |
| `Esc` | Go back |
| `q` | Quit |

## Next Steps

- [Manual Mode Details](usage/manual-mode.md) - Single account migration
- [Bulk Mode Details](usage/bulk-mode.md) - Multiple accounts migration
- [Troubleshooting](troubleshooting.md) - If you encounter issues
