# Bulk Migration

Bulk mode allows migrating multiple account pairs from a CSV file.

## When to Use

- Migrating many accounts at once
- Batch operations
- Server-to-server migrations

## How It Works

1. Enter global source and destination server settings
2. Provide path to CSV file containing account pairs
3. MailMole tests all connections before migration
4. Migration runs automatically for all accounts
5. Failed accounts are skipped, successful ones continue

## CSV File Format

Create a `.csv` or `.txt` file with the following format:

```text
src_user,src_pass,dst_user,dst_pass
user1@source.com,password1,user1@dest.com,password1
user2@source.com,password2,user2@dest.com,password2
user3@source.com,password3,user3@dest.com,password3
```

### File Rules

- One account pair per line
- Four comma-separated fields (no spaces)
- Lines starting with `#` are comments
- Empty lines are ignored

### Example CSV

```text
# Source mail server migration
alice@oldsrv.com,Pass123,alice@newsrv.com,NewPass456
bob@oldsrv.com,Pass123,bob@newsrv.com,NewPass456
carol@oldsrv.com,Pass123,carol@newsrv.com,NewPass456
```

## Form Fields

| Field | Description | Example |
|-------|-------------|---------|
| Global Source Host | Source mail server address | `mail.source.com` |
| Global Dest Host | Destination mail server address | `mail.dest.com` |
| Accounts File | Path to CSV/TXT file | `/path/to/accounts.csv` |

## Connection Testing

Before migration, MailMole tests all accounts. Accounts with connection failures are automatically skipped.

## Progress Tracking

The dashboard shows:
- Overall progress (messages migrated)
- Current account / Total accounts
- Transfer speed (mails/s and KB/s)
- Estimated time remaining

## State File

Bulk migrations create a `migration_state.json` file to track completed accounts and enable resume functionality.

## Tips

- Use absolute paths for CSV file
- Test with 1-2 accounts first
- Check server rate limits before large migrations
- Keep CSV file secure (contains passwords)
