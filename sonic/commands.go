package sonic

import (
	"context"
)

type IngestItem struct {
	Collection string
	Bucket     string
	Object     string
	Item       string
	Lang       string
}

// Base
type BaseCmdable interface {
	Ping(ctx context.Context) *Cmd
	Quit(ctx context.Context) *Cmd
}

// ChannelMode: SEARCH
type Cmdable interface {
	Query(ctx context.Context, collection, bucket, terms string, limit, offset int, lang string) *Cmd
	Suggest(ctx context.Context, collection, bucket, word string, limit int) *Cmd

	BaseCmdable
}

// ChannelMode: INGEST
type IngestCmdable interface {
	Push(ctx context.Context, collection, bucket, object, text, lang string) *Cmd
	MPush(ctx context.Context, items []IngestItem) *Cmd
	Pop(ctx context.Context, collection, bucket, object, text string) *Cmd
	MPop(ctx context.Context, items []IngestItem) *Cmd
	Count(ctx context.Context, collection, bucket, object string) *IntCmd
	FlushCollection(ctx context.Context, collection string) *IntCmd
	FlushBucket(ctx context.Context, collection, bucket string) *IntCmd
	FlushObject(ctx context.Context, collection, bucket, object string) *IntCmd

	BaseCmdable
}

// ChannelMode: CONTROL
type ControlCmdable interface {
	Trigger(ctx context.Context, action, data string) *Cmd
	Info(ctx context.Context) *Cmd

	BaseCmdable
}

// ChannelMode: Uninitialized
type StatefulCmdable interface {
	Start(ctx context.Context, channelMode, authPassword string) *Cmd

	BaseCmdable
}

var (
	_ Cmdable = (*Client)(nil)
)

type baseCmdable func(ctx context.Context, cmd Cmder) error
type cmdable func(ctx context.Context, cmd Cmder) error
type ingestCmdable func(ctx context.Context, cmd Cmder) error
type controlCmdable func(ctx context.Context, cmd Cmder) error
type statefulCmdable func(ctx context.Context, cmd Cmder) error

//------------------------------------------------------------------------------
// QUERY <collection> <bucket> "<terms>" [LIMIT(<count>)]? [OFFSET(<count>)]? [LANG(<locale>)]?
// Return PENDING SKblCsMz <- this is marker
// After EVENT QUERY SKblCsMz user:1
func (c cmdable) Query(ctx context.Context, collection, bucket, terms string, limit, offset int, lang string) *Cmd {
	qb := NewQueryBuilder()
	qb.Command = CmdSearchQuery
	qb.Collection = collection
	qb.Bucket = bucket
	qb.Text = terms
	qb.Limit = limit
	qb.Offset = offset
	qb.Lang = lang

	cmd := NewCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

// SUGGEST <collection> <bucket> "<word>" [LIMIT(<count>)]?
// Return
func (c cmdable) Suggest(ctx context.Context, collection, bucket, word string, limit int) *Cmd {
	qb := NewQueryBuilder()
	qb.Command = CmdSearchSuggest
	qb.Collection = collection
	qb.Bucket = bucket
	qb.Text = word
	qb.Limit = limit

	cmd := NewCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

//------------------------------------------------------------------------------
// PUSH <collection> <bucket> <object> "<text>" [LANG(<locale>)]?
// Return OK
func (c ingestCmdable) Push(ctx context.Context, collection, bucket, object, text, lang string) *Cmd {
	qb := NewQueryBuilder()
	qb.Command = CmdIngestPush
	qb.Collection = collection
	qb.Bucket = bucket
	qb.Object = object
	qb.Text = text
	qb.Lang = lang

	cmd := NewCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

func (c ingestCmdable) MPush(ctx context.Context, items []IngestItem) *Cmd {
	cmd := NewCmd(ctx, "NOT_IMPLEMENTED")
	_ = c(ctx, cmd)
	return cmd
}

// POP <collection> <bucket> <object> "<text>"
func (c ingestCmdable) Pop(ctx context.Context, collection, bucket, object, text string) *Cmd {
	qb := NewQueryBuilder()
	qb.Command = CmdIngestPop
	qb.Collection = collection
	qb.Bucket = bucket
	qb.Object = object
	qb.Text = text

	cmd := NewCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

func (c ingestCmdable) MPop(ctx context.Context, items []IngestItem) *Cmd {
	cmd := NewCmd(ctx, "NOT_IMPLEMENTED")
	_ = c(ctx, cmd)
	return cmd
}

// COUNT <collection> [<bucket> [<object>]?]?
// Return RESULT 2
func (c ingestCmdable) Count(ctx context.Context, collection, bucket, object string) *IntCmd {
	qb := NewQueryBuilder()
	qb.Command = CmdIngestCount
	qb.Collection = collection
	qb.Bucket = bucket
	qb.Object = object

	cmd := NewIntCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

// FLUSHC <collection>
// Return RESULT 1
func (c ingestCmdable) FlushCollection(ctx context.Context, collection string) *IntCmd {
	qb := NewQueryBuilder()
	qb.Command = CmdIngestFlushc
	qb.Collection = collection

	cmd := NewIntCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

// FLUSHB <collection> <bucket>
// Return RESULT 1
func (c ingestCmdable) FlushBucket(ctx context.Context, collection, bucket string) *IntCmd {
	qb := NewQueryBuilder()
	qb.Command = CmdIngestFlushb
	qb.Collection = collection
	qb.Bucket = bucket

	cmd := NewIntCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

// FLUSHO <collection> <bucket> <object>
// Return RESULT 0
func (c ingestCmdable) FlushObject(ctx context.Context, collection, bucket, object string) *IntCmd {
	qb := NewQueryBuilder()
	qb.Command = CmdIngestFlusho
	qb.Collection = collection
	qb.Bucket = bucket
	qb.Object = object

	cmd := NewIntCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

//------------------------------------------------------------------------------
// TRIGGER [<action: consolidate, backup, restore>]? [<data: backup, restore>]?
// Return OK
func (c controlCmdable) Trigger(ctx context.Context, action, data string) *Cmd {
	qb := NewQueryBuilder()
	qb.Command = CmdControlTrigger
	qb.Action = action
	qb.Data = data

	cmd := NewCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

// INFO
// Return RESULT uptime(118725) clients_connected(2) commands_total(54) command_latency_best(1) command_latency_worst(25) kv_open_count(0) fst_open_count(0) fst_consolidate_count(0)
func (c controlCmdable) Info(ctx context.Context) *Cmd {
	qb := NewQueryBuilder()
	qb.Command = CmdControlInfo

	cmd := NewCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

//------------------------------------------------------------------------------
// START <mode: search,ingest> <password: channel.auth_password>
// Return STARTED search protocol(1) buffer(20000)
func (c statefulCmdable) Start(ctx context.Context, channelMode, authPassword string) *Cmd {
	cmd := NewCmd(ctx, CmdSearchStart, channelMode, authPassword)
	_ = c(ctx, cmd)
	return cmd
}

//------------------------------------------------------------------------------
// PING
// Return PONG
func (c baseCmdable) Ping(ctx context.Context) *Cmd {
	qb := NewQueryBuilder()
	qb.Command = CmdPing

	cmd := NewCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}

// QUIT
// Return ENDED quit
func (c baseCmdable) Quit(ctx context.Context) *Cmd {
	qb := NewQueryBuilder()
	qb.Command = CmdQuit

	cmd := NewCmd(ctx, qb.Encode()...)
	_ = c(ctx, cmd)
	return cmd
}
