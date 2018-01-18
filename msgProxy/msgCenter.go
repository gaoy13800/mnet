package msgProxy

import (
	"log"
	"mnet/event"
	"mnet/dbhelp"
	"mnet/push"
	"net"
	"strings"
	cache "mnet/cahce"
	"strconv"
	"time"
	"fmt"
	"sync"
)

type distributeFocus struct {
	iRedis dbhelp.IRedis
	caches *cache.Cache
	udpDeal *udpMsgProxy
	udpSocket *net.UDPConn
	syncTex sync.RWMutex
}

func (this *distributeFocus) MsgFocus(eventN interface{}, t event.Proto)  {

	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	switch t {
	case event.Proto_UDP:

		data := eventN.(*push.UdpEvent)

		this.udpSocket = data.Socket

		NewUdpMsgProxy(data, this.iRedis, this.caches).Notice()

		break
	default:
		log.Println("未能识别类型", t)
		break
	}
}

func (this *distributeFocus) FocusSub(data string, conn *net.UDPConn) {
	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	list := strings.Split(data, "|")

	if len(list) == 2 {

		deviceId, cmd := list[0],list[1]

		cahceStr := "publishcache" + data

		result, _ := this.caches.Get(cahceStr)

		if result != nil {
			return
		}

		this.caches.Set(cahceStr, data, time.Second * 10)

		address := this.iRedis.IsExistGuId(deviceId)

		if address == "" {
			log.Println("FocusSub from deviceId get address is null")
			return
		}

		ipstring := strings.Split(address, ":")
		port, _ := strconv.Atoi(ipstring[1])
		addr := net.UDPAddr{
			IP:net.ParseIP(ipstring[0]),
			Port:port,
		}

		//realAddress,_  :=  net.ResolveUDPAddr("udp4", address)

		key := "wt" + cmd + "0"

        log.Printf("sub business %s  and send device %s --", data, key)

		_, err := conn.WriteToUDP([]byte(key), &addr)
		log.Println("发送数据：", &addr,key)
		if err != nil{
			log.Println(err)
			return
		}

		value := deviceId + "|" + key + "|" + "0"
		this.caches.Set(deviceId, value,time.Second * event.CACHE_DEADLINE)

	}
}
//缓存回调
func (this *distributeFocus) CallBack(k string, v interface{}){
	log.Println("in callback ", k, v)

	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	if strings.Contains(k, "publishcache") {
		return
	}
	//心跳过期
	if strings.Contains(k,"hblvxintiao") {
		this.iRedis.SetClientUdp(v.(string),"2")
		return
	}

	sp := strings.Split(v.(string),"|")

	guid ,cmd ,times:= sp[0],sp[1],sp[2]

		log.Println("send times ----",times)

		if t,_ := strconv.Atoi(times);t < 10 {

			value := guid + "|" + cmd + "|" + strconv.Itoa(t+1)

		address := this.iRedis.IsExistGuId(guid)

		ipstring := strings.Split(address, ":")
		port, _ := strconv.Atoi(ipstring[1])
		addr := net.UDPAddr{
			IP:net.ParseIP(ipstring[0]),
			Port:port,
		}

		//realAddress, _  :=  net.ResolveUDPAddr("udp4", addr)

		_, err := this.udpSocket.WriteToUDP([]byte(cmd), &addr)
		log.Println("发送数据：", &addr,cmd)

		log.Printf("send address %s send content %s", addr, cmd)

		if err != nil {
			log.Println("callback send error:", err)
			return
		}

		this.caches.Set(guid, value, time.Second * event.CACHE_DEADLINE)
	}
}

func (this *distributeFocus) VipFocus(data string, conn *net.UDPConn){

	this.syncTex.Lock()
	defer this.syncTex.Unlock()

	if len(data) != 23 {
		log.Println("vip send data wrong")
		return
	}

	deviceId := data[:20]

	tmpAction := data[20:]

	addr := this.iRedis.IsExistGuId(deviceId)


	ipstring := strings.Split(addr, ":")
	port, _ := strconv.Atoi(ipstring[1])
	address := net.UDPAddr{
		IP:net.ParseIP(ipstring[0]),
		Port:port,
	}

	log.Println("收到vip消息:", address,data)

	index, err := conn.WriteToUDP([]byte(tmpAction), &address)
	log.Println("发送数据：", &address,tmpAction)

	fmt.Println("write err :", err, index)

	if err != nil {
		log.Println("send deviceId error", err)
		return
	}

}

func NewFocus(redis dbhelp.IRedis,caches *cache.Cache) *distributeFocus{
	this := &distributeFocus{
		iRedis:redis,
		caches:caches,
	}

	this.caches.OnEvicted(this.CallBack)

	return this
}