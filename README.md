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
@<bot-name> help
    Lists commands

@<bot-name> mdn <search terms>
    Queries https://developer.mozilla.org/ and returns a link to results
    
@<bot-name> go <search terms>
    Queries https://pkg.do.dev/ and returns first 3 results
    
@<bot-name> gh <search terms>
    Searches Github for relevant git repositories and returns the first 3 results

@<bot-name> insult me
    If you dare
```