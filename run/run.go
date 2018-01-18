package run

import (
	"log"
	"os"
	"fmt"
	"mnet/conf"
	"mnet/socket"
	"mnet/event"
	"mnet/dbhelp"
	"mnet/msgProxy"
	"mnet/task"
	"mnet/util"
	cache "mnet/cahce"
	"time"
	work2 "mnet/work"
	"strconv"
	"mnet/push"
	"strings"
	"mnet/IBase"
	"mnet/common"
)


func Run(){
	// 读取配置文件
	chanNum, _ := strconv.Atoi(conf.Conf["chanNum"])

	listenAddress := fmt.Sprintf("0.0.0.0:%s", conf.Conf["socket_tcp_port"])

	udpListenAddress := fmt.Sprintf("0.0.0.0:%s", conf.Conf["socket_udp_port"])

	redisAddr := conf.Conf["redis_addr"]

	redisDb, _ := strconv.Atoi(conf.Conf["redis_db"])

	fmt.Printf("chanNum start: %d\nlisten tcp address: %s\nlisten udp address: %s select DB %d \n", chanNum, listenAddress,udpListenAddress, redisDb)

	// 获取本地ip信息

	localIp := util.GetMyIp()

	fullIp, err := util.TranferIpToStringFull(localIp)

	if err != nil {
		log.Println("tranfer Full ip error:", err)
		os.Exit(0)
	}

	fmt.Println("ip:", fullIp)

	ip64, err := util.TranferIpToint64(localIp)

	if err != nil {
		log.Println("transfer ip to int64 error", err)
		os.Exit(0)
	}

	// work 正式运行 all program
	work(ip64, listenAddress,udpListenAddress, fullIp, redisAddr, conf.Conf["redis_passwd"], chanNum, redisDb)
}


/**
	port
		tcp 监听	 6001
		udp 监听 	 6002
		web server	 10014

	subscribe redis channel : wtClintChan
 */


func work(ip int64, listenAddress,udpListenAddress, ipStr, redisAddr, password string, chanNum, db int){


	//启动TCP服务器监听, 维护chan队列进行消息通讯
	peerQueue := task.NewEventQueue(chanNum)

	peer := socket.NewPeer(peerQueue, ip)

	peer.Start(listenAddress)


	//启动UDP服务器监听，维护chan队列进行消息通讯
	udpQueue := task.NewEventQueue(chanNum)

	vipQueue := task.NewEventQueue(chanNum)

	storageUdp := work2.NewUdpManage() // udpManage 已经不用了

	udpPeer := socket.NewUdpPeer(udpQueue,ip, vipQueue, storageUdp)

	udpPeer.UdpStart(udpListenAddress)



	//---------------新增两个chan队列 分配给redis广播订阅和过期事件订阅

	subQue := task.NewEventQueue(chanNum)

	//exQue := task.NewEventQueue(chanNum)  //redis key 过期已经不用了

	redis := dbhelp.NewRedisDeal(redisAddr, password, db, subQue, nil)


	//初始化 wt:sevice
	redis.Save_(event.INIT, "", ip)

	//proxy := msgProxy.NewMsgDeal(redis, peer)

	msgCenter := msgProxy.NewMessageCenter(redis)

	go timerDispose(msgCenter.IDevices)

	caches := cache.New(10 * time.Minute, 10 * time.Second)

	focus := msgProxy.NewFocus(redis, caches)
	

	//开启web server

	go socket.NewWebServer().RunWork()



	//tcp 消息接收
	go func() {

		for single := range peerQueue.Queue {

			data := single.(*push.SessionEvent)

			msgCenter.Notice(data)
		}
	}()


	//临时 tcp  接收业务端发送的信息 redis subscribe
	go func() {
		for {
			data := <- subQue.Queue

			msgCenter.FocusSub(data.(string))
		}
	}()


	//udp 订阅业务端
	//go func() {
	//	for {
	//		data := <-subQue.Queue
	//
	//		focus.FocusSub(data.(string), udpPeer.UdpManage)
	//	}
	//}()

	// udp消息处理

	go func() {

		for single := range udpQueue.Queue{

			focus.MsgFocus(single, event.Proto_UDP)

		}

	}()

	go func(){

		for  {
			vipData := <-vipQueue.Queue

			focus.VipFocus(vipData.(string), udpPeer.UdpManage)

		}
	}()

}


func timerDispose(devices IBase.ITerminals){

	time.Sleep(time.Second * 5)


	ticker := time.NewTicker(time.Second * 3)

	globalDB := cache.SelectMemory(common.Message_Global)


	for  {
		select {
		case <- ticker.C:

			sessManage :=devices.GetDeviceIds()

			if len(sessManage) > 0 {
				deviceIds := strings.Join(sessManage, "|")
				globalDB.Set("wt:tao:deviceList", deviceIds, common.Long_Time_Expires)
			}else {
				globalDB.Delete("wt:tao:deviceList")
			}
		}
	}


}
