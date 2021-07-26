package main

import (
	"goquake2/client"
	"goquake2/common"
	"goquake2/server"
)

func main() {

	cl := client.CreateClient()
	sv := server.CreateServer()
	quake := common.CreateQCommon(cl, sv)
	quake.Init()

}
