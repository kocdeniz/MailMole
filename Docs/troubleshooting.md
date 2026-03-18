# Troubleshooting

Common issues and their solutions.

## Connection Errors

### "Connection refused"

**Cause:** IMAP port is blocked or server is down.

**Solutions:**
- Verify the server address is correct
- Check if IMAP port (993 for SSL, 143 for TLS) is open
- Ensure the mail server IMAP service is running
- Try pinging the server: `ping mail.example.com`

### "Authentication failed"

**Cause:** Invalid credentials.

**Solutions:**
- Verify username and password
- Use app password for Gmail/Outlook if 2FA is enabled
- Check if the account has IMAP access enabled

### "Timeout"

**Cause:** Network issues or slow server response.

**Solutions:**
- Check internet connection
- Try again in a few minutes
- Server might be rate-limiting connections

## Gmail Specific Issues

### "Application-specific password required"

**Cause:** Gmail account has 2-Step Verification enabled.

**Solution:**
1. Go to [Google Account Security](https://myaccount.google.com/security)
2. Enable "2-Step Verification" (if not already)
3. Go to "App passwords"
4. Generate a new app password for "Mail"
5. Use this password instead of your regular password

### "Please log in via your web browser"

**Cause:** Google blocking "less secure apps".

**Solution:**
- Use an App Password (see above) instead
- Or enable "Less secure app access" (not recommended)

## Outlook/Office 365 Issues

### "Authentication failed" with correct password

**Cause:** May need app password or modern auth.

**Solution:**
1. Go to Microsoft Account Security
2. Enable "Two-step verification"
3. Create an app password
4. Use the app password in MailMole

### "Mailbox not available"

**Cause:** Shared mailbox or permission issues.

**Solution:**
- Ensure you have full access permissions to the mailbox
- For shared mailboxes, use delegate credentials

## Migration Errors

### "Folder creation failed"

**Cause:** Destination server restrictions.

**Solutions:**
- Check if you have permission to create folders
- Folder name might contain invalid characters
- Server might have folder name restrictions

### "Message transfer failed"

**Cause:** Various possible issues.

**Solutions:**
- Network interruption - retry usually works
- Server throttling - wait and retry
- Message too large - check size limits
- Invalid message format - some emails may fail

### "Duplicate detection not working"

**Cause:** Messages lack Message-ID header.

**Solutions:**
- This is normal for some legacy systems
- Run migration again - will skip already transferred
- Accept that some duplicates may occur

## Performance Issues

### Migration is slow

**Possible causes:**
- Server rate limits
- Network bandwidth
- Large email attachments

**Solutions:**
- Be patient - rate limits are respected
- Consider migrating in off-peak hours
- Large attachments slow down transfer

### High memory usage

MailMole is designed for low memory usage. If you see high memory:
- Check for very large folders
- Server might be sending data inefficiently

## Log Files

Check `mailmole.log` for detailed error information:

```bash
# View recent errors
grep -i error mailmole.log

# View specific account logs
grep "user@example.com" mailmole.log
```

## Getting Help

If you encounter an issue not listed here:

1. Check the log file for detailed error
2. Search existing issues on [GitHub](https://github.com/kocdeniz/MailMole/issues)
3. Create a new issue with:
   - Error message from logs
   - Steps to reproduce
   - Mail server types (Gmail, Outlook, etc.)
