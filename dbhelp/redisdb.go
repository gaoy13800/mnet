package dbhelp

import (
	"github.com/go-redis/redis"
	"sync"
	"fmt"
	"log"
	"mnet/event"
	"strconv"
	"time"
	"net"
	"mnet/task"
)

type IRedis interface {

	Save_(t event.ACTION, guid string, data interface{})

	IsNormalStus(guid string) bool

	IsExistClient(guid string) bool

	InitClient(guid string, ipAddr int64)error

	Save_Udp(guid ,status string,addr *net.UDPAddr)

	IsExistGuId(guid string) string

	DelGuId(guid string)

	SetClientUdp(guid,status string)
}

const (
	IPSKEY       = "wt:sevice"
	SubKey       = "wtClintChan"
	ClientKey    = "client:"
	ClientEx     = "client:expire:"
	ClientStusEx = "client:expire:stus:"
	Service      = "service:"
	FlagKey      = "wtFlag"

	FlagClientKey = "flag:detail:"


	UdpListKey = "lock:client:list"
	UdpDetailKey = "lock:client:detail:"
)

type redisDeal struct {

	redis *redis.Client

	wait  sync.WaitGroup

	closeSub bool

	//closeExSub bool
}

func (this *redisDeal) pass() error {

	pong, err := this.redis.Ping().Result()

	if err != nil || pong != "PONG" {
		fmt.Println("redis ping error:", err)

		return err
	}

	return nil
}

func (this *redisDeal) Save_(t event.ACTION, guid string, data interface{}) {

	switch t {
	case event.INIT:
		this.init(data.(int64))
		break
	case event.SAVE_CSTU:
		//save cstu 保存状态
		this.saveCstu(guid, data.(string))
		break
	case event.SAVE_STUS:
		//保存电量 save stus
		this.saveStus(guid, data.(string))
		break
	case event.SAVE_FLAG:
		//保存车牌号
		this.saveFlag(data.(string))
		break
	case event.BELONG_IP:
		this.belongToIp(guid, data.(int64))
		break
	case event.SAVE_BRUT:
		this.saveLockBrut(guid, data.(string))
		break
	default:
		log.Println("unknown action")
	}

}

func (this *redisDeal) InitClient(guid string, ipAddr int64) error {

	if index, err := this.redis.Exists(ClientKey + guid).Result(); err != nil {
		if index > 0 {
			return nil
		}
	}

	if _, err := this.redis.HSet(ClientKey+guid, "FmIP", strconv.FormatInt(ipAddr, 10)).Result(); err != nil {
		return err
	}

	return nil
}

func (this *redisDeal) IsNormalStus(guid string) bool {
	n, _ := this.redis.TTL(ClientStusEx + guid).Result()

	if !(n > 0) {
		return false
	}

	return true
}

func (this *redisDeal) IsExistClient(guid string) bool {
	ret, _ := this.redis.Exists(ClientKey + guid).Result()

	if ret > 0 {
		return true
	}

	return false
}


func (this *redisDeal) init(ipAddr int64) error {

	if _, err := this.redis.SAdd(IPSKEY, strconv.FormatInt(ipAddr, 10)).Result(); err != nil {
		return err
	}

	return nil
}

func (this *redisDeal) belongToIp(guid string, ipAddr int64) error {

	ip := strconv.FormatInt(ipAddr, 10)

	if _, err := this.redis.SAdd(Service+ip, guid).Result(); err != nil {
		return err
	}

	return nil
}

func (this *redisDeal) saveCstu(guid, cstu string)bool{
	_, err := this.redis.HSet(ClientKey + guid, "cstu", cstu).Result()

	if err != nil{
		return false
	}

	return true
}

func (this *redisDeal) saveStus(guid, stus string)bool{

	_, err := this.redis.HSet(ClientKey + guid, "stus", stus).Result()

	this.redis.Set(ClientStusEx + guid, time.Now().Format("150405"), time.Second * 60 * 60 * 24)

	if err != nil{
		return false
	}

	return true
}

