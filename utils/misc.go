package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"

	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
	_ "github.com/garyburd/redigo/redis"
	_ "github.com/gomodule/redigo/redis"
)

//连接函数
func RedisOpen(server_name, redis_addr, redis_port, redis_dbnum string) (bm cache.Cache, err error) {
	redis_config_map := map[string]string{
		"key":   server_name,
		"conn":  redis_addr + ":" + redis_port,
		"dbnum": redis_dbnum,
	}
	redis_config, _ := json.Marshal(redis_config_map)
	bm, err = cache.NewCache("redis", string(redis_config))
	if err != nil {
		log.Println("连接redis错误", err)
		return nil, err
	}
	return bm, nil
}

//md5加密
func Getmd5string(s string) string {
	m := md5.New()
	return hex.EncodeToString(m.Sum([]byte(s)))
}
