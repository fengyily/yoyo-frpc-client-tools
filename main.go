package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var path = "/Users/fengyi/frpc"

func init() {
	flag.StringVar(&path, "c", "", "配置文件(frpc.ini)的存放路径，如：/Users/fengyi/frpc")
}

func main() {
	flag.Parse()

	if path == "" {
		panic("请使用-c指定配置文件，配置文件(frpc.ini)的存放路径，如：/Users/fengyi/frpc.")
	} else {
	reset_path:
		if dir_exists(path) {
			println("文件所在路径：" + path)
		} else {
			println("文件所在路径：" + path + "，不存在，请重新输一个：")
			fmt.Scanln(&path)
			goto reset_path
		}
	}
	cmd_list := map[string]string{
		"s":    "show all frpc info.",
		"add":  "add frpc client uuid.",
		"exit": "exit",
	}
	for key, intro := range cmd_list {
		println(key, ":\t", intro)
	}

retry:
	var cmd string
	var uuid string
	var shop_name string
	fmt.Println("please enter a command.")

	fmt.Scanln(&cmd, &uuid, &shop_name)

	if cmd == "s" {
		show_all()
		goto retry
	}

	if cmd == "add" {
		if len(uuid) > 10 {

			port := 8817

			write_frpc_content_conf(uuid, port, shop_name)

			write_frpc_conf()
			restart_frpc()

			show_all()

			println("ssh -p " + strconv.Itoa(port) + " yoyo@127.0.0.1")
		} else {
			println("please enter uuid.")
		}
		goto retry
	}

	if cmd != "exit" {
		for key, intro := range cmd_list {
			println(key, ":\t", intro)
		}
		goto retry
	}

	println("quit frpc manager tools now.")
}

// 执行Shell脚本
func restart_frpc() (error, string, string) {
	defer func() {
		if err := recover(); err != nil {
			println("Shell", "发生例外了", err)
			err = errors.New("发生例外了")
		}
	}()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", "killall frpc")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	cmd.Wait()
	out := stdout.String()
	//logger.Debug("Shell", command)

	return err, out, stderr.String()
}

func show_all() {
	defer func() {
		if err := recover(); err != nil {
			println("Shell", "发生例外了", err)
			err = errors.New("发生例外了")
		}
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", "grep \"#ssh\" "+path+"/frpc_content.ini")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run()

	cmd.Wait()
	out := stdout.String()
	println(out)
}

func write_frpc_conf() (err error) {

	var file *os.File
	file, err = os.OpenFile(path+"/frpc.ini", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		println("修改文件异常：/frpc.ini", err.Error())
		return
	}
	defer file.Close()

	file.WriteString("[common]\n")
	file.WriteString("server_addr = frp.yoyo.link\n")
	file.WriteString("server_port = 7000\n")
	file.WriteString("token = yoyo_ai_token\n")

	content, _ := read_small_file(path + "/frpc_content.ini")

	file.WriteString("\n")
	file.WriteString(content)
	file.WriteString("\n")
	return
}

func write_frpc_content_conf(uuid string, port int, shop_name string) {
	var file *os.File
	content, _ := read_small_file(path + "/frpc_content.ini")

	clients := strings.Split(content, "###")

	for _, client := range clients {
		if strings.Index(client, uuid) > 0 {
			println("already exists:\r\n", client)
			return
		}
	}

	index := 0
	if len(clients) > 10 {
		index = 1
	}

	file, err := os.OpenFile(path+"/frpc_content.ini", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		println("修改文件异常：/frpc_content.ini", err.Error())
		return
	}
	defer file.Close()
	for i, client := range clients {
		if i >= index && len(client) > 0 {
			file.WriteString("\n###\n")
			file.WriteString(client)
		}
	}
	file.WriteString("\n###\n")
	file.WriteString("#\t" + shop_name + "\n")
	file.WriteString("[" + uuid + "]\n")
	file.WriteString("sk=" + uuid + "\n")
	file.WriteString("server_name=" + uuid + "\n")
	file.WriteString("type=stcp\n")
	file.WriteString("role=visitor\n")
	file.WriteString("bind_addr=127.0.0.1\n")
	file.WriteString("bind_port=" + strconv.Itoa(port+len(clients)) + "\n")
	file.WriteString("#ssh -p " + strconv.Itoa(port+len(clients)) + " yoyo@127.0.0.1\t\t" + shop_name + " \n")
}

// 用于读小文件，一次性读取
func read_small_file(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("read_small_file", filePath, err)
	} else {
		//logger.Debug("read_small_file", string(content))
	}

	lenght := len(content)
	//去除末尾\0
	if lenght > 1 {
		c := content[:lenght-1]
		return string(c), nil
	}
	return string(content), nil
}

// DirExists 检查目录是否存在
func dir_exists(fileAddr string) bool {
	s, err := os.Stat(fileAddr)
	if err != nil {
		println("DirExists", err)
		return false
	}
	return s.IsDir()
}
