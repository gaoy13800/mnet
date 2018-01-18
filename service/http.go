package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mnet/conf"
	"net/http"
	"net/url"
)

type CallType int

const (
	WPCT CallType = iota + 1
	WPCC
	WPOEND
	WVCC
	WVUPO
	WVPOEND
)

type PCC struct {
	Result string `json:"result"`
}

type PCT struct {
	Result int `json:"result"`
}

type SingleResult struct {
	Result string `json:"result"`
}

//_________________


/**
	接口说明

	PCT 获取摄像头所在类型 > 1、 停车场、小区 2、 进口 出口  每次车牌号的识别及接口的调用都会去先识别其门禁类型
	PCC 获取是否为本系统会员并且开启订单
	POEND 获取是否有权限开启门禁 > true 开启门禁 false 不进行操作

	VCC  获取门禁入口是否放行 true 放行 false 不进行操作
	VUPO 通知业务端开启订单接口
	VPOEND  通知业务端结束订单，并询问是否放行 true 放行 false 不进行任何操作
 */


func httpGet(serviceUrl string, param map[string]string) ([]byte, error) {

	if len(param) == 0 {
		return nil, errors.New("param length invalid")
	}

	u, _ := url.Parse(serviceUrl)
	q := u.Query()

	for k, v := range param {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()

	res, err := http.Get(u.String())

	fmt.Println("call url:", u.String())

	if err != nil {
		log.Print(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	return body, nil
}

func Call(t CallType, param map[string]string) (interface{}, bool) {

	taskUrl := conf.Conf["Service_url"]

	var isPct bool = false
	var isPcc bool = false
	var callUrl string

	switch t {
	case WPCT:
		callUrl = taskUrl + "PCT"
		isPct = true
		break
	case WPCC:
		callUrl = taskUrl + "PCC"
		isPcc = true
		break
	case WPOEND:
		callUrl = taskUrl + "POEND"
		break
	case WVCC:
		callUrl = taskUrl + "VCC"
		break
	case WVUPO:
		callUrl = taskUrl + "VUPO"
		break
	case WVPOEND:
		callUrl = taskUrl + "VPOEND"
		break
	default:
		log.Println("接口调用类别无效")
		break
	}

	bytesData, err := httpGet(callUrl, param)

	if err != nil {
		return nil, false
	}

	if isPcc {
		var data PCC

		json.Unmarshal(bytesData, &data)

		response := map[string]interface{}{"result": data.Result}

		log.Println("调用接口结果:", data.Result)

		return response, true
	}

	if isPct {
		var data PCT

		json.Unmarshal(bytesData, &data)

		log.Println("调用接口结果:", data.Result)

		return data.Result, true
	}

	var data SingleResult

	json.Unmarshal(bytesData, &data)

	log.Println("调用接口结果:", data.Result)

	return data.Result, true
}
