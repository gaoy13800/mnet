package msgProxy

//import (
//	"fmt"
//	"github.com/admpub/mahonia"
//	"log"
//	"mnet/IBase"
//	"mnet/dbhelp"
//	"mnet/event"
//	"mnet/push"
//	"mnet/service"
//	"mnet/work"
//	"strconv"
//	"sync"
//	"strings"
//)
//
//type msgDeal struct {
//	iTerminal IBase.ITerminals
//
//	iRedis dbhelp.IRedis
//
//	iBase IBase.IBaseNet
//
//	terminalHeart map[string]int
//
//	syncTex sync.RWMutex
//}
//
//func (this *msgDeal) Focus(proxy *push.SessionEvent) {
//
//	mtype := proxy.Type
//
//	switch mtype {
//	case event.Connect:
//		log.Println("connect build, session id is", proxy.Sess.ID())
//		break
//	case event.Close:
//		log.Println("close connect, session id is", proxy.Sess.ID())
//		this.saveCstu(proxy.Sess.TerminalId(), event.BREAK_DOWN)
//		break
//	case event.Msg:
//		this.call(proxy)
//		break
//	default:
//		log.Println("unknown msg type")
//	}
//}
//
//func (this *msgDeal) call(proxy *push.SessionEvent) {
//
//	fmt.Println("msg:", proxy.Data)
//
//	//fmt.Println(mahonia.NewDecoder("gbk").ConvertString(proxy.Data))
//
//	guid, action, data, isNormal := resolve_no_splice(proxy.Data)
//
//	if !isNormal {
//		log.Println("无法解析此条消息")
//		return
//	}
//
//	switch action {
//	case "open":
//		//修改地锁状态 发送over
//		this.open(guid, proxy.Sess)
//		break
//	case "clse":
//		//修改地锁状态 发送over
//		this.clse(guid, proxy.Sess)
//		break
//	case "brut":
//		//无状态更改
//		break
//	case "nlve":
//		//告知车辆未驶离 更改状态
//		this.nlve(guid, proxy.Sess)
//		break
//	case "cstu":
//		//告知地锁当前状态 发送over
//		this.saveCstu(guid, data)
//		break
//	case "stus":
//		//修改电量信息
//		this.saveStus(guid, data)
//		break
//	case "hblv":
//		//终端心跳 三次响应
//		//保存 终端 session 关系
//		this.heart(guid, data, proxy.Sess)
//		break
//	case "flag":
//		//告知车牌号 业务处理 存储
//		this.flagDeal(guid, data, proxy.Sess)
//		break
//	default:
//		log.Println("未能识别动作", action, isNormal)
//		break
//
//	}
//}
//
//func (this *msgDeal) open(guid string, sess IBase.ISession) {
//
//	if sess.IsLock() == true{
//		this.saveCstu(guid, event.LOCK_OPEN)
//	}
//
//	this.sendOver(guid, sess)
//}
//
//func (this *msgDeal) clse(guid string, sess IBase.ISession) {
//	this.saveCstu(guid, event.LOCK_CLOSE)
//	this.sendOver(guid, sess)
//}
//
//func (this *msgDeal) nlve(guid string, sess IBase.ISession) {
//	this.saveCstu(guid, event.NOT_LEAVE)
//	this.sendOver(guid, sess)
//}
//
//func (this *msgDeal) heart(guid string, status string, sess IBase.ISession) {
//
//
//	this.syncTex.Lock()
//
//	defer this.syncTex.Unlock()
//
//	bestatus, _ := strconv.Atoi(status)
//
//	if !this.iTerminal.IsExists(guid) && bestatus != 0 {
//
//		this.iTerminal.Add(guid, sess)
//
//		sess.SetTerminalId(guid)
//	}
//
//	if !this.iRedis.IsExistClient(guid) && bestatus != 0 {
//		this.iRedis.InitClient(guid, sess.FromIBase().IP())
//	}
//
//	if !this.iRedis.IsNormalStus(guid) && bestatus != 0 {
//		sess.RawSend("stus")
//	}
//
//	if bestatus != 0 {
//		sess.SetClientType(true)
//		this.saveCstu(guid, status)
//		this.iRedis.Save_(event.BELONG_IP, guid, sess.FromIBase().IP())
//	}
//
//
//
//	var num int
//
//	if v, ok := this.terminalHeart[guid]; ok {
//		if v == 2 {
//			sess.RawSend("OK") //send heart
//			this.terminalHeart[guid] = 0
//			return
//		}
//
//		num = v + 1
//	} else {
//		num = 1
//	}
//
//	this.terminalHeart[guid] = num
//}
//
//func (this *msgDeal) sendOver(guid string, session IBase.ISession) {
//	session.RawSend(guid + "over")
//}
//
//func (this *msgDeal) sendOpen(guid string, session IBase.ISession) {
//
//	session.RawSend(guid + "open")
//}
//
//func (this *msgDeal) saveStus(guid, stus string) {
//	this.iRedis.Save_(event.SAVE_STUS, guid, stus)
//}
//
//func (this *msgDeal) saveCstu(guid, cstu string) {
//	this.iRedis.Save_(event.SAVE_CSTU, guid, cstu)
//}
//
////func (this *msgDeal) saveFlag(guid, flag string) {
////
////	//serializeFlag := mahonia.NewDecoder("gbk").ConvertString(flag)
////
////	this.iRedis.Save_(event.SAVE_FLAG, "", flag)
////}
//
////func (this *msgDeal) saveBrut(guid, status string){
////	this.iRedis.Save_(event.SAVE_BRUT, guid, status)
////}
//
////func (this *msgDeal) flagDeal(guid, flag string, sess IBase.ISession) {
////
////	serializeFlag := mahonia.NewDecoder("gbk").ConvertString(flag)
////
////	this.saveFlag(guid, serializeFlag)
////
////	instance := service.NewService(guid, sess, this.iRedis)
////
////	instance.Dispense(serializeFlag)
////}
//
//
//func (this *msgDeal) FocusSubEx(data string) {
//
//	//todo
//}
//
//
//func (this *msgDeal) FocusSub(data string) {
//
//
//	fmt.Println("recv", data)
//
//
//	list := strings.Split(data, "|")
//
//	if len(list) == 2 {
//		guid := list[0]
//		cmd := list[1]
//		session, err := this.iTerminal.GetSessionByTerminalId(guid)
//
//		//session.RawSend(guid + "test")
//
//		if err == nil {
//
//			log.Println("send :", guid + cmd)
//
//			session.RawSend(guid + cmd)
//		} else {
//			fmt.Println("FocusSub getSessionByGuid error", err.Error())
//		}
//	}
//}
//
//func resolve(msg string) (string, string, bool) {
//
//	length := len(msg)
//
//	switch {
//	case length == 4:
//		return msg, " ", true
//	case length == 5:
//		return msg[:4], msg[4:], true
//	case length == 6:
//		return msg[:4], msg[4:], true
//	case length > 6:
//		return msg[:4], msg[4:], true
//	default:
//		return msg, "", false
//	}
//}
//
//func resolve_no_splice(data string) (string, string, string, bool) {
//
//	length := len(data)
//
//	if length < 32 {
//		return "", "", "", false
//	}
//
//	guid := data[:32]
//
//	fmt.Println(guid)
//	action, any, isNormal := resolve(data[32:])
//
//	return guid, action, any, isNormal
//}
//
//func NewMsgDeal(redis dbhelp.IRedis, base IBase.IBaseNet) *msgDeal {
//
//	return &msgDeal{
//		iTerminal:     work.NewTerminalManage(),
//		iRedis:        redis,
//		iBase:         base,
//		terminalHeart: make(map[string]int),
//	}
//}
