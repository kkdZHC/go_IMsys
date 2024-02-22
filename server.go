package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap sync.Map
	//消息广播的channel
	Massage chan string
}

// 构造函数, 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: sync.Map{},
		Massage:   make(chan string),
	}
	return server
}

// 广播消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Massage <- sendMsg
	fmt.Printf("sendMsg: %v\n", sendMsg)
}

// 监听Message管道的goroutine，有消息就发送给User
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Massage
		// fmt.Printf("msg: %v\n", msg)
		this.OnlineMap.Range(func(key, value interface{}) bool {
			cli := value.(*User) //空接口类型断言
			cli.C <- msg
			return true
		})

	}
}

func (this *Server) Handler(conn net.Conn) {
	//当前连接的业务
	fmt.Println("连接建立成功...")
	//用户上线了，新建user
	user := NewUser(conn, this)
	//用户上线
	user.Online()

	isLive := make(chan bool)

	//接受客户端发送消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				//用户下线
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Printf("read from conn, err: %v\n", err)
				return
			}
			//一旦conn读到消息，就isLive记录
			isLive <- true
			//格式化消息
			msg := string(buf[:n-1]) //去除\n

			//用户对msg进行处理
			user.DoMessage(msg)
		}
	}()

	//超时强踢
	for {
		select {
		case <-isLive:
			//不需要执行操作, select进入这个case会执行下面所有case的条件句
			//所以自动重置计时器
		case <-time.After(time.Hour):
			//超时强制关闭当前User
			user.SendMsg("timeout...")

			//关闭管道和连接
			close(user.C)

			conn.Close()
			return //或runtime.Goexit()
		}
	}
}

// 启动服务器接口
func (this *Server) Start() {
	//1. socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Printf("listen err: %v\n", err)
		return
	}
	//4. close listen socket
	defer listener.Close()

	//启动监听massage管道
	go this.ListenMessage()

	for {
		//2.accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("listen accept err: %v\n", err)
			continue
		}
		//3. do handler
		go this.Handler(conn)

	}

}
