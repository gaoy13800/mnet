package msgProxy

import (
	"github.com/admpub/mahonia"
	"github.com/fatih/color"
	"log"
	"mnet/IBase"
	"mnet/cahce"
	"mnet/common"
	"mnet/dbhelp"
	"mnet/event"
	"mnet/push"
	"mnet/service"
	"mnet/util"
	"mnet/work"
	"strings"
	"sync"
	"time"
	"fmt"
)

const (
	Lock_Open_Cstu    = "3"
	Lock_Close_Cstu   = "4"
	Lock_Contact_Cstu = "2"
)

type MessageCenter struct {
	IDevices IBase.ITerminals

	iRedis dbhelp.IRedis

	syncTex sync.RWMutex

	taoMemory *cahce.Cache

	terminalHeart map[string]int
}

func (msg *MessageCenter) Notice(pushData *push.SessionEvent) {

	switch pushData.Type {
	case event.Connect:
		log.Printf("新会话建立  session id is %d", pushData.Sess.ID())
		return
	case event.Msg:
		msg.disposeMessage(pushData)
		return
	case event.Close:
		log.Printf(" 会话结束  session id is %d", pushData.Sess.ID())

		pushData.Sess.Close()
		return
	default:
		log.Println("未知的命令！ 不会处理")
		return
	}
}

func (this *MessageCenter) disposeMessage(pushData *push.SessionEvent) {

	cmsg := pushData.Data

	color.Green("收到信息: %s", cmsg)

	if strings.Contains(cmsg, "flag") || strings.Contains(cmsg, "open") {

		guid, action, data, _ := util.Resolve_no_splice(cmsg)

		switch action {
		case "flag":
			this.disposeFlag(guid, data, pushData.Sess)
			return
		case "open":
			this.disposeFlagOpen(pushData.Sess, guid)
			return
		}

		return
	} else if strings.Contains(cmsg, "hblv") {
		guid, _, data, _ := util.Resolve_no_splice(cmsg)

		this.heart(guid, data, pushData.Sess)

		return
	} else if strings.Contains(cmsg, "heartUnConnect"){

		this.disposeDoorCstu(pushData.Sess, "2") //摄像头失联

		return
	}

	if !strings.HasPrefix(cmsg, "wt") {
		//log.Println("未知消息信息, 会话将在5s后关闭")
		//
		//time.Sleep(time.Second * 5)
		//
		//this.closeAll(pushData.Sess, pushData.Sess.TerminalId())

		log.Println("未知消息，会话继续---------------------------------------------")

		return
	}

	if len(cmsg) != 21 && len(cmsg) != 6 {

		//log.Println("未知消息长度, 会话将在5s后关闭")
		//
		//time.Sleep(time.Second * 5)
		//
		//this.closeAll(pushData.Sess, pushData.Sess.TerminalId())

		log.Println("未知消息长度，会话继续---------------------------------------------")

		return
	}

	var action, deviceId string

	if len(cmsg) == 21 {
		action, deviceId = resolve(cmsg)
	} else if len(cmsg) == 6 {
		action = cmsg
		deviceId = pushData.Sess.TerminalId()
	}

	switch action {
	case "wtoveo":
		this.over_deal(deviceId, pushData.Sess, Lock_Open_Cstu)
		break
	case "wtgoid":
		this.build_connect(deviceId, pushData.Sess)
		break
	case "wtovec":
		this.over_deal(deviceId, pushData.Sess, Lock_Close_Cstu)
		break
	case "wtoveb":
		this.over_brut(deviceId, "1")

		time.Sleep(time.Second * 20)

		this.over_brut(deviceId, "0")
		break
	case "wthelo":
		log.Println("设备id:", deviceId)
		return
	case "wtCQMZ":

		//临时存储！！！！ 存储每个设备重启的数量

		global := cahce.SelectMemory(common.Temp_Global)

		var num int

		oldNum, err := global.Get("wt:reboot:num:" + deviceId)

		if !err{
			num = 1
		}else {
			num = oldNum.(int) + 1
		}

		global.Set("wt:reboot:num:" + deviceId, num, common.Long_Time_Expires)
	default:
		break
	}
}

func (this *MessageCenter) heart(guid string, status string, sess IBase.ISession) {

	this.syncTex.Lock()

	defer this.syncTex.Unlock()

	if !this.iRedis.IsExistClient(guid) {

		this.iRedis.InitClient(guid, sess.FromIBase().IP())

		this.iRedis.Save_(event.BELONG_IP, guid, sess.FromIBase().IP())
	}

	this.iRedis.Save_(event.SAVE_CSTU, guid, "5")//摄像头正常状态

	if !this.IDevices.IsExists(guid) {

		this.IDevices.Add(guid, sess)

		sess.SetTerminalId(guid)
	}

	if sess.(*work.SocketSession).ConnectStatus == "connect" {
		sess.StartPingDoor()

		sess.(*work.SocketSession).ConnectStatus = "interact"
	}

	sess.SetContainerTicker()

	if status == "0" {
		time.Sleep(time.Second * 2)
		sess.RawSend("OK") //send heart
	}




	/*var num int

	if v, ok := this.terminalHeart[guid]; ok {
		if v == 2 {
			sess.RawSend("OK") //send heart
			this.terminalHeart[guid] = 0
			return
		}

		num = v + 1
	} else {
		num = 1
	}

	this.terminalHeart[guid] = num*/
}

func (this *MessageCenter) disposeFlag(guid, flag string, sess IBase.ISession) {

	serializeFlag := mahonia.NewDecoder("gbk").ConvertString(flag)

	this.saveFlag(guid, serializeFlag)

	instance := service.NewService(guid, sess, this.iRedis)

	instance.Dispense(serializeFlag)
}

