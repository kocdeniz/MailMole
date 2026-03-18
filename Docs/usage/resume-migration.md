# Resume Migration

MailMole can continue interrupted bulk migrations from where they left off.

## How It Works

1. When bulk migration is interrupted (crash, network loss, etc.)
2. MailMole reads `migration_state.json`
3. Skips already completed accounts
4. Continues only with pending accounts

## When to Use

- Migration was interrupted
- Network connection lost
- Server went offline
- Accidental program closure

## How to Resume

1. Select option `3` from the main menu: "Resume Previous Migration"
2. Enter the same CSV file used in the original migration
3. MailMole automatically skips completed accounts
4. Migration continues from where it stopped

## State File Location

The state file `migration_state.json` is created in the working directory.

## Understanding the State File

```json
{
  "version": 1,
  "completed": {
    "mail.src.com|user1@src.com|mail.dest.com|user1@dst.com": {
      "source": "user1@src.com",
      "destination": "user1@dst.com",
      "completed_at": "2026-03-18T17:10:00+03:00"
    }
  }
}
```

## Incremental Updates

MailMole supports incremental migration. If you run the same CSV again later:

1. New emails on source are synced
2. Already-synced emails are skipped (duplicate detection via Message-ID)
3. Deleted emails are NOT deleted on destination (safe mode)

## Clearing State

To start fresh (ignore previous state):

```bash
# Remove the state file
rm migration_state.json

# Or rename it
mv migration_state.json migration_state_old.json
```

## Tips

- Always use the same CSV file for consistent resume
- Keep the `migration_state.json` file safe
- Check the log file for details on skipped accounts
