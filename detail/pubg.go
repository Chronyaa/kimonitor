package detail

import (
	"encoding/json"

	"github.com/Chronyaa/kimonitor/base"
	"github.com/Chronyaa/kimonitor/util"
)

/* 需要的Args:
 * 无
 */

func init() {
	base.RegisterCollectHandler("pubg", collectPUBGHandler)
}

func collectPUBGHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	// 反序列化得到的结果
	r := new(PUBGRecord)
	err := json.Unmarshal([]byte(res), r)
	util.Warning.Returnln(err, "Unmarshal res failed. Page:\n", res)
	if len(r.Matches.Items) == 0 {
		return
	}
	// 获取最近一次的游戏ID
	recent := r.Matches.Items[0].MatchID
	// 初始化
	if c.Args["recent"] == nil {
		c.Args["recent"] = recent
		return
	}
	// 获取Args
	prevRecent, ok := c.Args["recent"].(string)
	util.Error.Assertln(ok, `Args["recent"] is not a string. Please check your code.`)
	// 和上一次的状态比较
	if prevRecent != recent {
		c.ResChan <- r.Matches.Items[0].Participant.User.Nickname + "结束了一场PUBG游戏.\n"
		c.Args["recent"] = recent
	}
	util.Debug.Printf("[pubg][%s].recent[%s]\n", r.Matches.Items[0].Participant.User.Nickname, recent)
}

// PUBGRecord autogen
type PUBGRecord struct {
	Matches Matches `json:"matches"`
}

// Matches autogen
type Matches struct {
	Items []PUBGItem `json:"items"`
}

// PUBGItem autogen
type PUBGItem struct {
	MatchID     string      `json:"match_id"`
	StartedAt   string      `json:"started_at"`
	UpdatedAt   string      `json:"updated_at"`
	Participant Participant `json:"participant"`
}

// Participant autogen
type Participant struct {
	User User `json:"user"`
}

// User autogen
type User struct {
	Nickname string `json:"nickname"`
}
