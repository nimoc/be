# 守护进程

> **守护进程**\(daemon\)是一类在后台运行的特殊**进程**，用于执行特定的系统任务。 很多**守护进程**在系统引导的时候启动，并且一直运行直到系统关闭。 另一些只在需要的时候才启动，完成任务后就自动结束。

{% hint style="info" %}
部署必须涉及具体的编程语言，为了便于演示本节选择易于部署的 go 语言。你也可以换成你熟悉的语言。
{% endhint %}

{% code title="main.go" %}
```go
/*
    你不需要看懂 main.go 文件， 你只需要知道程序启动后
    访问 http://127.0.0.1:1111/ 会返回当前时间
    访问 http://127.0.0.1:1111/exit 会让程序退出
*/
package main

import (
    "log"
    "net/http"
    "time"
    "os"
)

func main() {
    http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
        _, err := writer.Write([]byte(time.Now().String())) ; if err != nil {
            log.Print(err)
            writer.WriteHeader(500)
        }
    })
    http.HandleFunc("/exit", func(writer http.ResponseWriter, request *http.Request) {
        os.Exit(0)
    })
    addr := ":1111"
    log.Print("http://127.0.0.1" + addr)
    log.Print(http.ListenAndServe(addr, nil))
}
```
{% endcode %}

运行 `go build main.go` 编译生成 `./main` 文件。

接下来只需要在终端运行`./main` 即可启动服务。

{% hint style="info" %}
当http端口被占用导致启动失败可以通过 lsof -i:端口 查看 PID 然后使用 kill+PID关闭序。
{% endhint %}

通过 `./main` 运行程序存在的问题有

1. 一旦终端关闭则程序自动退出
2. 访问[http://127.0.0.1:1111/exit](http://127.0.0.1:1111/exit) 之后再次访问 [http://127.0.0.1:1111/](http://127.0.0.1:1111/) 会发现服务关闭了，因为 /exit 让程序退出了。

当我们部署一个网站时是希望程序能“后台运行”的，当我们退出正式的服务器环境时程序依然在运行。并且因为疏忽写了一些代码导致程序退出时不应该整个网站关闭，而是程序能自动重启。

## pm2 <a id="pm2"></a>

最常见的进程守护工具是使用Python开发的 [supervisor](https://cn.bing.com/search?q=supervisor)，相对更容易上手的还有 [pm2](https://cn.bing.com/search?q=pm2).

安装 [Node](https://nodejs.org/)

```bash
curl --silent --location https://rpm.nodesource.com/setup_14.x | bash -
sudo yum install -y nodejs
node -v
```

安装 [pm2](https://pm2.keymetrics.io/)

```bash
npm install pm2@latest -g
```

使用 pm2 启动程序

```bash
pm2 list
pm2 start ./main --name=demo
pm2 log demo
pm2 restart demo
pm2 stop demo
pm2 delete demo
pm2 monit
```

{% hint style="info" %}
不要使用 pm2 restart 数字 因为这样容易误操作，尽量使用 pm2 start name。
{% endhint %}

使用 pm2 start 之后即使关闭终端，程序依然在启动状态

部署属于服务器运维知识，服务器运维的一个好习惯是考虑当服务器宕机重启后服务是否依然正常。

如果你是在测试服务器上试验请尝试重启服务器，如果是在电脑上试验请重启电脑。 重启后使用 pm2 list 命令查看服务状态时会发现没有任何服务在运行。

可以使用startup/save设置开启自动使用pm2启动程序。

```bash
pm2 startup # 只需要运行一次

pm2 save # 每次启动新进程都要运行
```

还记得刚才访问 [http://127.0.0.1:1111/exit](http://127.0.0.1:1111/exit) 会导致服务退出吗？

使用 pm2 启动的服务，在程序退出后会自动重启。你可以使用 pm2 启动服务后反复访问 / 和 /exit 试试。 然后不使用pm2 直接运行 ./main 后反复访问 / 和 /exit 试试。

## 安全退出 <a id="safe-exit"></a>

很多人喜欢叫优雅退出，我觉得安全退出更为合适。

没有什么服务是只启动一次永不停止的。当发布新版本时服务总是会终止。如果没有监听程序退出信号会导致一些“善后工作没有被处理”。

由于会涉及到具体的编程语言不再深入，读者可以自行使用搜索引擎搜索自己使用的编程语言如何实现安全退出/优雅退出.

## kubernetes <a id="k8s"></a>

pm2 supervisor 等项目只适合简单项目,复杂项目需要将同一份代码部署在多台机器,并且需要监控这些应用的状态.同时要确保运行环境一致. 为了做到这一条可以使用 kubernetes

{% page-ref page="k8s.md" %}

欢迎留言讨论： [https://github.com/nimoc/be/discussions/2](https://github.com/nimoc/be/discussions/2)

