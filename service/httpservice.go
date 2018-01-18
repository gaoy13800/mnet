package service

import (
	"errors"
	"log"
	"mnet/IBase"
	"mnet/dbhelp"
	"mnet/event"
)

var (
	Error_Notice_CloseParking = errors.New("notice close parking error")
	Error_Notice_OpenParking  = errors.New("notice open parking error")
	Error_Notice_OpenCourt    = errors.New("notice open court error")
	Error_Notice_CloseCourt   = errors.New("notice close court error")
)

type msgService struct {
	terminalT int

	terminalId string

	guids []string

	param map[string]string

	sess IBase.ISession

	redis dbhelp.IRedis
}

//总控Control

func (this *msgService) Dispense(carNum string) {
	t, err := this.getDeviceType()

	if err != nil {
		return
	}

	switch t {
	case event.Parking_Open:
		if isVip, err := this.isFact_parking(this.terminalId, carNum); !isVip || err != nil {
			return
		}
		break
	case event.Parking_Close:
		this.out_parking(carNum)
		break
	case event.Court_Open:
		if isPermission, err := this.check_court(carNum); err != nil || !isPermission {
			return
		}
		log.Println("send:", this.terminalId+"open")
		this.sess.RawSend(this.terminalId + "open")

		err1 := this.open_court(carNum)

		if err1 != nil {
			log.Println("notice open court call interface error:", err1)
			return
		}
		break
	case event.Court_Close:
		if ok, err := this.close_court(carNum); err != nil || !ok {
			break
		}
		log.Println("send:", this.terminalId+"open")
		this.sess.RawSend(this.terminalId + "open")
	default:
		log.Println("错误的类型 -- ")
		break
	}
}

//获取 guid 类型
func (this *msgService) getDeviceType() (event.GTYPE, error) {

	t, ok := Call(WPCT, this.param)

	if !ok {
		log.Println("Pct call error")
		return 0, nil
	}

	var curType event.GTYPE

	switch t.(int) {
	case 1:
		curType = event.Parking_Open
		break
	case 2:
		curType = event.Parking_Close
		break
	case 3:
		curType = event.Court_Open
		break
	case 4:
		curType = event.Court_Close
		break
	default:
		curType = event.Other
		break
	}

	//this.sess.SetTerminalType(curType)

	return curType, nil
}

//是否有权限
func (this *msgService) isFact_parking(terminalId, carNum string) (bool, error) {

	this.param["carNum"] = carNum

	data, ok := Call(WPCC, this.param)

	if !ok {
		return false, Error_Notice_OpenParking
	}

	isMember := data.(map[string]interface{})["result"]

	log.Println("call pcc result is:", isMember)

	if isMember.(string) == "false" {
		return false, nil
	}

	return true, nil
}

//停车场出口检测到车牌操作
func (this *msgService) out_parking(carNum string) {

	isOut, err := this.close_parking(carNum)

	if err != nil {
		return
	}

	if isOut {
		log.Println("device send:", this.terminalId+"open")
		this.sess.RawSend(this.terminalId + "open")
	}
}

//调用结束订单接口
func (this *msgService) close_parking(carNum string) (bool, error) {

	this.param["carNum"] = carNum

	data, ok := Call(WPOEND, this.param)

	log.Println("poend call result:", data)

	if !ok {
		return false, Error_Notice_CloseParking
	}

	check := data.(string)

	if check == "false" {
		return false, nil
	} else {
		return true, nil
	}
}

//小区 车牌扫描后检测是否有权限进入
func (this *msgService) check_court(carNum string) (bool, error) {
	this.param["carNum"] = carNum

	result, ok := Call(WVCC, this.param)

	if !ok {
		return false, Error_Notice_OpenCourt
	}

	check := result.(string)

	if check == "false" {
		return false, nil
	} else {
		return true, nil
	}
}

//订单正式开始 小区
func (this *msgService) open_court(carNum string) error {
	this.param["carNum"] = carNum

	_, ok := Call(WVUPO, this.param)

	if !ok {
		return Error_Notice_OpenCourt
	}

	return nil
}

//结束订单 小区类型
func (this *msgService) close_court(carNum string) (bool, error) {

	param := this.param

	param["carNum"] = carNum

	result, ok := Call(WVPOEND, this.param)

	if !ok {
		return false, Error_Notice_CloseCourt
	}

	check := result.(string)

	if check == "false" {
		return false, nil
	} else {
		return true, nil
	}
}

func NewService(deviceId string, sess IBase.ISession, redis dbhelp.IRedis) *msgService {

	this := &msgService{
		terminalId: deviceId,
		sess:       sess,
		redis:      redis,
	}

	this.param = map[string]string{"guid": deviceId}

	return this
}
