package main

import (
	"github.com/gogf/gf/frame/g"
	_ "jingyu/boot"
)

func main() {
	g.Server(g.Config().GetString("client-name")).Start()
	g.Wait()
}