package sonic

import (
	"context"
	"io"
	"net"
	"strings"

	"github.com/uretgec/go-sonic/pool"
	"github.com/uretgec/go-sonic/proto"
)

// ErrClosed performs any operation on the closed client will return this error.
var ErrClosed = pool.ErrClosed

type Error interface {
	error

	// SonicError is a no-op function but
	// serves to distinguish types that are Sonic
	// errors from ordinary errors: a type is a
	// Sonic error if it has a SonicError method.
	SonicError()
}

var _ Error = proto.SonicError("")

func shouldRetry(err error, retryTimeout bool) bool {
	switch err {
	case io.EOF, io.ErrUnexpectedEOF:
		return true
	case nil, context.Canceled, context.DeadlineExceeded:
		return false
	}

	if v, ok := err.(timeoutError); ok {
		if v.Timeout() {
			return retryTimeout
		}
		return true
	}

	s := err.Error()
	if s == "ERR max number of clients reached" {
		return true
	}
	if strings.HasPrefix(s, "ERR ") {
		return true
	}

	return false
}

func isSonicError(err error) bool {
	_, ok := err.(proto.SonicError)
	return ok
}

func isBadConn(err error, allowTimeout bool, addr string) bool {
	switch err {
	case nil:
		return false
	case context.Canceled, context.DeadlineExceeded:
		return true
	}

	if isSonicError(err) {
		return false
	}

	if allowTimeout {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return !netErr.Temporary()
		}
	}

	return true
}

type timeoutError interface {
	Timeout() bool
}
