# Example: Bulk Migration Accounts File

This file contains example account pairs for bulk migration.

## Format

```
src_user,src_pass,dst_user,dst_pass
```

## Example File

```csv
# Example accounts - replace with your own
alice@oldsrv.com,Password123,alice@newsrv.com,NewPassword456
bob@oldsrv.com,Password123,bob@newsrv.com,NewPassword456
carol@oldsrv.com,Password123,carol@newsrv.com,NewPassword456
```

## Rules

- One account pair per line
- Four comma-separated fields
- Lines starting with `#` are comments
- Empty lines are ignored
- No spaces between commas

## Usage

1. Create your accounts file
2. In MailMole, select "Bulk Migration"
3. Enter server details
4. Enter the path to this file
