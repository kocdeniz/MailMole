# Example: Webhook Configuration

This file contains example webhook configurations.

## Slack

```json
{
  "slack": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
}
```

## Discord

```json
{
  "discord": "https://discord.com/api/webhooks/YOUR/DISCORD/WEBHOOK"
}
```

## Telegram

```json
{
  "telegram": {
    "bot_token": "123456789:ABCDefGHIjklMNOpqrsTUVwxyz123456789",
    "chat_id": "123456789"
  }
}
```

## All Platforms

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

## Usage

1. Create `webhook.json` in your working directory
2. Add your webhook URLs/tokens
3. MailMole will automatically send notifications to all configured platforms
