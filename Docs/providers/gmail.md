# Gmail Setup Guide

This guide explains how to set up Gmail for use with MailMole.

## Requirements

- Gmail account (personal or Google Workspace)
- 2-Step Verification enabled (required for app passwords)

## Step 1: Enable 2-Step Verification

1. Go to [Google Account Security](https://myaccount.google.com/security)
2. Sign in to your Google account
3. Find "2-Step Verification" and click "Get Started"
4. Follow the instructions to enable 2-Step Verification

## Step 2: Generate App Password

1. After enabling 2-Step Verification, go to "App passwords"
   - Note: You may need to search for "App passwords" or go directly to: https://myaccount.google.com/apppasswords
2. Select "Mail" as the app
3. Select "Other (Custom name)" as the device
4. Enter "MailMole" as the custom name
5. Click "Generate"
6. Copy the 16-character password shown

## Step 3: Use in MailMole

1. In MailMole, select Gmail preset or enter:
   - Host: `imap.gmail.com`
   - Port: `993`
   - SSL/TLS: Enabled
2. Enter your full Gmail address as username
3. Enter the **App Password** (not your regular password)

## Troubleshooting

### "Application-specific password required"
This means you haven't created an app password yet. Follow the steps above.

### "Please log in via your web browser"
This is Google's security feature. You must use an App Password, not your regular password.

### "Username and password not accepted"
- Double-check you're using the App Password
- Make sure 2-Step Verification is enabled
- Check if the password was copied correctly (no spaces)

## IMAP Settings Summary

| Setting | Value |
|---------|-------|
| Host | `imap.gmail.com` |
| Port | `993` |
| SSL/TLS | Enabled (required) |
| Username | Your full Gmail address |
| Password | 16-character App Password |

## Additional Resources

- [Google App Passwords Help](https://support.google.com/accounts/answer/185833)
- [Enable 2-Step Verification](https://support.google.com/accounts/answer/185839)
