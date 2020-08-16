package detail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Chronyaa/kimonitor/base"
	"github.com/Chronyaa/kimonitor/util"
)

/* 需要的Args:
 * keyword 要寻找的关键词列表,[]string类型
 * sender 要寻找的用户，[]string类型
 */

/*
 *
 */

/* TODO 该功能是实验版本，不稳定
 * 具体来说，只要直播间有待机页面，能获取continuation信息，那么该直播间就不会从continuationMap中删除。
 * 从而导致continuationMap不断增大，会消耗更多的内存和流量。
 * 尽管能获得待机页面的livechat，并且有上限（2000~3000左右），但实际上并不值得去获取没有太大价值的待机页面的livechat。
 */

// Deprecated: Youtube API已经改变，莫得精力去改了
func init() {
	base.RegisterCollectFunc("livechat", collectChannelInfo)
	base.RegisterCollectHandler("livechat", collectLiveChatHandler)
}
func collectChannelInfo(c *base.Collector) string {
	defer util.Recover()
	req, err := http.NewRequest(
		"POST",
		"https://hiyoko.sonoj.net/f/avtapi/live/fetch_current_v2", nil)
	util.Error.Returnln(err, "Create HTTP request failed.", err)
	req.Header.Add("Origin", "https://hiyoko.sonoj.net")
	req.Header.Add("Referer", "https://hiyoko.sonoj.net/")
	req.Header.Add("Accpet", "application/json;charset=UTF-8")
	req.Header.Add("Content-Type", "application/json")
	rsp, err := c.Client.Do(req)
	util.Error.Returnln(err, "POST Request Error: ", err)
	buf, err := ioutil.ReadAll(rsp.Body)
	util.Error.Returnln(err, "Read all failed.", err)
	util.Debug.Println("[hiyoko]collect channel info finished.")
	return string(buf)
}

func collectLiveChatHandler(c *base.Collector, res string) {
	defer util.Recover()
	var infos ChannelInfos
	// 反序列化channel信息
	err := json.Unmarshal([]byte(res), &infos)
	util.Error.Returnln(err, "Parse json error. Please check your code.", err)
	// 从Args获取continuationMap，保存从频道信息到continuation的映射
	if c.Args["continuationMap"] == nil {
		c.Args["continuationMap"] = make(map[string]ChannelInfo)
	}
	continuationMap, ok := c.Args["continuationMap"].(map[string]ChannelInfo)
	util.Error.Assertln(ok, `Args["continuationMap"] is not a map[string]ChannelInfo. Please check your code.`)
	keyword := loadStringSliceFromMap(c.Args, "keyword")
	sender := loadStringSliceFromMap(c.Args, "sender")
	// 同步所有获取continuation信息的goroutine
	var wg sync.WaitGroup
	// 保护map
	var mtx sync.Mutex
	// 记录当前的channel信息，用于和上一次的比较，并删除已经不在直播的channel
	cur := map[string]ChannelInfo{}
	// 遍历所有channel信息
	for _, info := range infos.ChannelInfo {
		// 非youtube频道，暂不支持
		if info.ChType != 1 {
			continue
		}
		info := info
		cur[info.ChID] = info
		wg.Add(1)
		// 如果之前不存在该continuation，先通过直播间页面获取
		// 如果存在，则跳过该步骤
		go func() {
			defer wg.Done()
			// continuation信息为空，实际上就是ChannelInfo整个为空
			// 因为要被加入到continuationMap，必须先获取一次continuation信息
			// 如果获取失败，则被移除，更新continuation信息失败时也会被移除
			if continuationMap[info.ChID].continuation == "" {
				info.continuation = getFirstContinuationString(c, "https://www.youtube.com/channel/"+info.ChID+"/live")
				// 上锁保护map
				mtx.Lock()
				defer mtx.Unlock()
				// 不存在，删除该channel的信息，避免下一次再寻找到该channel对应的continuation
				if info.continuation == "" {
					delete(continuationMap, info.ChID)
				} else {
					continuationMap[info.ChID] = info
				}
			}
		}()
	}
	wg.Wait()
	// 遍历所有channel的continuation，获取livechat内容，并查找
	for k, v := range continuationMap {
		k, v := k, v
		wg.Add(1)
		go func() {
			defer wg.Done()
			messages, err := getLiveChat(c, &v.continuation)
			// 上锁保护map
			mtx.Lock()
			defer mtx.Unlock()
			// 更新continuation信息
			// 如果已经没有continuation信息，或者出现错误，或已经不在直播列表中，则从continuationMap中删除该信息
			if v.continuation == "" || err != nil || cur[k].ChID == "" {
				delete(continuationMap, k)
				return
			}
			continuationMap[k] = v
			// 遍历所有message并查找发送者中有无sender或文本中有无keyword
			for _, message := range messages {
				for _, s := range keyword {
					if strings.Contains(message.Message, s) {
						c.ResChan <- fmt.Sprintln("频道:", v.Name, "\n用户:", message.AuthorName, "\n评论:", message.Message)
					}
				}
				for _, s := range sender {
					if strings.Contains(message.AuthorName, s) {
						c.ResChan <- fmt.Sprintln("频道:", v.Name, "\n用户:", message.AuthorName, "\n评论:", message.Message)
					}
				}
			}
		}()
	}
	wg.Wait()
	c.Interval = time.Second * 30
	c.Args["continuationMap"] = continuationMap
	util.Debug.Printf("[livechat]continuationMap.size[%d]\n", len(continuationMap))
}

