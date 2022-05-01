# Go Client For Sonic Search

This package contains all commands at Sonic Protocol Docs (https://github.com/valeriansaliou/sonic/blob/master/PROTOCOL.md)

Same Conn Pool and Commands infrastructure used one-to-one: (https://github.com/go-redis/redis)

For usage example go to examples folder


NOTE:
> This is dirty code package and do not use production

## Install

```
go get  github.com/uretgec/go-sonic
```

## Sonic Search Commands
```
// Base
Ping(ctx context.Context) *Cmd
Quit(ctx context.Context) *Cmd

// ChannelMode: SEARCH
Query(ctx context.Context, collection, bucket, terms string, limit, offset int, lang string) *Cmd
Suggest(ctx context.Context, collection, bucket, word string, limit int) *Cmd

// ChannelMode: INGEST
Push(ctx context.Context, collection, bucket, object, text, lang string) *Cmd
MPush(ctx context.Context, items []IngestItem) *Cmd // NOT IMPLEMENTED YET
Pop(ctx context.Context, collection, bucket, object, text string) *Cmd
MPop(ctx context.Context, items []IngestItem) *Cmd // NOT IMPLEMENTED YET
Count(ctx context.Context, collection, bucket, object string) *IntCmd
FlushCollection(ctx context.Context, collection string) *IntCmd
FlushBucket(ctx context.Context, collection, bucket string) *IntCmd
FlushObject(ctx context.Context, collection, bucket, object string) *IntCmd

// ChannelMode: CONTROL
Trigger(ctx context.Context, action, data string) *Cmd
Info(ctx context.Context) *Cmd
```

## TODO
- Add test files
- Add new examples

## Links

Sonic Search (https://github.com/valeriansaliou/sonic)

Go-Redis (https://github.com/go-redis/redis)

Go-Sonic (https://github.com/expectedsh/go-sonic)