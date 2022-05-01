package sonic

import (
	"context"
	"github.com/uretgec/go-sonic/pool"
	"net"
	"runtime"
	"time"
)

// Options keeps the settings to setup sonic connection.
type Options struct {
	// host:port address.
	Addr string

	// Dialer creates new network connection and has priority over
	// Network and Addr options.
	Dialer func(ctx context.Context, network, addr string) (net.Conn, error)

	// Hook that is called when new connection is established.
	OnConnect func(ctx context.Context, cn *Conn) error

	//  Secret Password for client with server
	AuthPassword string

	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration

	// Type of connection pool.
	// true for FIFO pool, false for LIFO pool.
	// Note that fifo has higher overhead compared to lifo.
	PoolFIFO bool
	// Maximum number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	PoolSize int
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	// Look at: sonic.cfg[tcp_timeout]
	// Default Max 300
	MaxConnAge time.Duration
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration

	// Enables read only queries on slave nodes.
	readOnly bool

	// Channel Mode
	// Select one of them - search, control, ingest
	// Default: search
	ChannelMode string

	// Max Buffered Size
	// Default: 20000
	// Real buffer size return from first connect sonic server response
	// STARTED search protocol(1) buffer(20000)
	MaxBufferedSize int
}

func (opt *Options) init() {
	if opt.Addr == "" {
		opt.Addr = "localhost:1491"
	}
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 5 * time.Second
	}
	if opt.Dialer == nil {
		opt.Dialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   opt.DialTimeout,
				KeepAlive: 2 * time.Minute,
			}
			return netDialer.DialContext(ctx, network, addr)
		}
	}
	if opt.PoolSize == 0 {
		opt.PoolSize = 10 * runtime.GOMAXPROCS(0)
	}
	switch opt.ReadTimeout {
	case -1:
		opt.ReadTimeout = 0
	case 0:
		opt.ReadTimeout = 3 * time.Second
	}
	switch opt.WriteTimeout {
	case -1:
		opt.WriteTimeout = 0
	case 0:
		opt.WriteTimeout = opt.ReadTimeout
	}
	if opt.PoolTimeout == 0 {
		opt.PoolTimeout = opt.ReadTimeout + time.Second
	}
	if opt.MaxConnAge == 0 {
		opt.IdleTimeout = 2 * time.Minute
	}
	if opt.IdleTimeout == 0 {
		opt.IdleTimeout = 2 * time.Minute
	}
	if opt.IdleCheckFrequency == 0 {
		opt.IdleCheckFrequency = time.Minute
	}

	if opt.MaxRetries == -1 {
		opt.MaxRetries = 0
	} else if opt.MaxRetries == 0 {
		opt.MaxRetries = 3
	}
	switch opt.MinRetryBackoff {
	case -1:
		opt.MinRetryBackoff = 0
	case 0:
		opt.MinRetryBackoff = 8 * time.Millisecond
	}
	switch opt.MaxRetryBackoff {
	case -1:
		opt.MaxRetryBackoff = 0
	case 0:
		opt.MaxRetryBackoff = 512 * time.Millisecond
	}

	// Channel Mode
	if opt.ChannelMode == "" {
		opt.ChannelMode = ChannelSearch
	}

	// Max BuffereSize
	if opt.MaxBufferedSize == 0 {
		opt.MaxBufferedSize = 20000
	}

	/*opt.OnConnect = func(ctx context.Context, cn *Conn) error {
		// Connect Sonic Server First Time
		bufferSize, err := cn.Start(ctx, opt.ChannelMode, opt.AuthPassword).Int()
		if err != nil {
			fmt.Printf("BURAYA GELDİ ERR: %v\n", err.Error())
			return err
		}

		// Set MaxBufferedSize
		cn.opt.MaxBufferedSize = bufferSize

		return nil
	}*/
}

func (opt *Options) clone() *Options {
	clone := *opt
	return &clone
}

func newConnPool(opt *Options) *pool.ConnPool {
	return pool.NewConnPool(&pool.Options{
		Dialer: func(ctx context.Context) (net.Conn, error) {
			return opt.Dialer(ctx, "tcp", opt.Addr)
		},
		PoolFIFO:           opt.PoolFIFO,
		PoolSize:           opt.PoolSize,
		MinIdleConns:       opt.MinIdleConns,
		MaxConnAge:         opt.MaxConnAge,
		PoolTimeout:        opt.PoolTimeout,
		IdleTimeout:        opt.IdleTimeout,
		IdleCheckFrequency: opt.IdleCheckFrequency,
		/*OnClose: func(c *pool.Conn) error {
			fmt.Printf("Quit: %s\n", c.RemoteAddr().String())
			return nil
		},*/
	})
}
