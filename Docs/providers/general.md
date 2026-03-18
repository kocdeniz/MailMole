# Email Provider Setup

This section contains setup guides for various email providers.

## Supported Providers

MailMole supports the following providers via preset configuration:

- Gmail / Google Workspace
- Outlook / Office 365
- Yahoo Mail
- Yandex Mail
- ProtonMail (Bridge)
- Zoho Mail
- iCloud Mail
- FastMail
- Custom servers

## Quick Setup

When adding credentials in MailMole:

1. Press `F1` on the Host field
2. Select your provider from the list
3. MailMole auto-fills:
   - IMAP Host address
   - Correct port (993 for SSL)
   - SSL/TLS settings

4. Enter your username and password

## Provider-Specific Notes

### Gmail / Google Workspace
Requires an App Password if 2FA is enabled. See [Gmail Setup Guide](gmail.md).

### Outlook / Office 365
May require app password or modern authentication. See [Outlook Setup Guide](outlook.md).

### ProtonMail
Requires ProtonMail Bridge application running locally.

### Yahoo Mail
May require app password. Enable 2-step verification in Yahoo account settings.

### Custom Servers
Enter server details manually:
- Host: Your mail server address
- Port: 993 (SSL) or 143 (TLS)
- SSL/TLS: Enabled

## Common Requirements

Most providers require:
- IMAP access enabled
- Correct port (993 for SSL)
- App password instead of account password (if 2FA enabled)
