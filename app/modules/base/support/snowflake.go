package support

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"eqray-flavor/app/modules/base/types"
)

// 雪花 ID 位宽：1 符号位 + 41 毫秒时间戳 + 10 节点 + 12 序列。
const (
	epochMs      int64 = 1704067200000 // 2024-01-01 00:00:00 UTC
	nodeBits     uint8 = 10
	sequenceBits uint8 = 12

	nodeMax     int64 = -1 ^ (-1 << nodeBits)
	sequenceMax int64 = -1 ^ (-1 << sequenceBits)
	shiftBits   uint8 = nodeBits + sequenceBits

	// maxBackwardMs 是允许自旋等待的时钟回拨上限。
	maxBackwardMs int64 = 5000
)

type snowflake struct {
	mu       sync.Mutex
	node     int64
	lastMs   int64
	sequence int64
}

var (
	idOnce sync.Once
	idGen  *snowflake
)

// NewID 生成 64 位雪花 ID。
// 节点号优先取环境变量 SNOWFLAKE_NODE_ID（0 ~ nodeMax），未配置时回退到进程号；
// 多实例部署必须显式配置，否则进程号相同时会生成重复 ID。
// 时钟回拨在 maxBackwardMs 内自旋等待，超过阈值直接 panic，宁可失败也不生成重复 ID。
func NewID() types.ID {
	idOnce.Do(func() {
		idGen = &snowflake{node: nodeID()}
	})

	idGen.mu.Lock()
	defer idGen.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < idGen.lastMs {
		if back := idGen.lastMs - now; back > maxBackwardMs {
			panic(fmt.Sprintf("snowflake: 时钟回拨 %dms，超过阈值 %dms", back, maxBackwardMs))
		}
		for now < idGen.lastMs {
			time.Sleep(time.Millisecond)
			now = time.Now().UnixMilli()
		}
	}

	if now == idGen.lastMs {
		idGen.sequence = (idGen.sequence + 1) & sequenceMax
		if idGen.sequence == 0 {
			// 当前毫秒序列耗尽，等到下一毫秒
			for now <= idGen.lastMs {
				time.Sleep(time.Millisecond)
				now = time.Now().UnixMilli()
			}
		}
	} else {
		idGen.sequence = 0
	}
	idGen.lastMs = now

	return types.ID((now-epochMs)<<shiftBits | idGen.node<<sequenceBits | idGen.sequence)
}

// nodeID 解析节点号：环境变量优先，未配置或非法时回退到进程号并截断到节点位宽。
func nodeID() int64 {
	if raw := os.Getenv("SNOWFLAKE_NODE_ID"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil && v >= 0 && v <= nodeMax {
			return v
		}
	}

	return int64(os.Getpid()) & nodeMax
}
