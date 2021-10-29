
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
