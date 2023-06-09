package Common

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	//if err != nil {
	//	return "", err
	//}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}

	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}

	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}

func GetServerIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return ""
	}
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.IsGlobalUnicast() && !ipnet.IP.IsInterfaceLocalMulticast() {
			if ipnet.IP.To4() != nil {
				//log.Println(ipnet.IP.String())
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func Page404(c *gin.Context) {
	 
	c.JSON(http.StatusNotFound, gin.H{
		"code": http.StatusNotFound,
		"msg":  "404, page not exists!",
		"data": "",
	})
}

var FailSlaveTimes = sync.Map{}
 
func GetActiveSlaveList() (slaveList []ServerInfo, err error) {
	list, err := RedisClient.HGetAll(Config.GetString("key.RedisSlaveServerKey")).Result()

	log.Println("GetActiveSlaveList:start")
	if nil == err {
		for _, v := range list {
			log.Println("GetActiveSlaveList:loop:", v)

			var slaveInfo ServerInfo
			err := Json.UnmarshalFromString(v, &slaveInfo)

			if nil == err {
				if true == slaveInfo.Status &&
					!(slaveInfo.HostIP == Config.GetString("server.hostIP") &&
						slaveInfo.HostPort == Config.GetString("server.hostPort")&&
						slaveInfo.HostIP == Config.GetString("server.ip") &&
						slaveInfo.HostPort == Config.GetString("server.port")) {

					ok, err := slaveInfo.CheckServerHealth()
					if true == ok {
						slaveList = append(slaveList, slaveInfo)
					} else {
						log.Println("GetActiveSlaveList:err:", err)
						failTime, ok := FailSlaveTimes.LoadOrStore(slaveInfo.GetName(), 1)
						if true == ok {
							if failTime.(int) > 10 {
								RedisClient.HDel(Config.GetString("key.RedisSlaveServerKey"), slaveInfo.GetName())
							}
							FailSlaveTimes.Store(slaveInfo.GetName(), failTime.(int)+1)
						}
					}
				} else {
				 
					log.Println("GetActiveSlaveList:CheckStatusFail:",
						"true == slaveInfo.Status",
						true == slaveInfo.Status,
						"slaveInfo.HostIP == Config.GetString(server.hostIP)",
						slaveInfo.HostIP == Config.GetString("server.hostIP"),
						slaveInfo.HostIP, Config.GetString("server.hostIP"),
						"slaveInfo.HostPort == Config.GetString(server.hostPort)",
						slaveInfo.HostPort == Config.GetString("server.hostPort"),
						slaveInfo.HostPort, Config.GetString("server.hostPort"),
					)
				}
			} else {
		 
				log.Println("GetActiveSlaveList:err2:", err)

			}

		}
	} else {
		log.Println("GetActiveSlaveList:err3:", err)
		return slaveList, err
	}

	return slaveList, nil
}

func FastAtoi(num string) int {
	ret, _ := strconv.Atoi(num)
	return ret
}

func FastJsonMarshal(_json interface{}) string {
	str, _ := Json.MarshalToString(_json)
	return str
}

func Md5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}
