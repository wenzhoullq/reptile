## 配合八爪鱼爬取的URL进行图片爬取
根据八爪鱼爬取的EXCEL进行下载
excel里必须带有图片或则图片字段的文本
```shell
windos环境
go build 
.\reptile.exe C:/Users/54910/Desktop/sougou.xlsx 1
linux环境
先运行run.bat脚本,然后对文件授权
chmod +x reptile
.\reptile sougou.xlsx 0
```
第一个参数是xlsx的路径
参数是文件路径,文件路径不要带空格
第二个参数代表运行环境,如果是1代表是window上执行的,如果是0代表在linux上执行的
windows环境下 xlsx文件路径需要 / 而不是 \ 进行隔离
### 功能说明

最大并发数在MAXGET进行设置,一般不超过1000,http请求会开启新的协程,最后实际运行的协程是5倍

出现大量timeout后需要修改MAXGET
