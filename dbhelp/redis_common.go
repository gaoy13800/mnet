package dbhelp

import (
	"github.com/go-redis/redis"
	"sync"
	"mnet/conf"
	"log"
)

type RedisCommon struct {
	redis    *redis.Client

	syncWait sync.WaitGroup
}

func (this *RedisCommon) PublishMessage(msg string){

	this.redis.Publish("wtClintChan", msg)
}

func (this *RedisCommon) SetClientMessage(key ,hashKey, hashValue string){

	this.redis.HSet(key, hashKey, hashValue)
}

func (this *RedisCommon) GetClientData(key string)map[string]string{

	datas, err := this.redis.HGetAll(key).Result()

	if err != nil {
		log.Println("RedisCommon GetClientData error", datas)
		return nil
	}

	return datas
}


func NewRedisCommon() *RedisCommon{

	this := &RedisCommon{
		redis: redis.NewClient(&redis.Options{
			Addr:     conf.Conf["redis_addr"],
			Password: conf.Conf["redis_passwd"],
			DB:       1,
		}),
	}

	return this
}