func (this *redisDeal) saveFlag(flag string)bool{
	_, err := this.redis.SAdd(FlagKey, flag).Result()

	if err != nil{
		return false
	}

	return true
}

func (this *redisDeal) saveLockBrut(deviceId, stus string) bool{
	if _, err := this.redis.HSet(ClientKey + deviceId, "bstu", stus).Result();  err != nil{
		return false
	}

	return true
}

func (this *redisDeal) subThead(queue task.EventQueue) {

	subHandler := this.redis.Subscribe(SubKey)

	for  {
		if !this.closeSub {
			break
		}

		subData, err := subHandler.ReceiveMessage()

		if err != nil {
			subHandler.Close()
			log.Println("subThead receive errors:", err)
			break
		}

		queue.PostData(subData.Payload)
	}

	subHandler.Close()

	this.wait.Done()
}

/*func (this *redisDeal) subExpiredThead(db int, queue chanqueue.EventQueue){
	key := "__keyevent@" + strconv.Itoa(db) + "__:expired"

	subHandler := this.redis.PSubscribe(key)

	for  {
		if !this.closeExSub {
			break
		}

		subData, err := subHandler.ReceiveMessage()

		if err != nil {
			subHandler.Close()
			log.Println("subExpiredThead receive errors:", err)
			break
		}

		queue.PostData(subData.Payload)
	}

	subHandler.Close()

	this.wait.Done()
}*/

func (this *redisDeal) Flag_(guid string, data interface{}) {
	_, err := this.redis.Set("test:flag:out:guids:"+guid, data.(string), -1).Result()

	if err != nil {
		fmt.Println(err)
	}
}

func (this *redisDeal) saveCamera(deviceId string, isOpen bool){
	this.redis.HSet(ClientKey + deviceId, "status", isOpen)
}

func (this *redisDeal) Save_Udp(guid ,status string, addr *net.UDPAddr)  {
	ip := addr.IP.String()
	port := strconv.Itoa(addr.Port)
	address := ip + ":" + port
	//保存终端信息
	this.redis.SAdd(UdpListKey,guid)
    this.redis.HSet(UdpDetailKey + guid,"addr",address).Result()
	//设置地锁状态
	this.SetClientUdp(guid,status)
}

//初始化地锁状态为4
func (this *redisDeal) initClientUdp(guid string){

	ret, _ := this.redis.Exists(ClientKey + guid).Result()

	if ret > 0 {
		return
	}
	this.redis.HSet(ClientKey + guid, "cstu", 4)
}

//设置地锁状态（心跳传过来）
func (this *redisDeal) SetClientUdp(guid,status string)  {
	this.redis.HSet(ClientKey + guid, "cstu", status)
}

//判断guid是否属于udp
func (this *redisDeal)IsExistGuId(guid string) string {

	detail,err := this.redis.HGet(UdpDetailKey+guid, "addr").Result()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return detail
	//return this.redis.SIsMember(UdpListKey,guid).Val()
}

//删除属于udp的guid
func (this *redisDeal) DelGuId(guid string) {
	this.redis.SRem(UdpListKey,guid)
	this.redis.Del(UdpDetailKey,guid)
}

func NewRedisDeal(dbAddr, password string, db int, subQueue task.EventQueue, exSub task.EventQueue) IRedis {

	this := &redisDeal{
		redis: redis.NewClient(&redis.Options{
			Addr:     dbAddr,
			Password: password,
			DB:       db,
		}),
		closeSub: true,
	}

	err := this.pass()

	if err != nil {
		log.Println("NewRedisDeal ping error:", err)
		return nil
	}

	this.wait.Add(2)

	go func() {
		this.wait.Wait()

		this.closeSub = false
	}()

	go this.subThead(subQueue)

	//go this.subExpiredThead(db, exSub)

	return this
}




