package util

import (
	"cmnfunc"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func InitCmnConfig(root string, cfgName string) error {
	bytes, err := ioutil.ReadFile(root +"config\\"+ cfgName)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &cmnfunc.Cfg); err != nil {
		return err
	}
	cmnfunc.Cfg["root"] = root
	return nil
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func GetCurrentPath() string {
	s, err := exec.LookPath(os.Args[0])
	checkErr(err)
	i := strings.LastIndex(s, "\\")
	path := string(s[0 : i+1])
	return path
}