// Via: https://github.com/go-redis/redis
package sonic

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/uretgec/go-sonic/pool"
	"github.com/uretgec/go-sonic/proto"
)

type baseClient struct {
	opt      *Options
	connPool pool.Pooler

	onClose func() error
}

func newBaseClient(opt *Options, connPool pool.Pooler) *baseClient {
	return &baseClient{
		opt:      opt,
		connPool: connPool,
	}
}

func (c *baseClient) clone() *baseClient {
	clone := *c
	return &clone
}

func (c *baseClient) withTimeout(timeout time.Duration) *baseClient {
	opt := c.opt.clone()
	opt.ReadTimeout = timeout
	opt.WriteTimeout = timeout

	clone := c.clone()
	clone.opt = opt

	return clone
}

/*
func (c *baseClient) newConn(ctx context.Context) (*pool.Conn, error) {
	fmt.Printf("KULLANDI %s\n", "newConn")
	cn, err := c.connPool.NewConn(ctx)
	if err != nil {
		return nil, err
	}

	err = c.initConn(ctx, cn)
	if err != nil {
		_ = c.connPool.CloseConn(cn)
		return nil, err
	}

	return cn, nil
}*/

func (c *baseClient) getConn(ctx context.Context) (*pool.Conn, error) {
	cn, err := c._getConn(ctx)
	if err != nil {
		return nil, err
	}

	return cn, nil
}

func (c *baseClient) _getConn(ctx context.Context) (*pool.Conn, error) {
	cn, err := c.connPool.Get(ctx)
	if err != nil {
		return nil, err
	}
	if cn.Inited {
		return cn, nil
	}

	if err := c.initConn(ctx, cn); err != nil {
		c.connPool.Remove(ctx, cn, err)
		if err := errors.Unwrap(err); err != nil {
			return nil, err
		}
		return nil, err
	}

	return cn, nil
}

func (c *baseClient) initConn(ctx context.Context, cn *pool.Conn) error {
	if cn.Inited {
		return nil
	}
	cn.Inited = true

	if c.opt.AuthPassword == "" && !c.opt.readOnly && c.opt.OnConnect == nil {
		return nil
	}

	connPool := pool.NewSingleConnPool(c.connPool, cn)
	conn := newConn(ctx, c.opt, connPool)

	// Connect Sonic Server First Time
	bufferSize, err := conn.Start(ctx, c.opt.ChannelMode, c.opt.AuthPassword).Int()
	if err != nil {
		bufferSize, err = conn.Start(ctx, c.opt.ChannelMode, c.opt.AuthPassword).Int()
		if err != nil {
			return err
		}
		//return err
	}

	// Set MaxBufferedSize
	c.opt.MaxBufferedSize = bufferSize
	fmt.Printf("INIT CONN %#v\n", c.connPool.Stats())
	if c.opt.OnConnect != nil {
		return c.opt.OnConnect(ctx, conn)
	}
	return nil
}

func (c *baseClient) releaseConn(ctx context.Context, cn *pool.Conn, err error) {
	if isBadConn(err, false, c.opt.Addr) {
		c.connPool.Remove(ctx, cn, err)
	} else {
		c.connPool.Put(ctx, cn)
	}
}

func (c *baseClient) withConn(ctx context.Context, fn func(context.Context, *pool.Conn) error) error {
	cn, err := c.getConn(ctx)
	if err != nil {
		return err
	}

	defer func() {
		c.releaseConn(ctx, cn, err)
	}()

	done := ctx.Done() //nolint:ifshort

	if done == nil {
		err = fn(ctx, cn)
		return err
	}

	errc := make(chan error, 1)
	go func() { errc <- fn(ctx, cn) }()

	select {
	case <-done:
		_ = cn.Close()
		// Wait for the goroutine to finish and send something.
		<-errc

		err = ctx.Err()
		return err
	case err = <-errc:
		return err
	}
}

func (c *baseClient) process(ctx context.Context, cmd Cmder) error {
	var lastErr error
	for attempt := 0; attempt <= c.opt.MaxRetries; attempt++ {
		attempt := attempt

		retry, err := c._process(ctx, cmd, attempt)
		if err == nil || !retry {
			return err
		}

		lastErr = err
	}
	return lastErr
}

func (c *baseClient) _process(ctx context.Context, cmd Cmder, attempt int) (bool, error) {
	if attempt > 0 {
		if err := Sleep(ctx, c.retryBackoff(attempt)); err != nil {
			return false, err
		}
	}

	retryTimeout := uint32(1)
	err := c.withConn(ctx, func(ctx context.Context, cn *pool.Conn) error {
		err := cn.WithWriter(ctx, c.opt.WriteTimeout, func(wr *proto.Writer) error {
			return writeCmd(wr, cmd)
		})
		if err != nil {
			return err
		}

		err = cn.WithReader(ctx, c.cmdTimeout(cmd), cmd.readReply)
		if err != nil {
			if cmd.readTimeout() == nil {
				atomic.StoreUint32(&retryTimeout, 1)
			}
			return err
		}

		return nil
	})
	if err == nil {
		return false, nil
	}

	retry := shouldRetry(err, atomic.LoadUint32(&retryTimeout) == 1)
	return retry, err
}

