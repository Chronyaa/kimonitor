package detail

import (
	"strings"

	"github.com/Chronyaa/kimonitor/base"
	"github.com/Chronyaa/kimonitor/util"
)

/* 需要的Args:
 * name 名称，仅用于输出
 */

func init() {
	base.RegisterCollectHandler("bilibili", collectBilibiliHandler)
}

func collectBilibiliHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	i := strings.Index(res, `"live_status":`)
	util.Warning.Assertln(i >= 0, "Can not find live status info. Page: \n", res)
	isLiving := res[i+len(`"live_status":`)] == '1'
	// 初始化
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
			c.ResChan <- name + " 在Bilibili推流开始。"
			c.Args["isLiving"] = true
		}
	} else {
		if prevIsLiving {
			c.ResChan <- name + " 在Bilibili推流结束。"
			c.Args["isLiving"] = false
		}
	}
	util.Debug.Printf("[bilibil][%s].isLiving[%v]\n", name, isLiving)
}
