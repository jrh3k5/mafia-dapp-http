# Mafia Mock Server

This is a mock server to help emulate the game engine built using [mafia-dapp](https://github.com/jrh3k5/mafia-dapp) and the UI built in [mafia-dapp-ui](https://github.com/jrh3k5/mafia-dapp-ui), but implemented with an HTTP service to allow for local testing. This removes the need to have multiple wallets running concurrently, as each tab within the browser can have its own game session with the state stored in this server.

To run the server, execute:

```
go run main.go
```

Note that the game state is stored in-memory, so cycling the server will erase all game state.