func (this *MessageCenter) disposeFlagOpen(sess IBase.ISession, deviceId string){

	sess.RawSend(deviceId + "over")
}

func (this *MessageCenter) saveFlag(guid, flag string) {
	this.iRedis.Save_(event.SAVE_FLAG, "", flag)
}

func (this *MessageCenter) disposeDoorCstu(sess IBase.ISession, status string){
	deviceId := sess.TerminalId()

	this.iRedis.Save_(event.SAVE_CSTU, deviceId, status)
}

// 地锁终端回应over所做处理
func (this *MessageCenter) over_deal(deviceId string, sess IBase.ISession, status string) {

	if ok := sess.CheckConnect(); !ok {
		log.Println("请检查是否建立真实连接")
		return
	}

	if len(deviceId) != 15 {
		log.Println("deviceId 长度无效！ ", deviceId)
		return
	}

	if status == Lock_Open_Cstu {
		this.taoMemory.Delete(deviceId + "_action_" + "open")
	} else {
		this.taoMemory.Delete(deviceId + "_action_" + "clse")
	}

	this.iRedis.Save_(event.SAVE_CSTU, deviceId, status)
}

func (this *MessageCenter) over_brut(deviceId string, status string) {

	this.taoMemory.Delete(deviceId + "_action_" + "brut")
	this.iRedis.Save_(event.SAVE_BRUT, deviceId, status)
}

//终端建立连接 存储设备id、新增sessionId
func (this *MessageCenter) build_connect(deviceId string, sess IBase.ISession) {

	this.syncTex.Lock()

	defer this.syncTex.Unlock()

	//干掉 之前的session 清除会话相关信息

	oldSession, err := this.IDevices.GetSessionByTerminalId(deviceId)

	if err == nil {
		this.closeAll(oldSession, deviceId)
	}

	sess.SetTerminalId(deviceId)

	if !this.IDevices.IsExists(deviceId) {
		this.IDevices.Add(deviceId, sess)
	}

	if !this.iRedis.IsExistClient(deviceId) {

		this.iRedis.InitClient(deviceId, sess.FromIBase().IP())

		this.iRedis.Save_(event.BELONG_IP, deviceId, sess.FromIBase().IP())
	}

	this.iRedis.Save_(event.SAVE_CSTU, deviceId, "4")

	//if !this.iRedis.IsNormalLockElectric(deviceId){
	//	sess.Encode("stus")
	//}

}

func (this *MessageCenter) FocusSub(data string) {
	log.Print("接收到订阅信息:")

	color.Green(data)

	if _, ok := this.taoMemory.Get(data + "_sub"); ok {
		return
	}

	this.taoMemory.Set(data+"_sub", data, time.Second*40)

	list := strings.Split(data, "|")

	if len(list) == 2 {
		deviceId := list[0]
		action := list[1]

		sess, err := this.IDevices.GetSessionByTerminalId(deviceId)

		if err == nil {
			//sendAction := rebuildAction(action)

			log.Print("向终端发送信息:")

			color.Green(deviceId + " wt" + action)

			this.taoMemory.Set(deviceId+"_action_"+action, 1, time.Second*10)

			sess.RawSend("wt" + action)
		} else {
			fmt.Println("不存在此会话！", err.Error())
		}
	}
}

/**
	缓存过期回调
*/
func (this *MessageCenter) cacheCallBack(key string, value interface{}) {

	if strings.Contains(key, "_sub") {
		return
	} else if strings.Contains(key, "_action_") {

		callValue := value.(int) + 1

		log.Println("动作回调发送！ 次数:", callValue)

		list := strings.Split(key, "_")

		if len(list) != 3 {
			return
		}

		deviceId, action := list[0], list[2]

		sess, err := this.IDevices.GetSessionByTerminalId(deviceId)

		if err != nil {
			log.Println("cache callback getSessionByDeviceId errors:", err)
			return
		}

		if callValue == 5 {
			this.iRedis.Save_(event.SAVE_CSTU, deviceId, Lock_Contact_Cstu)

			//todo  如果终端不回应是否干掉会话？  yes

			log.Println("发送次数超限 会话会断开")

			this.closeAll(sess, deviceId)

			this.taoMemory.Delete(key)

			return
		}

		sess.RawSend("wt" + action)

		this.taoMemory.Set(key, callValue, time.Second*10)
	}
}

/**
	关闭会话 remove 设备id与sess的对应关系
 */
func (this *MessageCenter) closeAll(sess IBase.ISession, deviceId string){

	if sess.TerminalId() != "" {
		this.taoMemory.Delete(sess.TerminalId() + "_action_open")
		this.taoMemory.Delete(sess.TerminalId() + "_action_clse")
		this.taoMemory.Delete(sess.TerminalId() + "_action_brut")

		this.taoMemory.Delete(sess.TerminalId() + "|open_sub")
		this.taoMemory.Delete(sess.TerminalId() + "|brut_sub")
		this.taoMemory.Delete(sess.TerminalId() + "|clse_sub")
	}

	sess.Close()

	this.IDevices.Remove(deviceId)
}


func resolve(data string) (string, string) {

	action := data[:6]

	deviceId := data[6:]

	return action, deviceId
}

func NewMessageCenter(redis dbhelp.IRedis) *MessageCenter {

	msgInstance := &MessageCenter{
		IDevices:      work.NewTerminalManage(),
		iRedis:        redis,
		taoMemory:     cahce.SelectMemory(common.Message_Global),
		terminalHeart: make(map[string]int),
	}

	msgInstance.taoMemory.OnEvicted(msgInstance.cacheCallBack)

	return msgInstance
}
