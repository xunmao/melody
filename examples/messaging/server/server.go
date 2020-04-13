package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xunmao/melody"
)

const (
	ctxKeyChannel = "channel"
)

func main() {

	router := gin.Default()
	wsServer := melody.New()

	router.GET("/messaging", func(ctx *gin.Context) {
		wsServer.HandleRequest(ctx.Writer, ctx.Request)
	})

	wsServer.HandleMessage(func(session *melody.Session, msg []byte) {

		wsServer.BroadcastFilter(msg, func(otherSession *melody.Session) bool {

			channel := session.Request.URL.Query()[ctxKeyChannel][0]
			otherChannel := otherSession.Request.URL.Query()[ctxKeyChannel][0]

			return (channel == otherChannel) && (session != otherSession)
		})
	})

	router.Run(":8080")
}
