package detail

import (
	"strings"

	"github.com/Chronyaa/kimonitor/base"
	"github.com/Chronyaa/kimonitor/util"
)

/* 需要的Args:
 * name string类型 名称，仅用于输出
 */

func init() {
	base.RegisterCollectHandler("lol", collectLoLHandler)
}

func collectLoLHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	isPlaying := strings.Contains(res, "Green")
	// 初始化
	if c.Args["isPlaying"] == nil {
		c.Args["isPlaying"] = isPlaying
		return
	}
	// 获取Args
	prevIsPlaying, ok := c.Args["isPlaying"].(bool)
	util.Error.Assertln(ok, `Args["isPlaying"] is not a bool. Please check your code.`)
	name, ok := c.Args["name"].(string)
	util.Warning.PrintlnIfNot(ok, `Args["name"] is not a string. Please check your config.`)
	// 与上一次的状态比较
	if isPlaying {
		if !prevIsPlaying {
			c.ResChan <- name + "开始了一场LOL游戏.\n"
			c.Args["isPlaying"] = true
		}
	} else {
		if prevIsPlaying {
			c.ResChan <- name + "结束了一场LOL游戏.\n"
			c.Args["isPlaying"] = false
		}
	}
	util.Debug.Printf("[lol][%s].isPlaying[%v]\n", name, isPlaying)
}
