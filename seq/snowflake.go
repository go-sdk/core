package seq

import (
	"strconv"
	"time"

	"github.com/sony/sonyflake/v2"
)

var sf *sonyflake.Sonyflake

func init() {
	sf, _ = sonyflake.New(sonyflake.Settings{
		StartTime: time.Date(2023, 4, 5, 6, 7, 8, 9, time.UTC),
		MachineID: func() (int, error) { return 56565, nil },
	})
}

// NextID 返回十进制字符串形式的下一个 Snowflake ID。
// 固定机器编号的唯一性范围为同一时间仅运行一个生成器实例。
// 生成器耗尽可用时间范围时会触发 panic。
func NextID() string {
	id, err := sf.NextID()
	if err != nil {
		panic(err)
	}
	return strconv.FormatInt(id, 10)
}