func (c *baseClient) retryBackoff(attempt int) time.Duration {
	return RetryBackoff(attempt, c.opt.MinRetryBackoff, c.opt.MaxRetryBackoff)
}

func (c *baseClient) cmdTimeout(cmd Cmder) time.Duration {
	if timeout := cmd.readTimeout(); timeout != nil {
		t := *timeout
		if t == 0 {
			return 0
		}
		return t + 10*time.Second
	}
	return c.opt.ReadTimeout
}

// Close closes the client, releasing any open resources.
//
// It is rare to Close a Client, as the Client is meant to be
// long-lived and shared between many goroutines.
func (c *baseClient) Close() error {
	var firstErr error
	if c.onClose != nil {
		if err := c.onClose(); err != nil {
			firstErr = err
		}
	}
	if err := c.connPool.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

/*
func (c *baseClient) getAddr() string {
	return c.opt.Addr
}
*/
//------------------------------------------------------------------------------

// Client is a Sonic client representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple
// goroutines.
type Client struct {
	*baseClient
	baseCmdable
	cmdable
	ingestCmdable
	controlCmdable
	ctx context.Context
}

// NewClient returns a client to the Sonic Server specified by Options.
func NewClient(opt *Options) *Client {
	opt.init()

	c := Client{
		baseClient: newBaseClient(opt, newConnPool(opt)),
		ctx:        context.Background(),
	}
	if opt.ChannelMode == ChannelSearch {
		c.cmdable = c.Process
	} else if opt.ChannelMode == ChannelIngest {
		c.ingestCmdable = c.Process
	} else if opt.ChannelMode == ChannelControl {
		c.controlCmdable = c.Process
	}

	c.baseCmdable = c.Process

	return &c
}

func (c *Client) clone() *Client {
	clone := *c
	clone.cmdable = clone.Process
	return &clone
}

func (c *Client) WithTimeout(timeout time.Duration) *Client {
	clone := c.clone()
	clone.baseClient = c.baseClient.withTimeout(timeout)
	return clone
}

func (c *Client) Context() context.Context {
	return c.ctx
}

func (c *Client) WithContext(ctx context.Context) *Client {
	if ctx == nil {
		panic("nil context")
	}
	clone := c.clone()
	clone.ctx = ctx
	return clone
}

func (c *Client) Conn(ctx context.Context) *Conn {
	return newConn(ctx, c.opt, pool.NewStickyConnPool(c.connPool))
}

// Do creates a Cmd from the args and processes the cmd.
func (c *Client) Do(ctx context.Context, args ...string) *Cmd {
	cmd := NewCmd(ctx, args...)
	_ = c.Process(ctx, cmd)
	return cmd
}

func (c *Client) Process(ctx context.Context, cmd Cmder) error {
	retErr := c.baseClient.process(ctx, cmd)
	cmd.SetErr(retErr)
	return retErr
}

// Options returns read-only Options that were used to create the client.
func (c *Client) Options() *Options {
	return c.opt
}

// Check push content is too big for buffered size
func (c *Client) IsPushContentReady(str string) bool {
	str = sanitize(str)
	return utf8.RuneCountInString(str) >= c.opt.MaxBufferedSize
}

// Chunks content
func (c *Client) SplitPushContent(str string) []string {
	var splits []string

	str = sanitize(str)
	limitter := c.opt.MaxBufferedSize

	var l, r int
	for l, r = 0, limitter; r < len(str); l, r = r, r+limitter {
		for !utf8.RuneStart(str[r]) {
			r--
		}
		splits = append(splits, str[l:r])
	}
	splits = append(splits, str[l:])
	return splits
}

// Suggest word checker
// Suggest only one word, not multiple
func (c *Client) IsSuggestWordReady(words string) bool {
	index := findDummyIndex(words)
	return index == -1
}

func (c *Client) GetSuggestWord(words string) string {
	index := findDummyIndex(words)
	if index != -1 {
		return words
	}

	return truncateText(words, index)
}

// Pool Stats
type PoolStats pool.Stats

// PoolStats returns connection pool stats.
func (c *Client) PoolStats() *PoolStats {
	stats := c.connPool.Stats()
	return (*PoolStats)(stats)
}

//------------------------------------------------------------------------------

type conn struct {
	baseClient
	baseCmdable
	cmdable
	ingestCmdable
	controlCmdable
	statefulCmdable
}

// Conn represents a single Sonic connection rather than a pool of connections.
// Prefer running commands from Client unless there is a specific need
// for a continuous single Sonic connection.
type Conn struct {
	*conn
	ctx context.Context
}

func newConn(ctx context.Context, opt *Options, connPool pool.Pooler) *Conn {
	c := Conn{
		conn: &conn{
			baseClient: baseClient{
				opt:      opt,
				connPool: connPool,
			},
		},
		ctx: ctx,
	}
	if opt.ChannelMode == ChannelSearch {
		c.cmdable = c.Process
	} else if opt.ChannelMode == ChannelIngest {
		c.ingestCmdable = c.Process
	} else if opt.ChannelMode == ChannelControl {
		c.controlCmdable = c.Process
	}

	c.baseCmdable = c.Process
	c.statefulCmdable = c.Process
	return &c
}

func (c *Conn) Process(ctx context.Context, cmd Cmder) error {
	retErr := c.baseClient.process(ctx, cmd)
	cmd.SetErr(retErr)
	return retErr
}
