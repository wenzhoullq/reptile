## 配合八爪鱼爬取的URL进行图片爬取

```shell
go build 
.\reptile.exe C:/Users/54910/Desktop/sougou.xlsx
```
参数是文件路径,文件路径不要带空格
xlsx文件路径需要 / 而不是 \ 进行隔离
### 功能说明

最大并发数在MAXGET进行设置,一般不超过1000,http请求会开启新的协程,最后实际运行的协程是5倍

出现大量timeout后需要修改MAXGET
