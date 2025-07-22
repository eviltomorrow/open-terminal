package snowflake

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/snowflake"
)

var (
	machineID     int64 = 1
	node          *snowflake.Node
	limiterLength = 19
)

func init() {
	n, err := snowflake.NewNode(machineID)
	if err != nil {
		panic(fmt.Errorf("snowflake NewNode failure, nest error: %v", err))
	}
	node = n
}

func GenerateID() string {
	var (
		id     = node.Generate()
		result = id.String()
	)
	switch {
	case len(result) < limiterLength:
		var (
			n   = limiterLength - len(result)
			buf strings.Builder
		)
		for i := 0; i < n; i++ {
			buf.WriteString("0")
		}
		buf.WriteString(result)
		return buf.String()
	default:
		return result
	}
}
