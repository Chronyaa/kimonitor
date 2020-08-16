package detail

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Chronyaa/kimonitor/base"
	"github.com/Chronyaa/kimonitor/util"
)

/* 需要的Args:
 * name 名称，仅用于输出
 */

func init() {
	base.RegisterCollectHandler("amazon", collectAmazonHandler)
}
// TODO 关掉愿望单的会让程序不断请求最后OOM
func collectAmazonHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	// 循环获取show more的URL并发起HTTP请求
	moreReg := regexp.MustCompile(`"showMoreUrl":"([^"]+)"`)
	all := res
	for {
		moreURL := moreReg.ReplaceAllString(moreReg.FindString(res), "$1")
		if strings.Index(moreURL, "lek=&") != -1 {
			break
		}
		moreURL = "https://www.amazon.co.jp" + moreURL
		req, err := http.NewRequest("GET", moreURL, nil)
		if err != nil {
			util.Error.Println("Create request failed: ", moreURL, err)
			break
		}
		rsp, err := c.Client.Do(req)
		if err != nil {
			util.Error.Println("Get more failed: ", moreURL, err)
			break
		}
		buf, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			util.Error.Println("Read all failed.", err)
			break
		}
		all += string(buf)
	}
	// 已经获取全部内容，开始解析，提取名称和价格
	giftPattern := `<a class="a-link-normal" title="([^"]*)"`
	pricePattern := `class="a-offscreen">(￥[0-9]+,{0,1}[0-9]+)<`
	giftReg := regexp.MustCompile(giftPattern)
	priceReg := regexp.MustCompile(pricePattern)
	gifts := giftReg.FindAllString(all, -1)
	prices := priceReg.FindAllString(all, -1)
	// 建立映射关系
	for i := range gifts {
		gifts[i] = giftReg.ReplaceAllString(gifts[i], "$1")
		prices[i] = priceReg.ReplaceAllString(prices[i], "$1")
	}
	giftMap := make(map[string]string)
	for i := range gifts {
		giftMap[gifts[i]] = prices[i]
	}
	// 初始化
	if c.Args["giftMap"] == nil {
		c.Args["giftMap"] = giftMap
		return
	}
	// 从Args获取参数
	prevGiftMap, ok := c.Args["giftMap"].(map[string]string)
	util.Error.Assertln(ok, `Args["giftMap"] is not a map[string]string. Please check your code.`)
	name, ok := c.Args["name"].(string)
	util.Warning.PrintlnIfNot(ok, `Args["name"] is not a string. Please check your config.`)
	// 和之前一次的比较
	for k, v := range giftMap {
		if prevGiftMap[k] == "" {
			c.ResChan <- name + fmt.Sprintln("添加了礼物:", k, "价格:", v)
		}
	}
	for k, v := range prevGiftMap {
		if giftMap[k] == "" {
			c.ResChan <- name + fmt.Sprintln("移除了礼物:", k, "价格:", v)
		}
	}
	// 更新Args
	c.Args["giftMap"] = giftMap
	util.Debug.Printf("Current gift list:\n%v", giftMap)
}
