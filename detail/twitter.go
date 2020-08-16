package detail

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Chronyaa/kimonitor/base"
	"github.com/Chronyaa/kimonitor/util"
)

/* 需要的Args:
 * name 名称，仅用于输出
 */

func init() {
	base.RegisterCollectHandler("twitter.timeline", collectTimelineHandler)
	base.RegisterCollectHandler("twitter.like", collectLikeHandler)
}

// collectTimelineHandler 处理Timeline信息
func collectTimelineHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	// 反序列化结果
	var tweets []Tweet
	err := json.Unmarshal([]byte(res), &tweets)
	util.Warning.Returnln(err, "Parse json error. Page:\n", res)
	if len(tweets) == 0 {
		return
	}
	// 解析结果，加入到timeline，并求出tweet_id范围
	timeline := make(map[int64]string)
	max := int64(0)
	min := int64(^uint64(0) >> 1)
	for _, tweet := range tweets {
		timeline[tweet.ID] = tweet.Text
		if tweet.ID > max {
			max = tweet.ID
		}
		if tweet.ID < min {
			min = tweet.ID
		}
	}
	// 初始化
	if c.Args["timeline"] == nil {
		c.Args["timeline"] = timeline
		c.Args["max"] = max
		return
	}
	// 获取Args
	prevTimeline, ok := c.Args["timeline"].(map[int64]string)
	util.Error.Assertln(ok, `Args["timeline"] is not a map[int64]string. Please check your code.`)
	prevMax, ok := c.Args["max"].(int64)
	util.Error.Assertln(ok, `Args["max"] is not a int64. Please check your code.`)
	name, ok := c.Args["name"].(string)
	util.Warning.PrintlnIfNot(ok, `Args["name"] is not a string. Please check your config.`)
	// 和上一次的timeline比较，多出来的并且ID大于上一次最大ID的就是新的tweet
	for id, text := range timeline {
		if prevTimeline[id] == "" && id > prevMax {
			if strings.HasPrefix(text, `RT @`) {
				c.ResChan <- name + " 转发:\n" + text[3:] + "\n"
			} else if strings.HasPrefix(text, `@`) {
				c.ResChan <- name + " 回复:\n" + text + "\n"
			} else {
				c.ResChan <- name + " 发推:\n" + text + "\n"
			}
		}
	}
	// 和上一次的timeline比较，少掉的而且id大于本次最小ID的就是被移除的tweet
	for id, text := range prevTimeline {
		if timeline[id] == "" && id > min {
			c.ResChan <- name + " 移除:\n" + text + "\n"
		}
	}
	// 更新Args
	c.Args["timeline"] = timeline
	c.Args["max"] = max
	// twitter API频率限制
	c.Interval = time.Minute * 2
	util.Debug.Printf("[twitter.timeline][%s].size[%d]\n", name, len(timeline))
}

// collectLikeHandler 处理Like信息
func collectLikeHandler(c *base.Collector, res string) {
	defer util.Recover()
	if res == "" {
		return
	}
	// 反序列化结果
	var tweets []Like
	err := json.Unmarshal([]byte(res), &tweets)
	util.Warning.Returnln(err, "Parse json error. Page:\n", res)
	if len(tweets) == 0 {
		return
	}
	// 初始化 & 获取Args
	init := true
	var ok bool
	var likes map[int64]string
	if c.Args["likes"] == nil {
		init = false
		likes = make(map[int64]string)
	} else {
		likes, ok = c.Args["likes"].(map[int64]string)
		util.Error.Assertln(ok, `Args["likes"] is not a map[int64]string. Please check your code.`)
	}
	name, ok := c.Args["name"].(string)
	util.Warning.PrintlnIfNot(ok, `Args["name"] is not a string. Please check your config.`)
	// 由于从twitter API获得的Likes列表是以被like的tweet的发布时间排序，而不是like的时间排序
	// 因此只能判断是否加入而不能判断是否移除（而且如果like了一个上古tweet这里也会误判）
	for _, tweet := range tweets {
		if likes[tweet.ID] == "" && init {
			c.ResChan <- name + " 喜欢了: @" + tweet.User.Name + "\n" + tweet.Text + "\n"
		}
		likes[tweet.ID] = tweet.Text
	}
	// twitter API频率限制
	c.Interval = time.Minute * 2
	c.Args["likes"] = likes
	util.Debug.Printf("[twitter.like][%s].size[%d]\n", name, len(likes))
}

// Tweet 一条推特
type Tweet struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

// TwitterUser 一名用户
type TwitterUser struct {
	Name string `json:"screen_name"`
}

// Like 一条喜欢
type Like struct {
	ID   int64       `json:"id"`
	Text string      `json:"text"`
	User TwitterUser `json:"user"`
}
