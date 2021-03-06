package main

import (
	"html"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

func index(c *gin.Context) {
	c.Redirect(301, "/room/hn")
}

func roomGET(c *gin.Context) {
	roomid := c.ParamValue("roomid")
	nick := c.FormValue("nick")
	if len(nick) < 2 {
		nick = ""
	}
	if len(nick) > 13 {
		nick = nick[0:12] + "..."
	}
	c.HTML(200, "room_login.templ.html", gin.H{
		"roomid":    roomid,
		"nick":      nick,
		"timestamp": time.Now().Unix(),
	})

}

func roomPOST(c *gin.Context) {
	roomid := c.ParamValue("roomid")
	nick := c.FormValue("nick")
	message := c.PostFormValue("message")

	validMessage := len(message) > 1 && len(message) < 200
	validNick := len(nick) > 1 && len(nick) < 14
	if !validMessage || !validNick {
		c.JSON(400, gin.H{
			"status": "failed",
			"error":  "the message or nickname is too long",
		})
		return
	}

	post := gin.H{
		"nick":    html.EscapeString(nick),
		"message": html.EscapeString(message),
	}
	messages.Add("inbound", 1)
	room(roomid).Submit(post)
	c.JSON(200, post)
}

func streamRoom(c *gin.Context) {
	roomid := c.ParamValue("roomid")
	listener := openListener(roomid)
	ticker := time.NewTicker(1 * time.Second)
	defer closeListener(roomid, listener)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg := <-listener:
			messages.Add("outbound", 1)
			c.SSEvent("message", msg)
		case <-ticker.C:
			c.SSEvent("stats", Stats())
		}
		return true
	})
}
