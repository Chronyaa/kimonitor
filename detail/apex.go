package detail

import (
	"regexp"

	"github.com/Chronyaa/kimonitor/base"
	"github.com/Chronyaa/kimonitor/util"
)

/* TODO:
 * 收集到的数据不稳定，暂时关闭warning输出
 * 考虑更换数据源
 */

/* 需要的Args:
 * name 名称，仅用于输出
 */

func init() {
	base.RegisterCollectHandler("apex", collectApexHandler)
}

func collectApexHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	// 寻找最近一次的结束时间
	r := regexp.MustCompile(`\d{1,2}/\d{1,2}/\d{4} \d{1,2}:\d{2}:\d{2} [AP]M`)
	s := r.FindString(res)
	util.Warning.Assertln(len(s) > 0, "Search apex game finish time failed")
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
	// 和上一次结束时间比较
	if prevRecent != s {
		c.Args["recent"] = s
		c.ResChan <- name + " 结束了一场APEX游戏.\n"
	}
	util.Debug.Printf("[apex][%s].recent[%s]\n", name, s)
}
