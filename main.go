package main

import (
	"bufio"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FileMax = 5000
	MaxGet  = 400
)

var channel chan struct{} = make(chan struct{}, MaxGet)

func main() {
	//cmd := exec.Command("reptile.exe")
	//err := cmd.Run()
	//if err != nil {
	//	log.Println("启动失败:", err)
	//} else {
	//	log.Println("启动成功!")
	//}
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	excelFilePath := os.Args[1]
	excelFileName := "Sheet1"
	savePath, _ := os.Getwd()
	var wg sync.WaitGroup
	var split = "/"
	// 如果第二个启动参数是 1 则代表是在window环境运行 其它参数则是在linux上运行
	if os.Args[2] == "1" {
		split = "\\"
	}
	f, err := excelize.OpenFile(excelFilePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
		}
	}()
	rows, err := f.GetRows(excelFileName)
	if err != nil {
		log.Println(err)
		return
	}
	strs := strings.Split(excelFilePath, "/")
	//读取的文件的文件名,如 "C:/Users/54910/Desktop/搜狗图片搜索 - 泥地.xlsx" ,读取的是 搜狗图片搜索 - 泥地.xlsx
	str := strs[len(strs)-1]
	//去除后缀.xlsx,最后名称是 搜狗图片搜索 - 泥地
	str = str[:len(str)-5]
	savePath = savePath + split + str
	_, err = os.Stat(savePath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(savePath, os.ModePerm)
		if err != nil {
			panic(err)
			return
		}
	}
	log.Println("程序执行:", time.Now())
	//第一行为标题,确定URL所在的列
	index := 0
	for i := 0; i < len(rows[0]); i++ {
		row := rows[0][i]
		if strings.Contains(row, "图片") {
			index = i
			break
		}
	}
	for i := 0; i < MaxGet; i++ {
		channel <- struct{}{}
	}
	var subSavePath string
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		//每5000个创建一个文件夹
		if i%FileMax == 1 {
			subSavePath = savePath + split + time.Now().Format("20060102150405") + "_file" + strconv.Itoa(i/FileMax)
			fmt.Println(subSavePath)
			err = os.Mkdir(subSavePath, os.ModePerm)
			if err != nil {
				log.Println(subSavePath, "文件夹已存在", err)
				return
			}
		}
		if len(row) < 1 || !strings.HasPrefix(row[0], "https://") {
			continue
		}
		//wg 并发执行
		wg.Add(1)
		<-channel
		go downloadPic(&wg, row[index], subSavePath, str, i%FileMax)
	}
	wg.Wait()
	log.Println("程序结束:", time.Now())
}

func downloadPic(wg *sync.WaitGroup, imgUrl, savePath, fileName string, num int) {
	defer func() {
		wg.Done()
		channel <- struct{}{}
	}()
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := client.Get(imgUrl)
	if err != nil {
		log.Println("A error occurred!", err)
		return
	}
	defer res.Body.Close()
	// 获得get请求响应的reader对象
	reader := bufio.NewReaderSize(res.Body, 32*1024)
	fileName = fmt.Sprintf("%s%d.jpg", fileName, num)
	file, err := os.Create(savePath + "/" + fileName)
	if err != nil {
		panic(err)
	}
	// 获得文件的writer对象
	writer := bufio.NewWriter(file)
	_, _ = io.Copy(writer, reader)
	//written, _ := io.Copy(writer, reader)
	//fmt.Printf("Total length: %d", written)
}
