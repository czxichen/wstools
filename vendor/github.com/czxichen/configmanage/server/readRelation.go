package server

import (
	"errors"
	"strings"

	"github.com/tealeg/xlsx"
)

//文件类型必须是xlsx格式的.
func relationConfig(path string) (map[string][]string, map[string][]string, error) {
	file, err := xlsx.OpenFile(path)
	if err != nil {
		return nil, nil, err
	}
	if len(file.Sheets) < 2 {
		return nil, nil, errors.New("xlsx format error.")
	}
	configpath, ok := file.Sheet["configpath"]
	if !ok {
		return nil, nil, errors.New("can't find configpath sheet.")
	}
	configPath := make(map[string][]string)
	for _, row := range configpath.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		var list []string
		for _, v := range row.Cells[1:] {
			if strings.TrimSpace(v.Value) == "" {
				continue
			}
			list = append(list, v.Value)
		}
		if len(list) == 0 {
			continue
		}
		configPath[row.Cells[0].Value] = list
	}
	variable, ok := file.Sheet["variable"]
	if !ok {
		return nil, nil, errors.New("can't find variable sheet.")
	}
	if len(variable.Rows) < 1 {
		return nil, nil, errors.New("variable format error.")
	}
	relationVariable := make(map[string][]string)
	var list []string
	for _, v := range variable.Rows[0].Cells {
		list = append(list, v.Value)
	}

	//变量关系表中不能出现_relationVariable_,不然会替换掉key的值.
	relationVariable["_relationVariable_"] = list
	for _, row := range variable.Rows[1:] {
		if len(row.Cells) != len(variable.Rows[0].Cells) {
			continue
		}
		var list []string
		for _, cell := range row.Cells {
			list = append(list, cell.Value)
		}
		if key := row.Cells[0].Value; key != "" {
			relationVariable[key] = list
		}
	}
	return configPath, relationVariable, nil
}
