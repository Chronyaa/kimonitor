# kimonitor

キーーモーー

**请通过公开、合法来源获取信息，本程序仅从用户提供的信息来源及个别公开的信息来源（如livechat.go中的hiyoko、youtube）获取信息并进行处理。使用本程序产生的一切后果作者都不对此负责。**

**请遵循AGPLv3协议。**

## 简介

本程序一共由三部分组成：`Collector`, `Handler`, `Displayer`.

### Collector

`Collector`负责收集信息，并将信息传给`Handler`。

默认的collect函数`defaultCollector()`的功能就是发送一个HTTP请求到指定URL，然后返回得到的结果。

该部分可以根据需求定制，最后在`init()`函数中通过`RegisterCollectFunc()`注册即可。如果没有指定，则会使用默认collect函数。

### Handler

`Handler`负责处理信息，并将处理后的信息传给`Displayer`。所有`Handler`在`detail/`目录下。

该部分可以根据需求定制，最后在`init()`函数中通过`RegisterCollectHandler()`注册即可。如果没有指定，则无法正常工作。

### Displayer

`Displayer`负责打印处理后的信息，一般来说它会在主线程工作，通过一个channel接收其它协程的传来的信息，并以合适的方式打印。

目前只提供了telegram机器人作为Displayer，要使用它，需要先获取一个robot，然后将robot加入群组，获得群组id，最后将token和group_id写入`telegram.json`（通过复制`template.telegram.json`）中。如果没有配置，默认使用stdout作为输出。

## 配置

通过`config.json`进行配置，一个例子如下：

```js
{
  // 由多个config组成
  "configs":[
    {
      // 该名称必须和在init()函数，RegisterCollectHandler/RegisterCollectFunc中用于注册的名称相同
      // Collector初始化时会根据该名称找到对应的Handler/CollectFunc
      "name": "youtube",
      // URL即是要获取信息的地址，一般来说可能是直播间地址，或者一个API调用
      "URL": "https://www.youtube.com/channel/${channel_id}/live",
      // HTTP请求的方法，一般来说都是GET，当然某些API调用可能需要POST方法
      "method": "GET",
      // HTTP请求头部，一般来说可以为空，但是某些网站需要
      // 它是一个map[string][]string，因为HTTP头部中，一个key可以对应多个value
      "header": {},
      // 使用POST请求时的请求体，使用GET请求时无效
      "content": "",
      // 该参数用于向Handler传递参数，绝大多数情况下都只需要name，以便于输出。
      // 在detail/*.go的开头标注有使用该Handler需要什么参数
      // 它是一个map[string]interface{}，此处可以写任意类型，在Handler中则需要转换成对应的类型。
      "args": {
        "name": "${name}"
      }
    },
    {
      "name": "bilibili",
      "method": "GET",
      "url": "https://live.bilibili.com/${live_id}",
      // 对于bilibili直播，它需要一个Accept字段才能正常获取信息
      "header": {
        "Accept": [
          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3"
        ]
      },
      "content": "",
      "args": {
        "name": "${name}"
      }
    },
    {
      "name": "twitter.timeline",
      "method": "GET",
      "url": "https://api.twitter.com/1.1/statuses/user_timeline.json/screen_name=${name}",
      // 对于twitter API，它需要一个Bearer Token用于认证，如何获取请参考Twitter Developer
      "header": {
        "Authorization": [
          "Bearer ${your_bearer_token}"
        ]
      },
      "content": "",
      "args": {
        "name": "${name}"
      }
    }
  ]
}
```

提供了一个模板`template.config.json`，将其复制为`config.json`然后修改其中的`${args}`即可。

### Twitter API Bearer Token

通过`https://developer.twitter.com/`申请得到Consumer Key和Consumer Secret后，可以使用`tools/twitter_gen_token/`下的工具获取Bearer Token:
```
./twitter_gen_token ${consumerKey} ${consumerSecret}
Bearer XXXXYYYYYYYYYYZZZZZZZZZZZ
```

将打印出来的`Bearer XXXXYYYYYYYYYYZZZZZZZZZZZ`填入配置文件即可。

## 扩展

如上所述，需要添加功能很简单，例如需要添加xxx平台，在detail/下创建xxx.go，然后按以下模式编写：

```go
func init() {
  // 如果没有特殊需求，下面这行可以不要，使用默认collect函数即可
  // base.RegisterCollectFunc("xxx", collectXXXFunc)
  base.RegisterCollectHandler("xxx", collectXXXHandler)
}

func collectXXXFunc(c *base.Collector) string {
  // Collector中几乎所有成员都是公开的，可以根据需要改变
  // do something...
  return result
}

func collectXXXHandler(c *base.Collector, res string) {
  // res就是通过collect得到的信息
  // 也许需要改变URL，或者间隔，HTTP方法、头部……
  // c.URL = "new url"
  last := doSomething(res)
  // 通过c.Args[key]来储存上一次的状态，当然可以储存不只一个参数
  // 当然，第一次调用的时候不会有上一次的状态，此处进行初始化
  if c.Args["last"] == nil {
    c.Args["last]" = last
    return
  }
  // 通过转换得到上一次的状态，或者其他储存的参数
  prevLast, ok := c.Args["last"].(string)
  // 比较两次状态...
  result := compare(last, prevLast)
  // 向channel发送处理后的结果
  c.ResChan <- c.Args["name"].(string) + result
  // 可以打印DEBUG日志，在util/log.go中将debugFile设置为os.DevNull就可以关闭该等级的输出
  util.Debug.Printf("[xxx][%s].last[%s]\n", name, last)
}
```

然后在`config.json`中加上对应的配置就可以了。
