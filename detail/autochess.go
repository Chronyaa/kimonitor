package detail

import (
	"regexp"

	"github.com/Chronyaa/kimonitor/base"
	"github.com/Chronyaa/kimonitor/util"
)

/* 需要的Args:
 * name 名称，仅用于输出
 */

func init() {
	base.RegisterCollectHandler("autochess", collectAutochessHandler)
}

func collectAutochessHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	// 寻找最近一次的结束的key
	r := regexp.MustCompile(`"match_key":"([^"]+)"`)
	s := r.FindString(res)
	util.Warning.Assertln(len(s) > 0, "Search autochess game finish time failed. Page:\n", res)
	// 初始化
	if c.Args["recent"] == nil {
		c.Args["recent"] = s
		return
	}
	// 获取Args
	prevRecent, ok := c.Args["recent"].(string)
	util.Error.Assertln(ok, `Args["recent"] is not a string. Please check your code.`)
	name, ok := c.Args["name"].(string)
	util.Warning.PrintlnIfNot(ok, `Args["name"] is not a string. Please check your config.`)
	// 和上一次的key比较
	if prevRecent != s {
		c.Args["recent"] = s
		c.ResChan <- name + " 结束了一场自走棋.\n"
	}
	util.Debug.Printf("[autochess][%s].recent[%s]\n", name, s)
}
