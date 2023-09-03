package snowflake

import (
	"errors"
	"time"

	sf "github.com/bwmarrin/snowflake"
)

const (
	_dafaultStartTime = "2020-12-31" // 默认开始时间
)

var node *sf.Node

// Init 雪花算法组件初始化,正常应该把雪花算法当成一个独立的服务部署
// startTime 开始时间
// machineID 机器id 
func Init(startTime string, machineID int64) (err error) {
	if machineID < 0 {
		return errors.New("snowflake need machineID")
	}
	if len(startTime) == 0 {
		startTime = _dafaultStartTime
	}
	var st time.Time
	st, err = time.Parse("2006-01-02", startTime)
	if err != nil {
		return
	}
	sf.Epoch = st.UnixNano() / 1000000 // 时间戳的开始时间，默认从1970年开始计算
	node, err = sf.NewNode(machineID)  // 机器编号，最多1024
	return
}

func GenID() int64 {
	return node.Generate().Int64()
}

func GenIDStr() string {
	return node.Generate().String()
}
