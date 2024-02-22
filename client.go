package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn

	flag int //当前模式
}

func NewClient(Ip string, port int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   Ip,
		ServerPort: port,
		flag:       999,
	}
	//连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", Ip, port))
	if err != nil {
		fmt.Printf("Dial server err: %v\n", err)
		return nil
	}
	client.conn = conn

	//返回对象
	return client
}

// 处理server回应消息
func (this *Client) DialResponse() {
	for {
		buf := make([]byte, 4096)
		n, err := this.conn.Read(buf)
		if err != nil {
			fmt.Printf("read from server err: %v\n", err)
		} else {
			fmt.Println(string(buf[:n]))
		}
	}
	//简写：
	//io.Copy(os.Stdout, this.conn)
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>请输入合法数字<<<")
		return false
	}

}

// 更新用户名
func (this *Client) UpdateName() bool {
	fmt.Println(">>>请输入用户名<<<")
	fmt.Scanln(&this.Name)
	sendMsg := fmt.Sprintf("rename|%s\n", this.Name)
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Printf("write rename err: %v\n", err)
		return false
	}
	return true
}

// 公聊
func (this *Client) PublicChat() {
	//输入
	var msg string
	fmt.Println(">>>请输入聊天内容, exit退出<<<")
	fmt.Scanln(&msg)
	//发给服务器
	for msg != "exit" {
		if len(msg) != 0 {
			_, err := this.conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Printf("send msg err: %v\n", err)
				break
			}
		}
		msg = ""
		fmt.Println(">>>请输入, exit退出<<<")
		fmt.Scanln(&msg)
	}
}

// 查询在线用户
func (this *Client) SelectUsers() {
	msg := "who\n"
	_, err := this.conn.Write([]byte(msg))
	if err != nil {
		fmt.Printf("who err: %v\n", err)
		return
	}
}

// 私聊
func (this *Client) PrivateChat() {

	this.SelectUsers()

	var target string
	var msg string
	head := "to"

	fmt.Println("请输入你要私聊的用户,exit退出\n")
	fmt.Scanln(&target)
	for target != "exit" {
		head = fmt.Sprintf("%s|%s", head, target)
		fmt.Println("请开始聊天,exit退出: \n")
		fmt.Scanln(&msg)
		for msg != "exit" {
			if len(msg) != 0 {
				msg = fmt.Sprintf("%s|%s\n\n", head, msg)
				_, err := this.conn.Write([]byte(msg))
				if err != nil {
					fmt.Printf("send msg to %s err: %v\n", target, err)
					return
				}
				msg = ""
			}
			fmt.Scanln(&msg)
		}
		this.SelectUsers()
		fmt.Println("请输入你要私聊的用户,exit退出\n")
		fmt.Scanln(&target)
	}

}

func (this *Client) Run() {
	for this.flag != 0 {
		for !this.menu() {
		}
		//根据不同flag处理不同业务
		switch this.flag {
		case 1:
			fmt.Println("公聊模式")
			this.PublicChat()
		case 2:
			fmt.Println("私聊模式")
			this.PrivateChat()
		case 3:
			fmt.Println("更新用户")
			this.UpdateName()
		}
	}
}

// 命令行解析，传入到全局变量中
var serverIp string
var serverPort int

// 解析命令行需要再main函数之前，init函数中
//
//	./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP") //default: 127.0.0.1
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口")       //default:8888
}
func main() {
	//命令行解析
	flag.Parse() //用到flag包

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>链接服务器失败<<<<<")
		return
	}
	fmt.Println(">>>>>链接成功!<<<<<")
	//开启一个goroutine处理回复
	go client.DialResponse()

	//启动客户端业务
	client.Run()
}
