# Webhook Notifications

MailMole can send notifications to Slack, Discord, and Telegram when migrations complete.

## Overview

Configure webhooks to receive:
- Migration success/failure status
- Account statistics
- Transfer metrics
- Error reports

## Configuration File

Create `webhook.json` in the working directory:

```json
{
  "slack": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK",
  "discord": "https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK",
  "telegram": {
    "bot_token": "YOUR_BOT_TOKEN",
    "chat_id": "YOUR_CHAT_ID"
  }
}
```

You can configure one or all platforms - MailMole will send to all configured channels.

## Slack Setup

### 1. Create Slack App

1. Go to [Slack API](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Name your app and select your workspace

### 2. Enable Incoming Webhooks

1. In the left menu, click "Incoming Webhooks"
2. Toggle "Activate Incoming Webhooks" to ON
3. Click "Add New Webhook to Workspace"
4. Select a channel and click "Allow"

### 3. Copy Webhook URL

Copy the webhook URL (looks like `https://hooks.slack.com/services/XXX/YYY/ZZZ`)

### Example Notification

```
📧 Migration Complete

Accounts: 5
Total Mails: 1,250
Duration: 5m30s
```

## Discord Setup

### 1. Create Discord Webhook

1. Open Discord server settings
2. Go to "Integrations" → "Webhooks"
3. Click "New Webhook"
4. Name it (e.g., "MailMole")
5. Copy the webhook URL

### 2. Get Webhook URL

Right-click the webhook → "Copy Webhook URL"

### Example Notification

```
✅ Migration Complete

Accounts: 5
Total Mails: 1,250
Duration: 5m30s
```

## Telegram Setup

### 1. Create a Bot

1. Open Telegram and search for "@BotFather"
2. Send `/newbot`
3. Follow instructions and get your bot token (e.g., `123456789:ABCDef...`)

### 2. Get Chat ID

1. Start a chat with your bot
2. Send any message to the bot
3. Visit: `https://api.telegram.org/bot<YOUR_TOKEN>/getUpdates`
4. Find your `chat.id` in the response

### 3. Configure MailMole

```json
{
  "telegram": {
    "bot_token": "123456789:ABCDef...",
    "chat_id": "123456789"
  }
}
```

### Example Notification

```
✅ Success

📊 Statistics
├ Accounts: 5
├ Total Mails: 1,250
├ Duration: 5m30s

📝 Description
Migration completed successfully
```

## How It Works

1. Migration completes
2. MailMole reads `webhook.json` from working directory
3. Sends notification to all configured platforms
4. Notifications are sent in parallel

## Error Handling

- If a webhook fails, MailMole continues with others
- Failed notifications are logged but don't stop the migration
- Check `mailmole.log` for notification errors

## Security Notes

- Never commit `webhook.json` to version control
- Add to `.gitignore`:

```gitignore
webhook.json
```

- Consider using environment variables for production

## Future: Environment Variables

Coming soon:

```bash
export MAILMOLE_SLACK_WEBHOOK="https://hooks.slack.com/..."
export MAILMOLE_DISCORD_WEBHOOK="https://discord.com/..."
docker run -e MAILMOLE_SLACK_WEBHOOK="..." mailmole
```
