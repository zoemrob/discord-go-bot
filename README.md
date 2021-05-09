# Discord Bot built in Go
running bot:
```
go build
./discord-bot(.exe) -t <your bot token> (for windows)
pass -d to disable passthrough
```
Visit https://discord.com/developers/applications to build a bot and obtain your token.

### Commands:
```
@<bot-name> mdn <search terms>
    Queries https://developer.mozilla.org/ and returns first 3 results
```