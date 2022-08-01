package sonic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/uretgec/go-sonic/proto"
)

type Cmder interface {
	Name() string
	FullName() string
	Args() []string
	String() string

	readTimeout() *time.Duration
	readReply(rd *proto.Reader) error

	SetErr(error)
	Err() error
}

func writeCmd(wr *proto.Writer, cmd Cmder) error {
	return wr.WriteArgs(cmd.Args())
}

func cmdString(cmd Cmder, val interface{}) string {
	if err := cmd.Err(); err != nil {
		return err.Error()
	} else if val != nil {
		switch v := val.(type) {
		case string:
			return v
		case []byte:
			return string(v)
		case int:
			return strconv.Itoa(v)
		}
	}

	return ""
}

//------------------------------------------------------------------------------

type baseCmd struct {
	ctx  context.Context
	args []string
	err  error

	_readTimeout *time.Duration
}

var _ Cmder = (*Cmd)(nil)

func (cmd *baseCmd) Name() string {
	if len(cmd.args) == 0 {
		return ""
	}
	return cmd.stringArg(0)
}

func (cmd *baseCmd) FullName() string {
	return cmd.Name()
}

func (cmd *baseCmd) Args() []string {
	return cmd.args
}

func (cmd *baseCmd) stringArg(pos int) string {
	if pos < 0 || pos >= len(cmd.args) {
		return ""
	}
	return cmd.args[pos]
}

func (cmd *baseCmd) SetErr(e error) {
	cmd.err = e
}

func (cmd *baseCmd) Err() error {
	return cmd.err
}

func (cmd *baseCmd) readTimeout() *time.Duration {
	return cmd._readTimeout
}

//------------------------------------------------------------------------------

type Cmd struct {
	baseCmd

	val interface{}
}

func NewCmd(ctx context.Context, args ...string) *Cmd {
	return &Cmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *Cmd) String() string {
	return cmdString(cmd, cmd.val)
}

func (cmd *Cmd) SetVal(val interface{}) {
	cmd.val = val
}

func (cmd *Cmd) Val() interface{} {
	return cmd.val
}

func (cmd *Cmd) Result() (interface{}, error) {
	return cmd.val, cmd.err
}

func (cmd *Cmd) Text() (string, error) {
	if cmd.err != nil {
		return "", cmd.err
	}
	return toString(cmd.val)
}

func toString(val interface{}) (string, error) {
	switch val := val.(type) {
	case string:
		return val, nil
	default:
		err := fmt.Errorf("sonic: unexpected type=%T for String", val)
		return "", err
	}
}

func (cmd *Cmd) Int() (int, error) {
	if cmd.err != nil {
		return 0, cmd.err
	}
	switch val := cmd.val.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	default:
		err := fmt.Errorf("sonic: unexpected type=%T for Int", val)
		return 0, err
	}
}

func (cmd *Cmd) Bool() (bool, error) {
	if cmd.err != nil {
		return false, cmd.err
	}
	return toBool(cmd.val)
}

func toBool(val interface{}) (bool, error) {
	switch val := val.(type) {
	case int64:
		return val != 0, nil
	case string:
		return strconv.ParseBool(val)
	default:
		err := fmt.Errorf("sonic: unexpected type=%T for Bool", val)
		return false, err
	}
}

func (cmd *Cmd) Slice() ([]string, error) {
	if cmd.err != nil {
		return nil, cmd.err
	}
	switch val := cmd.val.(type) {
	case []string:
		return val, nil
	default:
		return nil, fmt.Errorf("sonic: unexpected type=%T for Slice", val)
	}
}

func (cmd *Cmd) StringSlice() ([]string, error) {
	slice, err := cmd.Slice()
	if err != nil {
		return nil, err
	}

	ss := make([]string, len(slice))
	for i, iface := range slice {
		val, err := toString(iface)
		if err != nil {
			return nil, err
		}
		ss[i] = val
	}
	return ss, nil
}

func (cmd *Cmd) BoolSlice() ([]bool, error) {
	slice, err := cmd.Slice()
	if err != nil {
		return nil, err
	}

	bools := make([]bool, len(slice))
	for i, iface := range slice {
		val, err := toBool(iface)
		if err != nil {
			return nil, err
		}
		bools[i] = val
	}
	return bools, nil
}

func (cmd *Cmd) readReply(rd *proto.Reader) (err error) {
	fmt.Printf("BASE Komut: %s - CMD: %v\n", cmd.args[0], cmd.args)
	cmd.val, err = rd.ReadReply(sliceParser, "")
	return err
}

// sliceParser implements proto.MultiBulkParse.
func sliceParser(rd *proto.Reader, n int64) (interface{}, error) {
	vals := make([]interface{}, n)
	for i := 0; i < len(vals); i++ {
		v, err := rd.ReadReply(sliceParser, "")
		if err != nil {
			if err, ok := err.(proto.SonicError); ok {
				vals[i] = err
				continue
			}
			return nil, err
		}
		vals[i] = v
	}
	return vals, nil
}

//------------------------------------------------------------------------------

type IntCmd struct {
	baseCmd

	val int64
}

var _ Cmder = (*IntCmd)(nil)

func NewIntCmd(ctx context.Context, args ...string) *IntCmd {
	return &IntCmd{
		baseCmd: baseCmd{
			ctx:  ctx,
			args: args,
		},
	}
}

func (cmd *IntCmd) SetVal(val int64) {
	cmd.val = val
}

func (cmd *IntCmd) Val() int64 {
	return cmd.val
}

func (cmd *IntCmd) Result() (int64, error) {
	return cmd.val, cmd.err
}

func (cmd *IntCmd) String() string {
	return cmdString(cmd, cmd.val)
}

func (cmd *IntCmd) readReply(rd *proto.Reader) (err error) {
	fmt.Printf("INT Komut: %s - CMD: %v\n", cmd.args[0], cmd.args)
	cmd.val, err = rd.ReadIntReply()
	return err
}
