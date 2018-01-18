package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"mnet/conf"
)

func GetMyIp() net.IP {

	platform :=  os.Args[1]

	if platform == "master" {
		return net.ParseIP(conf.Conf["temp_en0_ip"])
	}

	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, addr := range addrs {

		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP

			}

		}
	}
	return nil
}

// TranferIpToint64 把ip地址转换为长整型
func TranferIpToint64(ipnr net.IP) (int64, error) {
	if ipnr == nil {
		return -1, errors.New("没有可用的ip地址")
	}
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64
	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum, nil
}

// TranferInt64ToIp 把ip地址转换为 类型
func TranferInt64ToIp(ipnr int64) net.IP {
	var bytes [4]byte

	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)
	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

func TranferIpToStringFull(ipnr net.IP) (string, error) {

	var result string
	if ipnr == nil {
		return result, errors.New("没有可以使用的ip地址")
	}
	list := strings.Split(ipnr.String(), ".")
	if len(list) == 4 {
		for i, v := range list {
			switch len(v) {
			case 1:
				list[i] = "00" + list[i]
				break
			case 2:
				list[i] = "0" + list[i]
				break
			}

		}
		result = strings.Join(list, ".")

		return result, nil

	}
	return result, errors.New("没有可以使用的ip地址")
}


func resolve(msg string) (string, string, bool) {

	length := len(msg)

	switch {
	case length == 4:
		return msg, " ", true
	case length == 5:
		return msg[:4], msg[4:], true
	case length == 6:
		return msg[:4], msg[4:], true
	case length > 6:
		return msg[:4], msg[4:], true
	default:
		return msg, "", false
	}
}

func Resolve_no_splice(data string) (string, string, string, bool) {

	length := len(data)

	if length < 32 {
		return "", "", "", false
	}

	guid := data[:32]

	action, any, isNormal := resolve(data[32:])

	return guid, action, any, isNormal
}
