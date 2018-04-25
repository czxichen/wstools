[![Build Status](https://travis-ci.org/czxichen/wstools.svg?branch=master)](https://travis-ci.org/czxichen/wstools)
[![GoDoc](https://godoc.org/github.com/czxichen/wstools?status.svg)](http://godoc.org/github.com/czxichen/wstools)
[![Go Report](https://goreportcard.com/badge/github.com/czxichen/wstools)](https://goreportcard.com/report/github.com/czxichen/wstools)

# Install:
	* make
	* go install -ldflags "-s -w" github.com/czxichen/wstools

# Wstools Usage:
	* wstools [command]
	* wstools help [command] for more information about a command

# The commands are:
	* http        用过http协议传输共享文件
	* mail        通过smtp协议发送邮件
	* compress    压缩解压文件
	* net         检测远程地址或端口是否通
	* find        根据条件查找文件
	* md5sum      计算指定路径的md5值,可以是目录
	* ssl         使用rsa对证书简单操作
	* compare     对文件或者目录经进行比较
	* fsnotify    可以用来监控文件或者目录的变化
	* ssh         使用ssh协议群发命令或发送文件
	* ftp         使用ftp协议下载或上传文件
	* replace     替换文本内容
	* sysinfo     查看系统信息
	* tail        从文件结尾或指定位置读取内容
	* deploy      快速搭建服务器
	* watchdog    进程守护

# Example:
	* wstools http -d /tmp/sharedir
	* wstools mail -u user -p passwd -H smtp.163.com:25 -f czxichen@163.com -t czxichen@163.com -c "Hello world"
	* wstools compress -c -s uuid -d uuid.zip
	* wstools net -a ping -i www.baidu.com,www.163.com -c 4 -q
	* wstools find -p ./ -s "1M" -m "-1d"
	* wstools md5sum -s sourcepath -o .*\.go
	* wstools rsa -n -c example.json
	* wstools compare -s uuid -d uuid_new -c diff
	* wstools fsnotify -d tools -s scripts.bat
	* wstools ssh -u root -p 123456 -H 192.168.1.2:22 -s main.go -d /tmp
	* wstools ftp -u root -p toor -s main.go -d /Server/main.go -H 127.0.0.1:21
	* wstools replace -p sourcepath -o "oldstr" -n "newstr"  -s .go
	* wstools sysinfo
	* wstools tail -f main.go -l 10 -n 5 -o tmp.txt
	* wstools deploy server|client -h
	* wstools watchdog -config watch.ini
