# Telegram BOT for assistance 
<br />

Features
```
Fetch exchanges for current exchange rate (USD, GEL, EUR)
Write debts to google table
Read sum of debts by person
```

Used API's
```
google API v4
telegram API v5
```

Developing
- create `.env` or `.env.dev` files
- add your tokens
- run `go run main.go`

.env example
```
BOT_TOKEN='' <-- for telegram bot
GOOGLE_SHEET_ID='' <-- for google sheets
ALPHA_VANTAGE_API_KEY='' <-- for official exchange rate
```