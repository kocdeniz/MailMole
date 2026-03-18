# Outlook / Office 365 Setup Guide

This guide explains how to set up Outlook or Office 365 for use with MailMole.

## Requirements

- Outlook.com, Hotmail.com, Live.com account, OR
- Microsoft 365 / Office 365 account

## Option 1: Basic Authentication (Less Secure)

### For Outlook.com / Hotmail / Live.com

1. Go to Microsoft Account Security page
2. Sign in to your account
3. Enable "Two-step verification"
4. Create an app password if required

### Settings

| Setting | Value |
|---------|-------|
| Host | `outlook.office365.com` |
| Port | `993` |
| SSL/TLS | Enabled |
| Username | Your full email address |
| Password | Your Microsoft account password |

## Option 2: App Password (Recommended)

If you have 2FA enabled on your Microsoft account:

1. Go to [Microsoft Security](https://account.microsoft.com/security)
2. Click "Password and security"
3. Under "App passwords", click "Create a new app password"
4. Copy the generated password
5. Use this password in MailMole

## Office 365 / Microsoft 365

### For work/school accounts

1. Ensure IMAP access is enabled in Exchange Admin Center
2. Allow less secure apps or create an app password
3. Use the following settings:

| Setting | Value |
|---------|-------|
| Host | `outlook.office365.com` |
| Port | `993` |
| SSL/TLS | Enabled |
| Username | Your work/school email |
| Password | Your work/school password or app password |

## Troubleshooting

### "Authentication failed" with correct password
- Microsoft may require an app password
- Try creating an app password in your Microsoft account settings

### "Mailbox not available"
- Ensure you have full access permissions
- For shared mailboxes, use delegate credentials

### "Username or password wrong"
- Check if your organization blocks IMAP
- Contact your IT administrator

## IMAP Settings Summary

| Provider | Host | Port | SSL |
|----------|------|------|-----|
| Outlook.com | `outlook.office365.com` | 993 | Yes |
| Office 365 | `outlook.office365.com` | 993 | Yes |

## Additional Resources

- [Microsoft IMAP Settings](https://support.microsoft.com/en-us/account-billing/pop-imap-and-smtp-settings-for-outlook-com-7f291be4-f50f-40ed-94ee-6a0af5a62fcf)
- [App Passwords](https://support.microsoft.com/en-us/account-billing/using-app-passwords-with-apps-that-don-t-support-two-step-verification-5892908f-e153-4a59-bf6d-836d0d871d5b)
