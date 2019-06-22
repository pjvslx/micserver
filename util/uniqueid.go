package util

import (
	"errors"
	"sync"
	"time"
)

type uniqueIDBuilder struct {
	lastLowlevelID uint16
	lastMidlevelID uint32
	// lastHeightlevelID uint16

	lastMidLowLevelID uint64
	mutex             sync.Mutex
}

var s_unique uniqueIDBuilder

func (this *uniqueIDBuilder) NewUniqueID(heightlevelID uint16) (uint64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	nowtime := time.Now().Unix()
	if nowtime < 1514736000 {
		return 0,
			errors.New("Server time error!!! Must late than 2018/1/1 00:00:00")
	}
	now := uint32(nowtime & 0x0ffffffff)
	if now > this.lastMidlevelID {
		// 时间已过去最后的兼容秒数
		this.lastLowlevelID = 0
		this.lastMidlevelID = now
	} else {
		if this.lastLowlevelID == 0x0ffff {
			// 生成频率超限
			return 0, errors.New("本秒内已随机出超出限制的唯一ID数量")
		} else {
			this.lastLowlevelID++
		}
	}

	subvalue := uint64(0)
	subvalue |= uint64(this.lastMidlevelID) << (16)
	subvalue |= uint64(this.lastLowlevelID) << (0)
	if subvalue <= this.lastMidLowLevelID {
		return 0, errors.New("生成的ID可能重复了")
	}
	this.lastMidLowLevelID = subvalue

	res := uint64(0)
	res |= uint64(heightlevelID) << (16 + 32)
	res |= uint64(subvalue) << (0)
	return res, nil
}

func NewUniqueID(heightlevelID uint16) (uint64, error) {
	return s_unique.NewUniqueID(heightlevelID)
}
