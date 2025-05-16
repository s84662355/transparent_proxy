package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync/atomic"
)

var confDataInstancePointer atomic.Pointer[confData]

func GetConf() *confData {
	return confDataInstancePointer.Load()
}

func Load(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("加载配置文件失败 config file: %s \n", err))
	}

	// 解析 JSON 数据
	var confDataT confData
	err = json.Unmarshal(data, &confDataT)
	if err != nil {
		panic(fmt.Errorf("解释配置文件失败 config file: %s \n", err))
	}
	confDataInstancePointer.Store(&confDataT)
}