// 首次获取continuation信息
// 包括两个阶段：
// 	1.获取直播间页面，提取出livechat页面的信息
//	2.获取livechat页面，提取出continuation的信息
func getFirstContinuationString(c *base.Collector, liveURL string) string {
	defer util.Recover()
	const pattern string = "{\"reloadContinuationData\":{\"continuation\":\""
	// 获取直播间网页
	req, err := http.NewRequest("GET", liveURL, nil)
	util.Error.Returnln(err, "Create request failed.", err, liveURL)
	rsp, err := c.Client.Do(req)
	util.Error.Returnln(err, "Get live page failed.", err, liveURL)
	buf, err := ioutil.ReadAll(rsp.Body)
	util.Error.Returnln(err, "Read all failed.", err, liveURL)
	// 找到livechat页面的URL
	reg := regexp.MustCompile(`src="(https://www.youtube.com/live_chat[^"]+)"`)
	buf = reg.Find(buf)
	// 停止了聊天功能的直播间，或者没有待机页面的频道是找不到livechat页面的，不报错直接返回
	if len(buf) == 0 {
		return ""
	}
	buf = reg.ReplaceAll(buf, []byte("$1"))
	util.Warning.Assertln(len(buf) > 0, "Search Livechat Page Information Error", liveURL)
	livechatURL := string(bytes.Replace(buf, []byte(`%3D`), []byte(`%253D`), -1))
	// 获取continuation信息
	req, err = http.NewRequest("GET", livechatURL, nil)
	util.Error.Returnln(err, "Create request failed.", err, livechatURL)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36")
	req.Header.Add("Host", "www.youtube.com")
	req.Header.Add("Accpet", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accpet-Encoding", "defalte")
	rsp, err = c.Client.Do(req)
	util.Error.Returnln(err, "Get livechat page failed.", err, livechatURL)
	buf, err = ioutil.ReadAll(rsp.Body)
	util.Error.Returnln(err, "Read all failed.", err, livechatURL)
	begin := bytes.Index(buf, []byte(pattern))
	util.Warning.Assertln(begin >= 0, "Search Continuation Information Error", livechatURL)
	begin += len(pattern)
	end := bytes.Index(buf[begin:], []byte("\""))
	util.Warning.Assertln(begin >= 0, "Search Continuation Information Error", livechatURL)
	return strings.Replace(string(buf[begin:begin+end]), "%3D", "%253D", -1)
}

// 获取一段livechat信息
func getLiveChat(c *base.Collector, continuation *string) (result []Comment, err error) {
	defer util.Recover()
	liveChatURL := "https://www.youtube.com/live_chat/get_live_chat?continuation=%s&hidden=false&pbj=1"
	URL := fmt.Sprintf(liveChatURL, *continuation)
	req, err := http.NewRequest("GET", URL, nil)
	util.Error.Returnln(err, "Create request failed.", err, URL)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/76.0.3809.100 Safari/537.36")
	req.Header.Add("Host", "www.youtube.com")
	req.Header.Add("Accpet", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accpet-Encoding", "defalte")
	// 发送获取LiveChat的请求
	rsp, err := c.Client.Do(req)
	util.Error.Returnln(err, "Get livechat failed.", err, URL)
	buf, err := ioutil.ReadAll(rsp.Body)
	util.Error.Returnln(err, "Read all failed.", err, URL)
	// 解析json
	var lc LiveChat
	err = json.Unmarshal(buf, &lc)
	util.Warning.Checkln(err, "Parse json error.", err, URL)
	// 获取评论信息
	for _, action := range lc.Response.ContinuationContents.LiveChatContinuation.Actions {
		if render := action.AddChatItemAction.Item.LiveChatTextMessageRenderer; render != nil {
			result = append(
				result,
				Comment{
					AuthorName: render.AuthorName.SimpleText,
					Message:    mergeMessage(render.Message),
					Purchase:   "",
					AuthorID:   render.ID,
					Timestamp:  render.TimestampUsec,
				},
			)
		} else if render := action.AddChatItemAction.Item.LiveChatPaidMessageRenderer; render != nil {
			result = append(
				result,
				Comment{
					AuthorName: render.AuthorName.SimpleText,
					Message:    mergeMessage(render.Message),
					Purchase:   render.PurchaseAmountText.SimpleText,
					AuthorID:   render.ID,
					Timestamp:  render.TimestampUsec,
				},
			)
		}
	}
	// 更新Continuation信息
	continuations := lc.Response.ContinuationContents.LiveChatContinuation.Continuations
	if len(continuations) > 0 {
		c := continuations[0]
		if c.TimedContinuationData != nil {
			*continuation = strings.Replace(c.TimedContinuationData.Continuation, "%3D", "%253D", -1)
		} else if c.InvalidationContinuationData != nil {
			*continuation = strings.Replace(c.InvalidationContinuationData.Continuation, "%3D", "%253D", -1)
		} else {
			*continuation = ""
		}
	}
	return result, nil
}

// 合并所有信息（由纯文本和emoji合并）
func mergeMessage(m Message) string {
	res := ""
	if r := m.Runs; r != nil {
		for _, s := range *r {
			if s.Text != nil {
				res += *s.Text
			}
			if s.Emoji != nil {
				for _, sh := range s.Emoji.ShortCuts {
					res += sh
				}
			}
		}
	}
	if s := m.Text; s != nil {
		res += *s
	}
	return res
}

func loadStringSliceFromMap(m map[string]interface{}, key string) (arr []string) {
	defer util.Recover()
	// 从Args获取keyword信息，第一次需要从[]interface{}转成[]string
	arr, ok := m[key].([]string)
	if !ok {
		temp, ok := m[key].([]interface{})
		if ok {
			for _, i := range temp {
				s, ok := i.(string)
				if ok {
					arr = append(arr, s)
				} else {
					break
				}
			}
		}
		util.Error.Assertln(ok, `m["`+key+`"] is not a []string. Please check your code.`)
	}
	return arr
}

// ChannelInfo autogen
type ChannelInfo struct {
	ChID   string `json:"ch_id"`
	ChType int64  `json:"ch_type"`
	Title  string `json:"title"`
	Name   string `json:"name"`
	// 记录continuation信息，不需要序列化和反序列化
	continuation string
}

// ChannelInfos autogen
type ChannelInfos struct {
	ChannelInfo []ChannelInfo `json:"current_live"`
}

// LiveChatMessage autogen
type LiveChatMessage struct {
	Sender   string `json:"sender"`
	Message  string `json:"message"`
	Purchase string `json:"purchase"`
}

// LiveChatResponse autogen
type LiveChatResponse struct {
	Continuation string            `json:"continuation"`
	Messages     []LiveChatMessage `json:"messages"`
}

// Comment autogen
type Comment struct {
	AuthorName string `json:"author_name"`
	Purchase   string `json:"purchase"`
	Message    string `json:"message"`
	AuthorID   string `json:"author_id"`
	Timestamp  string `json:"timestamp"`
}

// LiveChat autogen
type LiveChat struct {
	Response Response `json:"response"`
}

// Response autogen
type Response struct {
	ContinuationContents ContinuationContents `json:"continuationContents"`
}

// InvalidationContinuationData autogen
type InvalidationContinuationData struct {
	Continuation string `json:"continuation"`
}

// ContinuationContents autogen
type ContinuationContents struct {
	LiveChatContinuation LiveChatContinuation `json:"liveChatContinuation"`
}

// LiveChatContinuation autogen
type LiveChatContinuation struct {
	Continuations []Continuation `json:"continuations"`
	Actions       []Action       `json:"actions"`
}

// Action autogen
type Action struct {
	AddChatItemAction AddChatItemAction `json:"addChatItemAction"`
}

// AddChatItemAction autogen
type AddChatItemAction struct {
	Item Item `json:"item"`
}

// Item autogen
type Item struct {
	LiveChatTextMessageRenderer *LiveChatTextMessageRenderer `json:"liveChatTextMessageRenderer,omitempty"`
	LiveChatPaidMessageRenderer *LiveChatPaidMessageRenderer `json:"liveChatPaidMessageRenderer,omitempty"`
}

// Message autogen
type Message struct {
	Runs *[]Run  `json:"runs"`
	Text *string `json:"simpleText"`
}

// Emoji autogen
type Emoji struct {
	ShortCuts []string `json:"shortcuts"`
}

// Run autogen
type Run struct {
	Text  *string `json:"text"`
	Emoji *Emoji  `json:"emoji"`
}

// LiveChatPaidMessageRenderer autogen
type LiveChatPaidMessageRenderer struct {
	AuthorName         Title   `json:"authorName"`
	PurchaseAmountText Title   `json:"purchaseAmountText"`
	Message            Message `json:"message"`
	ID                 string  `json:"authorExternalChannelId"`
	TimestampUsec      string  `json:"timestampUsec"`
}

// Title autogen
type Title struct {
	SimpleText string `json:"simpleText"`
}

// LiveChatTextMessageRenderer autogen
type LiveChatTextMessageRenderer struct {
	Message       Message `json:"message"`
	AuthorName    Title   `json:"authorName"`
	ID            string  `json:"authorExternalChannelId"`
	TimestampUsec string  `json:"timestampUsec"`
}

// Continuation autogen
type Continuation struct {
	TimedContinuationData        *TimedContinuationData        `json:"timedContinuationData"`
	InvalidationContinuationData *InvalidationContinuationData `json:"invalidationContinuationData"`
}

// TimedContinuationData autogen
type TimedContinuationData struct {
	Continuation string `json:"continuation"`
}
