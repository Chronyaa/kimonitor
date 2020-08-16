package base

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Chronyaa/kimonitor/util"
)

type resultHandler func(c *Collector, res string)

type collector func(c *Collector) string

var handlerMap = map[string]resultHandler{}

var collectorMap = map[string]collector{}

// Collector 用于收集信息
type Collector struct {
	// Collector名称
	Name string
	// 目标URL
	URL string
	// HTTP方法
	Method string
	// HTTP头部
	Header http.Header
	// HTTP内容（仅POST方法生效）
	Content string
	// 额外的参数
	Args map[string]interface{}
	// HTTP客户端
	Client *http.Client
	// 获得的结果的时间
	LastTime time.Time
	// 到下一次Colloect的间隔，默认为1分钟
	Interval time.Duration
	// 用于收集信息的函数，默认使用defaultCollector
	Collector collector
	// 用于处理结果的回调函数，包括该Collector本身和结果
	Handler resultHandler
	// 用于写入结果的channel
	ResChan chan string
	// Request缓存
	req *http.Request
}

// Init 初始化Collector，返回是否成功初始化
func (c *Collector) Init(
	name string, url string, method string, header http.Header,
	content string, args map[string]interface{}, ch chan string) bool {
	c.Name = name
	c.Header = header
	if c.Header == nil {
		c.Header = make(http.Header)
	}
	c.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36")
	c.URL = url
	if c.URL == "" {
		util.Warning.Println("URL is empty, collector may not work!")
	}
	c.Content = content
	c.Args = args
	if c.Args == nil {
		c.Args = make(map[string]interface{})
	}
	c.Client = &http.Client{}
	c.LastTime = time.Now()
	c.Interval = time.Minute
	c.Collector = collectorMap[name]
	if c.Collector == nil {
		c.Collector = defaultCollector
	}
	c.Handler = handlerMap[name]
	if c.Handler == nil {
		util.Error.Printf("No such handler named %s, collector will not work!", c.Name)
		return false
	}
	c.ResChan = ch
	c.UpdateReq()
	return true
}

// UpdateReq 更新HTTP Request缓存，更新失败则仍然使用之前的Request
func (c *Collector) UpdateReq() {
	var r io.Reader
	if c.Method == "POST" {
		r = strings.NewReader(c.Content)
	}
	req, err := http.NewRequest(c.Method, c.URL, r)
	if err != nil {
		util.Warning.Printf("Update Request Failed. Request is %s[%s].\n", c.Method, c.URL)
	} else {
		c.req = req
	}
	c.req.Header = c.Header
}

// defaultCollector 默认的收集函数，发起HTTP请求并返回其内容
func defaultCollector(c *Collector) string {
	rsp, err := c.Client.Do(c.req)
	res := ""
	if err != nil {
		util.Error.Printf("Collect Failed. Current State:\nMethod: %s\nURL: %s\nHeader: %v\n",
			c.Method, c.URL, c.Header)
	} else {
		content, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			util.Error.Printf("Parse Content Failed. Current State:\nMethod: %s\nURL: %s\nHeader: %v\n",
				c.Method, c.URL, c.Header)
		} else {
			res = string(content)
		}
	}
	return res
}

// Start 开始循环Collect
func (c *Collector) Start() {
	for {
		c.LastTime = time.Now()
		c.Handler(c, c.Collector(c))
		time.Sleep(c.Interval)
	}
}

// RegisterCollectHandler 向handlerMap注册回调函数
func RegisterCollectHandler(s string, h resultHandler) {
	if handlerMap[s] != nil {
		util.Warning.Println("a handler function is already in handlerMap: ", s)
	} else {
		handlerMap[s] = h
	}
}

// RegisterCollectFunc 向collectorMap注册collect函数
func RegisterCollectFunc(s string, c collector) {
	if collectorMap[s] != nil {
		util.Warning.Println("a collector function is already in collectorMap: ", s)
	} else {
		collectorMap[s] = c
	}
}
