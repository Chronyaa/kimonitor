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
	base.RegisterCollectHandler("twitcasting", collectTwitcastingHandler)
}

func collectTwitcastingHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	// 初始化，不打印信息
	isLiving := !strings.Contains(res, `<span class="tw-player-page__live-status--offline">OFFLINE</span>`)
	if c.Args["isLiving"] == nil {
		c.Args["isLiving"] = isLiving
		return
	}
	// 获取Args
	prevIsLiving, ok := c.Args["isLiving"].(bool)
	util.Error.Assertln(ok, `Args["isLiving"] is not a bool. Please check your code.`)
	name, ok := c.Args["name"].(string)
	util.Warning.PrintlnIfNot(ok, `Args["name"] is not a string. Please check your config.`)
	// 和上一次的状态比较
	if isLiving {
		if !prevIsLiving {
			c.ResChan <- name + " 在Twicasting推流开始。"
			c.Args["isLiving"] = true
		}
	} else {
		if prevIsLiving {
			c.ResChan <- name + " 在Twicasting推流结束。"
			c.Args["isLiving"] = false
		}
	}
	util.Debug.Printf("[twicasting][%s].isLiving[%v]\n", name, isLiving)
}
