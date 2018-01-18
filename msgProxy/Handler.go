package msgProxy

import (
	"mnet/dbhelp"
	"mnet/event"
	cache "mnet/cahce"
)

type msgHandler struct {
	iRedis dbhelp.IRedis

	caches *cache.Cache
}

//保存状态
func (this *msgHandler) SaveCstu(guid ,cstu string)  {
	this.iRedis.Save_(event.SAVE_CSTU, guid, cstu)
}

//保存电量
func (this *msgHandler) SaveStus(guid ,cstu string)  {
	this.iRedis.Save_(event.SAVE_STUS, guid, cstu)
}

func NewMsgHandler(redis dbhelp.IRedis,caches *cache.Cache) *msgHandler  {
	this := &msgHandler{
		iRedis:redis,
		caches:caches,
	}
	return this
}