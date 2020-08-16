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
	base.RegisterCollectHandler("youtube", collectYoutubeHandler)
}

func collectYoutubeHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	isLiving := !strings.Contains(res, "LIVE_STREAM_OFFLINE")
	// isValid 指示isLiving是否有效，因为youtube可能会返回一个不包含LIVE_STREAM_OFFLINE和LIVE_INDEX的页面
	isValid := (!isLiving) || (isLiving && strings.Contains(res, "LIVE_INDEX"))
	if !isValid {
		return
	}
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
	// 判断当前状态是否与前一次相异
	if isLiving {
		if !prevIsLiving {
			c.ResChan <- name + " 在Youtube推流开始。"
			c.Args["isLiving"] = true
		}
	} else {
		if prevIsLiving {
			c.ResChan <- name + " 在Youtube推流结束。"
			c.Args["isLiving"] = false
		}
	}
	util.Debug.Printf("[youtube][%s].isLiving[%v]\n", name, isLiving)
}
