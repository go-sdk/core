package seq

import (
	"github.com/google/uuid"
)

func init() {
	uuid.EnableRandPool()
}

// UUID 返回随机生成的 UUID v7 字符串。
func UUID() string {
	return uuid.Must(uuid.NewV7()).String()
}

// UUIDShort 返回移除连字符后的 UUID v7 字符串。
func UUIDShort() string {
	s := UUID()
	return s[:8] + s[9:13] + s[14:18] + s[19:23] + s[24:]
}
