package cahce

import (
	"time"
	"mnet/common"
)

/**
	内存存储器
 */

var Message_Global = New(10 * time.Minute, 10 * time.Second)


func SelectMemory(t common.CacheType) *Cache{
	switch t {
	case common.Message_Global:
		return Message_Global
	default:
		return Message_Global
	}
}

