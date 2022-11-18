package kvs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func ToDuration(v string) (time.Duration, error) {
	if v == "" {
		return 0, errors.New("time: 字符串不能为空 ")
	}
	v = strings.ToLower(v)
	i, err := strconv.ParseInt(v, 10, 64)
	if err == nil {
		return time.Duration(i) * time.Second, nil
	}
	return time.ParseDuration(v)

	//if strings.LastIndex(v, TIME_MS) > 0 {
	//    i, err := strconv.ParseInt(strings.TrimSuffix(v, TIME_MS), 10, 0)
	//    return time.Duration(i) * time.Millisecond, err
	//} else {
	//    i, err := strconv.ParseInt(strings.TrimSuffix(v, TIME_S), 10, 0)
	//    return time.Duration(i) * time.Second, err
	//}
}

func ExecCommand(commandName string, params ...string) bool {

	cmd := exec.Command(commandName, params...)

	//显示运行的命令
	fmt.Println(commandName, cmd.Args)

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println(err)
		return false
	}

	cmd.Start()

	reader := bufio.NewReader(stdout)

	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			fmt.Println("exit cmd start.")
			break
		}
		fmt.Print(line)
	}
	err = cmd.Process.Signal(os.Kill)
	fmt.Println(err)
	cmd.Wait()
	return true
}

func GetCurrentFilePath(fileName string, skip int) string {
	dir1, _ := os.Getwd()
	dir2 := func() string {
		abs, err := filepath.Abs(os.Args[0])
		if err != nil {
			log.Fatal(err)
		}
		dir := filepath.Dir(abs)
		return dir
	}()

	dir3 := func() string {
		//获取当前函数Caller reports，取得当前调用对应的文件
		_, f, _, _ := runtime.Caller(skip)
		//解析出所在目录
		return filepath.Dir(f)
	}()

	//默认当前文件夹，如果运行时的二进制和当前工作文件夹一样，说明是通过二进制运行的，返回二进制当前目录
	dir := dir1
	//如果运行时的二进制和当前工作文件夹不一样，说明是通过go run运行的，使用runtime.Caller 路径
	if dir1 != dir2 {
		dir = dir3
	}
	//组装配置文件路径
	file := filepath.Join(dir, fileName)
	return file
}

func CurrentFilePath(fileName string, skip int) string {
	dir, _ := os.Getwd()
	file := filepath.Join(dir, fileName)
	return file
}

func CurrentFilePathRuntime(fileName string, skip int) string {
	//获取当前函数Caller reports，取得当前调用对应的文件
	_, f, _, _ := runtime.Caller(skip)
	//解析出所在目录
	dir := filepath.Dir(f)
	//组装配置文件路径
	file := filepath.Join(dir, fileName)
	return file
}
func ReadFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, err := ioutil.ReadAll(f)
	return src, err
}

func ParseBool(v string) (bool, error) {
	if strings.ToUpper(v) == "YES" || strings.ToUpper(v) == "Y" {
		return true, nil
	}
	if strings.ToUpper(v) == "NO" || strings.ToUpper(v) == "N" {
		return false, nil
	}
	b, err := strconv.ParseBool(v)
	return b, err
}

func ByProperties(content string) *MapProperties {
	y, err := ReadProperties(strings.NewReader(content))
	if err != nil {
		log.Error(err)
		return nil
	}
	return &y.MapProperties
}

func Join(elem ...string) string {
	var buf bytes.Buffer
	for _, e := range elem {
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			buf.WriteString(".")
		}
		if strings.HasSuffix(e, ".") {
			buf.WriteString(e[:len(e)-1])
		} else {
			buf.WriteString(e)
		}
	}
	return ""
}
