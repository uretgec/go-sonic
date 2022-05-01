package sonic

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrQueryBuilder = errors.New("struct not correct")

type QueryBuilder struct {
	Mode     string `json:"mode,omitempty" redis:"mode"`
	Password string `json:"password,omitempty" redis:"password"`

	Command    string `json:"command,omitempty" redis:"command"`
	Collection string `json:"collection,omitempty" redis:"collection"`
	Bucket     string `json:"bucket,omitempty" redis:"bucket"`
	Object     string `json:"object,omitempty" redis:"object"`
	Text       string `json:"text,omitempty" redis:"text"`
	Lang       string `json:"lang,omitempty" redis:"lang"`
	Limit      int    `json:"limit,omitempty" redis:"limit"`
	Offset     int    `json:"offset,omitempty" redis:"offset"`

	Action string `json:"action,omitempty" redis:"action"`
	Data   string `json:"data,omitempty" redis:"data"`
}

func NewQueryBuilder() QueryBuilder {
	return QueryBuilder{}
}

func (sb QueryBuilder) MarshalBinary() ([]byte, error) {
	return json.Marshal(sb)
}

func (sb QueryBuilder) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &sb); err != nil {
		return err
	}

	return nil
}

func (sb QueryBuilder) FindSuggestWord(words string) int {
	return strings.Index(words, " ")
}

func (qb QueryBuilder) Encode() []string {
	var args []string
	if qb.Command != "" {
		args = append(args, qb.Command)
	}

	if contains([]string{CmdSearchQuery, CmdSearchSuggest, CmdIngestPush, CmdIngestPop, CmdIngestCount, CmdIngestFlushc, CmdIngestFlushb, CmdIngestFlusho}, qb.Command) && qb.Collection != "" {
		args = append(args, qb.Collection)

		if qb.Bucket != "" {
			args = append(args, qb.Bucket)

			if qb.Object != "" {
				args = append(args, qb.Object)
			}

			if contains([]string{CmdSearchQuery, CmdSearchSuggest, CmdIngestPush, CmdIngestPop}, qb.Command) && qb.Text != "" {
				args = append(args, fmt.Sprintf("\"%s\"", qb.Text))

				if contains([]string{CmdSearchQuery, CmdSearchSuggest}, qb.Command) {
					if qb.Limit != -1 {
						args = append(args, "LIMIT("+strconv.Itoa(qb.Limit)+")")
					}

					if qb.Command == CmdSearchQuery && qb.Offset != -1 {
						args = append(args, "OFFSET("+strconv.Itoa(qb.Offset)+")")
					}
				}

				if contains([]string{CmdSearchQuery, CmdIngestPush}, qb.Command) && qb.Lang != "" {
					args = append(args, "LANG("+qb.Lang+")")
				}
			}
		}
	} else if CmdControlTrigger == qb.Command {
		if qb.Action != "" {
			args = append(args, qb.Action)
		}

		if qb.Data != "" {
			args = append(args, qb.Data)
		}
	} else if CmdSearchStart == qb.Command {
		if qb.Mode != "" {
			args = append(args, qb.Mode)
		}

		if qb.Password != "" {
			args = append(args, qb.Password)
		}
	}

	return args
}

// TODO
func (qb QueryBuilder) Decode(query string) QueryBuilder {
	return qb
}
