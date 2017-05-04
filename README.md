```
Install:
	go install -ldflags "-s -w" github.com/czxichen/wstools
	
Wstools Usage:
	wstools [arguments]
	
The commands are:
	fileserver  用过http协议传输共享文件
	mail        通过smtp协议发送邮件
	compress    压缩解压文件
	net         检测远程地址或端口是否通
	find        根据条件查找文件
	md5         计算指定路径的md5值,可以是目录
	compare     对文件或者目录经进行比较
	fsnotify    可以用来监控文件或者目录的变化
	ssh         使用ssh协议群发命令或发送文件
	ftp         使用ftp协议下载或上传文件
	replace     替换文本内容
	sysinfo     查看系统信息
	tail		从文件结尾或指定位置读取内容
	deploy		快速搭建服务器

Use "wstools help [command]" for more information about a command.

Example:
	wstools fileserver -d command  -l 192.168.0.2:8080 -i -a
	wstools mail -u root -p 123456 -F czxichen@163.com -T czxichen@163.com
	wstools compress -x -p tmp.zip -o ./
	wstools net -a telnet -H 127.0.0.1:80,www.baidu.com:80
	wstools find -d "./" -b 20160101 -l 10 -s ".go"
	wstools md5 -d "./" -e ".exe"
	wstools compare -s command -d command_new -c diff
	wstools fsnotify -d tools -s scripts.bat
	wstools ssh -u root -p 123456 -H 192.168.1.2:22 -s main.go -d /tmp
	wstools ftp -l main.go -r /mnt/main.go -g false
	wstools replace -o "Hello world" -n "World Hello" -d ./ -s ".json" -q=true
	wstools sysinfo
	wstools tail -f main.go -i 100 -s 200 -o tmp.txt
	wstools deploy server|client -h
	
```
