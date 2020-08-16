package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/Chronyaa/kimonitor/display"
	"github.com/Chronyaa/kimonitor/util"

	"github.com/Chronyaa/kimonitor/base"
	_ "github.com/Chronyaa/kimonitor/detail"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	// 读取配置
	var cfgs base.CollectorConfigs
	buf, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalln("Open config.json failed. Please check your config.", err)
	}
	err = json.Unmarshal(buf, &cfgs)
	if err != nil {
		log.Fatalln("Parse config.json failed. Please check your config.", err)
	}
	ch := make(chan string, 16)
	// 根据配置启动Collector
	for i, cfg := range cfgs.Configs {
		fmt.Println("Create collector", i, cfg.Name)
		fmt.Println("From", cfg)
		c := &base.Collector{}
		c.Init(cfg.Name, cfg.URL, cfg.Method, cfg.Header, cfg.Content, cfg.Args, ch)
		go c.Start()
	}
	// 此处应有一个Displayer，不过既然不复杂就不写了
	var displayer io.Writer = os.Stdout
	// 读取配置
	buf, err = ioutil.ReadFile("telegram.json")
	if err != nil {
		log.Println("Open telegram.json failed. Please check your config.", err, "Use os.Stdout to display.")
	} else {
		var config display.TelegramRobotConfig
		err = json.Unmarshal(buf, &config)
		if err != nil {
			log.Println("Parse config.json failed. Please check your config.", err, "Use os.Stdout to display")
		} else {
			displayer = display.NewTelegramRobotDefault(config.Token, config.Group)
		}
	}
	// 从channel中获取信息，然后写入displayer
	for {
		s := <-ch
		_, err := displayer.Write([]byte(s))
		if err != nil {
			util.Warning.Println("Write message failed.", err, "Message:", s)
		}
		// 避免某些点赞狂魔一下子点很多赞导致（可能出现的）请求丢失
		time.Sleep(3 * time.Second)
	}
}
