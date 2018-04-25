package parse

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
)

func GetId(user []byte) (uid, gid int) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return -1, -1
	}

	defer file.Close()
	buf := bufio.NewReader(file)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		if bytes.Index(line, user) == 0 {
			list := bytes.Split(line, []byte(":"))
			if len(list) != 7 {
				continue
			}
			uid, err = strconv.Atoi(string(list[2]))
			if err != nil {
				continue
			}
			gid, err = strconv.Atoi(string(list[3]))
			if err != nil {
				continue
			}
			return
		}
	}
	return -1, -1
}
