package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 构造函数：创建用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}
	//直接启动监听channel
	go user.ListenMessage()

	return user
}

// 监听当前User channel的go方法,一旦有消息就直接发客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C //从channel中读出来

		this.conn.Write([]byte(msg + "\n")) //发给客户端
	}
}

// 用户上线
func (this *User) Online() {

	//将当前用户加入OnlineMap中
	this.server.OnlineMap.Store(this.Name, this)
	//并广播
	this.server.BroadCast(this, "online")
}

// 用户下线
func (this *User) Offline() {
	//将当前用户从OnlineMap中删除
	this.server.OnlineMap.Delete(this.Name)
	//并广播
	this.server.BroadCast(this, "offline")
}

// 给当前User对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息
func (this *User) DoMessage(msg string) {
	if msg == "who" { //查询当前在线用户
		this.server.OnlineMap.Range(func(key, value interface{}) bool {
			cli := value.(*User) //空接口类型断言
			sendMsg := "[" + cli.Addr + "]" + cli.Name + ":" + "is online...\n"
			this.SendMsg(sendMsg)
			return true
		})
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//修改用户名格式: rename|xxx
		newName := strings.Split(msg, "|")[1] //按“|”截取后的第2元素
		//判断name是否已经存在
		_, ok := this.server.OnlineMap.Load(newName)
		if ok {
			this.SendMsg(fmt.Sprintf("新用户名:%s 已被使用\n", newName))
		} else {
			this.server.OnlineMap.Delete(this.Name)
			this.server.OnlineMap.Store(newName, this)

			this.Name = newName
			this.SendMsg(fmt.Sprintf("更新成功: %s\n", newName))
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//私聊用户格式: to|张三|...内容

		//1.获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("wrong format, please use \"to|name|content\"\n")
			return
		}
		//2.根据用户名得到对方User对象
		remoteUser, ok := this.server.OnlineMap.Load(remoteName)
		if ok == false {
			this.SendMsg("user not exist\n")
			return
		}
		//3.获取消息内容，通过对方User对象发送
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("error: no content\n")
			return
		}
		ru := remoteUser.(*User) //空接口类型断言
		ru.SendMsg(fmt.Sprintf("[%s] says to you : %s\n", this.Name, content))

	} else {
		this.server.BroadCast(this, msg) //广播消息
	}

}
