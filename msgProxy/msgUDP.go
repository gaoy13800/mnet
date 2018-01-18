package msgProxy

import (
	"fmt"
	"log"
	cache "mnet/cahce"
	"mnet/dbhelp"
	"mnet/event"
	"mnet/push"
	"net"
	"strconv"
	"strings"
	"time"
)

type IMsgU interface {
	SendUdpMsg(guid string) error

	Deal(data *push.UdpEvent)
}

type udpMsgProxy struct {
	Data *push.UdpEvent

	common *msgHandler

	Caches *cache.Cache
}

//信息处理
func (this *udpMsgProxy) Notice() {
	length := len(this.Data.Msg)

	var action, status, guid string

	var isTrue bool

	if length > 8 && length == 26 {

		action, status, guid, isTrue = this.spliceStr(this.Data.Msg)

	} else {

		isTrue, status, action = true, "0", this.Data.Msg

		port := strconv.Itoa(this.Data.RemoteAddr.Port)

		id, err := this.Caches.Get(this.Data.RemoteAddr.IP.String() + port)

		if !err {
			log.Println("没有心跳包数据！！！hblv")
			return
		}
		guid = id.(string)
	}

	if !isTrue {
		log.Println("无法解析此条消息")
		return
	}

	switch action {
	case "wthblv":
		//心跳包
		this.heblv("wtokabc", guid, status)
		break
	case "wtopen1":
		this.handle("wtopen1", guid)
		break
	case "wtopen3":
		this.handle("wtopen3", guid)
		break
		//case "wtopen5":
		//	this.handle("wtopen5", guid)
		//	break
	case "wtclse1":
		this.handle("wtclse1", guid)
		break
	case "wtclse3":
		this.handle("wtclse3", guid)
		break
		//case "wtclse5":
		//	this.handle("wtclse5", guid)
		//	break
	case "wtbrut1":
		this.brut("wtbrut2", guid)
		break
	default:
		log.Println("未能识别动作", action)
		break
	}
}

//向指定ip发送消息
func (this *udpMsgProxy) SendUdpMsg(cmd string, addr *net.UDPAddr) error {
	// 发送数据
	_, err := this.Data.Socket.WriteToUDP([]byte(cmd), addr)
	log.Println("发送数据：", addr, cmd)
	if err != nil {
		fmt.Println("发送数据失败!", err)
		return err
	}
	return err
}

//心跳包，保存信息
func (this *udpMsgProxy) heblv(cmd, guid, status string) {
	//保存终端信息到redis guid addr
	this.common.iRedis.Save_Udp(guid, status, this.Data.RemoteAddr)

	this.Caches.Set("hblvxintiao"+guid, guid, time.Second*600)

	//获取心跳缓存
	key := "wthblv" + guid
	value := this.GetCache(key)

	if value == nil || value == 0 {
		this.Caches.Set(key, 1, time.Hour*24*30*12)
	} else if value == 1 {
		this.Caches.Set(key, 2, time.Hour*24*30*12)
	} else if value == 2 {
		this.Caches.Set(key, 0, time.Hour*24*30*12)
		//给终端发送信息
		this.SendUdpMsg(cmd, this.Data.RemoteAddr)
	}

	port := strconv.Itoa(this.Data.RemoteAddr.Port)

	this.Caches.Set(this.Data.RemoteAddr.IP.String()+port, guid, time.Hour*24*30*12)
}

//open、close操作
func (this *udpMsgProxy) handle(cmd, guid string) {

	//设置缓存初始值
	this.SetInit(guid)
	log.Println("cmd", cmd)
	//保存地锁状态
	if cmd == "wtopen1" || cmd == "wtopen3" {
		this.common.SaveCstu(guid, event.LOCK_OPEN)
	} else if cmd == "wtclse1" || cmd == "wtclse3" {
		this.common.SaveCstu(guid, event.LOCK_CLOSE)
	}

	//if cmd == "wtopen5" || cmd == "wtclse5"{
	//	//保存地锁状态
	//	if  cmd == "wtopen5" {
	//		this.common.SaveCstu(guid, event.LOCK_OPEN)
	//	}else {
	//		this.common.SaveCstu(guid, event.LOCK_CLOSE)
	//	}
	//	//设置缓存初始值
	//	this.SetInit(guid)
	//}else {
	//	//存入缓存
	//	//log.Println("收到", guid,cmd,this.GetCacheTimes(guid,cmd), event.CACHE_DEADLINE)
	//	this.SetCache(guid, cmd, this.GetCacheTimes(guid,cmd), event.CACHE_DEADLINE)
	//	this.SendUdpMsg(cmd, this.Data.RemoteAddr)
	//}
}

//brut操作
func (this *udpMsgProxy) brut(cmd, guid string) {
	this.SetInit(guid)
	//给终端发送信息
	//this.SendUdpMsg(cmd , this.Data.RemoteAddr)
}

func (this *udpMsgProxy) spliceStr(data string) (string, string, string, bool) {
	cmd, status, guid := data[:6], data[6:7], data[7:]
	return cmd, status, guid, true
}

//设置缓存初始值
func (this *udpMsgProxy) SetInit(guid string) {
	fmt.Println("guid", guid)

	this.SetCache(guid, "ok", "0", 24)
}

//设置缓存
func (this *udpMsgProxy) SetCache(key, cmd, times string, t int) {
	//存入缓存
	var timed time.Duration

	if t == 4 {
		timed = time.Second * 4
	} else {
		timed = time.Hour * 24
	}

	value := key + "|" + cmd + "|" + times

	this.Caches.Set(key, value, timed)

}

//获取缓存
func (this *udpMsgProxy) GetCache(key string) interface{} {
	if foo, found := this.Caches.Get(key); found {
		return foo
	} else {
		return foo
	}
}

//获取缓存次数
func (this *udpMsgProxy) GetCacheTimes(key, cmd string) string {

	if foo, found := this.Caches.Get(key); found {
		sp := strings.Split(foo.(string), "|")
		if sp[1] == cmd {
			return string(sp[2])
		}
		return "1"
	} else {
		return "1"
	}
}

func NewUdpMsgProxy(data *push.UdpEvent, redis dbhelp.IRedis, caches *cache.Cache) *udpMsgProxy {
	this := &udpMsgProxy{
		Data:   data,
		Caches: caches,
	}

	this.common = NewMsgHandler(redis, this.Caches)

	return this
}
