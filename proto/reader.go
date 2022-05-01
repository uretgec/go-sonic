package proto

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// sonic resp protocol data type.
const (
	ErrorReply     = "ERR"
	ConnectedReply = "CONNECTED"
	StartedReply   = "STARTED"
	PongReply      = "PONG"
	PendingReply   = "PENDING"
	EventReply     = "EVENT"
	QueryReply     = "QUERY"
	ResultReply    = "RESULT"
	OkReply        = "OK"
	EndedReply     = "ENDED"
)

//------------------------------------------------------------------------------

type SonicError string

func (e SonicError) Error() string { return string(e) }

func (SonicError) SonicError() {}

//------------------------------------------------------------------------------

type MultiBulkParse func(*Reader, int64) (interface{}, error)

type Reader struct {
	rd   *bufio.Reader
	_buf []byte
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{
		rd:   bufio.NewReader(rd),
		_buf: make([]byte, 64),
	}
}

func (r *Reader) Buffered() int {
	return r.rd.Buffered()
}

func (r *Reader) Peek(n int) ([]byte, error) {
	return r.rd.Peek(n)
}

func (r *Reader) Reset(rd io.Reader) {
	r.rd.Reset(rd)
}

func (r *Reader) ReadLine() ([]byte, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}

	return line, nil
}

// readLine that returns an error if:
//   - there is a pending read error;
//   - or line does not end with \r\n.
func (r *Reader) readLine() ([]byte, error) {
	b, err := r.rd.ReadSlice('\n')
	if err != nil {
		if err != bufio.ErrBufferFull {
			return nil, err
		}

		full := make([]byte, len(b))
		copy(full, b)

		b, err = r.rd.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		full = append(full, b...) //nolint:makezero
		b = full
	}
	if len(b) <= 2 || b[len(b)-1] != '\n' || b[len(b)-2] != '\r' {
		return nil, fmt.Errorf("sonic: invalid reply: %q", b)
	}

	fmt.Printf("Readline: %s\n", string(b[:len(b)-2]))
	return b[:len(b)-2], nil
}

func (r *Reader) ReadReply(m MultiBulkParse, marker string) (interface{}, error) {
	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}

	w := bufio.NewScanner(bytes.NewBuffer(line))
	w.Split(bufio.ScanWords)
	w.Scan()

	fmt.Printf("BASE Find: %s\n", w.Text())
	switch w.Text() {
	case ErrorReply:
		return nil, SonicError(string(line[:]))
	case ConnectedReply:
		return r.ReadReply(m, "") // Read Next Line
	case StartedReply:
		// Get Buffer Size
		// Via: https://github.com/expectedsh/go-sonic/blob/master/sonic/connection.go
		ss := strings.FieldsFunc(string(line), func(r rune) bool {
			if unicode.IsSpace(r) || r == '(' || r == ')' {
				return true
			}
			return false
		})

		fmt.Printf("StartedReply Buffer: %s\n", ss[len(ss)-1])
		return strconv.ParseInt(ss[len(ss)-1], 10, 64)
	case PongReply:
		return string(line), nil
	case EventReply:
		if marker != "" && strings.HasPrefix(string(line), "EVENT QUERY "+marker) {
			// Make Slice
			results := []string{}
			result := strings.Replace(string(line), "EVENT QUERY "+marker+" ", "", 1)
			if len(result) == 0 {
				return results, nil
			}

			results = strings.Split(result, " ")
			if len(results) > 0 {
				return results, nil
			}
		} else if marker != "" && strings.HasPrefix(string(line), "EVENT SUGGEST "+marker) {
			// Make Slice
			results := []string{}
			result := strings.Replace(string(line), "EVENT SUGGEST "+marker+" ", "", 1)
			if len(result) == 0 {
				return results, nil
			}

			results = strings.Split(result, " ")
			if len(results) > 0 {
				return results, nil
			}
		}

		return nil, fmt.Errorf("sonic: conn ended marker %s not found", marker)
	case PendingReply:
		// Find marker for follow search result
		marker = strings.Replace(string(line), "PENDING ", "", 1)
		return r.ReadReply(m, marker) // Read Next Line
	case OkReply:
		return int64(1), nil
	case ResultReply:
		// Make Slice
		results := strings.Split(string(line), " ")
		if len(results) > 0 {
			return results[1:], nil
		}

		return nil, fmt.Errorf("sonic: conn ended")
	case EndedReply:
		return nil, nil
	}

	return nil, fmt.Errorf("sonic: can't parse %.100q", line)
}

func (r *Reader) ReadIntReply() (int64, error) {
	line, err := r.ReadLine()
	if err != nil {
		return 0, err
	}
	w := bufio.NewScanner(bytes.NewBuffer(line))
	w.Split(bufio.ScanWords)
	w.Scan()

	fmt.Printf("INT Find: %s\n", w.Text())
	switch w.Text() {
	case ErrorReply:
		return 0, SonicError(string(line[:]))
	case ResultReply:
		results := strings.Split(string(line), " ")
		return strconv.ParseInt(results[1], 10, 64)
	case EndedReply:
		return 0, nil
	default:
		return 0, fmt.Errorf("redis: can't parse int reply: %.100q", line)
	}
}
