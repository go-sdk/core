package seq

import (
	"strconv"
	"time"

	"github.com/sony/sonyflake/v2"

	"github.com/go-sdk/core/osx"
)

var sf *sonyflake.Sonyflake

func init() {
	year := osx.GetEnv[int](2020, "SONYFLAKE_START_YEAR")
	machineID := osx.GetEnv[int](56565, "SONYFLAKE_MACHINE_ID")
	var err error
	sf, err = sonyflake.New(sonyflake.Settings{
		StartTime: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID: func() (int, error) { return machineID, nil },
	})
	if err != nil {
		osx.Panicf("seq: sonyflake init failed, check SONYFLAKE_START_YEAR(%d) and SONYFLAKE_MACHINE_ID(%d): %v", year, machineID, err)
	}
}

// NextID 返回十进制字符串形式的下一个 Snowflake ID。
// 机器编号默认 56565，可通过 SONYFLAKE_MACHINE_ID 覆盖，唯一性范围为同一时间仅运行一个生成器实例。
// 生成器耗尽可用时间范围时会触发 panic。
func NextID() string {
	id, err := sf.NextID()
	if err != nil {
		osx.Panic(err)
	}
	return strconv.FormatInt(id, 10)
}
