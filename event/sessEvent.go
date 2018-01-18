package event

const (
	Connect = iota + 1
	Close
	Msg
)


type ACTION uint

const (
	SAVE_STUS ACTION = iota + 1 //电量
	SAVE_CSTU  //终端状态
	SAVE_FLAG
	INIT
	BELONG_IP
	SAVE_BRUT
)



const (
	TICKER  = 60
)

//车位锁状态

const (
	BREAK_DOWN = "1"
	OUT_CONTACT = "2"
	LOCK_OPEN = "3"
	LOCK_CLOSE = "4"
	NOT_LEAVE = "5"
)


//Guid 类型

type GTYPE int

const (
	Parking_Open GTYPE = iota + 1
	Parking_Close
	Court_Open
	Court_Close
	Other
)


type Proto int

const (
	Proto_UDP Proto = iota + 1
	Proto_TCP
)

const (
	CACHE_DEADLINE  = 20
	CACHE_INITIAL_TIME  = 3600//time.Second * 31536000
)

