package main

import (
	"bufio"
	"encoding/csv"
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

var channel = make(chan struct{}, MaxGet)

func main() {
	FilePath := os.Args[1]
	savePath, _ := os.Getwd()
	var split = "/"
	// 如果第二个启动参数是 1 则代表是在window环境运行 其它参数则是在linux上运行
	if os.Args[2] == "1" {
		split = "\\"
	}
	switch os.Args[3] {
	case "xlsx":
		downloadByEXCEL(FilePath, split, savePath)
	case "csv":
		downloadByCSV(FilePath, split, savePath)
	default:
		log.Println("启动参数错误,格式为 文件地址 系统环境(1是window环境,其余参数是linux环境) 文件格式(excel,csv)")
	}
}

func downloadPic(wg *sync.WaitGroup, imgUrl, savePath, fileName, split string, num int) {
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
	file, err := os.Create(savePath + split + fileName)
	if err != nil {
		panic(err)
	}
	// 获得文件的writer对象
	writer := bufio.NewWriter(file)
	_, _ = io.Copy(writer, reader)
}

func downloadByCSV(FilePath, split, savePath string) {
	f, err := os.Open(FilePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	rows, err := csvReader.ReadAll()
	//读取的文件的文件名,如 "C:/Users/54910/Desktop/搜狗图片搜索 - 泥地.csv" ,读取的是 搜狗图片搜索 - 泥地.csv
	strs := strings.Split(FilePath, split)
	str := strs[len(strs)-1]
	download(savePath, split, str, rows)

}

func downloadByEXCEL(FilePath, split, savePath string) {
	f, err := excelize.OpenFile(FilePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
		}
	}()
	excelFileName := "Sheet1"
	rows, err := f.GetRows(excelFileName)
	if err != nil {
		log.Println(err)
		return
	}
	strs := strings.Split(FilePath, split)
	//读取的文件的文件名,如 "C:/Users/54910/Desktop/搜狗图片搜索 - 泥地.xlsx" ,读取的是 搜狗图片搜索 - 泥地.xlsx
	str := strs[len(strs)-1]
	//去除后缀.xlsx,最后名称是 搜狗图片搜索 - 泥地
	str = str[:len(str)-5]
	download(savePath, split, str, rows)

}
func download(savePath, split, str string, rows [][]string) {
	wg := sync.WaitGroup{}
	savePath = savePath + split + str
	_, err := os.Stat(savePath)
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
	cnt := 0
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		//每5000个创建一个文件夹
		if i%FileMax == 1 {
			subSavePath = savePath + split + time.Now().Format("20060102150405") + "_file" + strconv.Itoa(i/FileMax)
			err = os.Mkdir(subSavePath, os.ModePerm)
			if err != nil {
				log.Println(subSavePath, "文件夹已存在", err)
				return
			}
		}
		if len(row) < 1 || !strings.HasPrefix(row[index], "https://") {
			continue
		}
		//wg 并发执行
		wg.Add(1)
		<-channel
		cnt++
		go downloadPic(&wg, row[index], subSavePath, str, split, i%FileMax)
	}
	wg.Wait()
	log.Println("程序结束:", time.Now(), " 共下载图片", cnt, "张")
}
