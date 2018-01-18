package common

import "time"

type PushType uint

const (

	Connect PushType = iota + 1

	Notice

	Close
)

type DBLockAction uint

const (
	Lock_Cstu DBLockAction = iota + 1

	Lock_Brut

	Lock_Stus

	Lock_Init

	Lock_Belong_IP
)

type CacheType uint

const (
	Message_Global CacheType = iota + 1

	Temp_Global


)

const Long_Time_Expires = time.Hour * 24 * 265 * 3
