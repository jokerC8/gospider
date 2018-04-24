package main

import (
	"fmt"
	. "gospider/fileutils"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
	"path/filepath"
	"io/ioutil"
	"sync"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	DIR string
	URL string
)

type information struct {
	Title      string  //用户昵称
	URL     string	//用户头像路径
	Gender     string //性别
	Age        int //年龄
}

type user struct{
	Id int `gorm:"type:int(16);not null;primary_key;auto_increment"`
	Name string `gorm:"type:varchar(64);not null;unique"`
	Age int `gorm:"type:int(16);not null"`
}

func init() {
	DIR = filepath.Join(os.Getenv("GOPATH"),"images")  //用户头像存放路径
	URL = "https://www.qiushibaike.com/pic/page/"  //糗事百科糗图板块路径
}

func main() {
	var wg sync.WaitGroup
	fmt.Println("start....")
	if !IsFileExist(DIR) {
		if err := os.Mkdir(DIR, 0755); err != nil {
			fmt.Println(err)
			return
		}
	}
	var infors []information
	for i := 0; i < 20; i++ {
		url_real := URL + strconv.Itoa(i + 1) + "/"
		fmt.Println(url_real)
		resp, err := http.Get(url_real)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var infor information
		doc.Find(".author").Each(func(i int, selection *goquery.Selection) {
			title := selection.Find("h2").Text()
			img_url, _ := selection.Find("img").Attr("src")
			articleGender := selection.Find(".articleGender")
			age, _ := strconv.Atoi(articleGender.Text())
			gender, _ := articleGender.Attr("class")
			if strings.Contains(gender,"manIcon") {
				gender = "man"
			} else if strings.Contains(gender, "womenIcon"){
				gender = "women"
			}
			infor = information{Title:title,URL:"https:"+img_url,Gender:gender,Age:age}
			if !isAlreadyExist(infors,infor.Title) {  //抓取到同一个用户的信息只保存一次，不管该用户发了多少糗图
				infors = append(infors,infor)
			}
		})
	}
	db, err := gorm.Open("mysql","root:chenyanpeng@/testing?parseTime=True&loc=Local")
	if err != nil {log.Fatal(err)}
	if db.HasTable("users") {
		db.DropTable("users")
	}
	db.Set("gorm:table_options","Engine=InnoDb Charset=utf8").CreateTable(&user{})
	if db.NewRecord(&user{}) {
		fmt.Println("create table users succeed")
	} else {
		return
	}
	var u user
	wg.Add(len(infors))
	for index, infor := range infors {
		fmt.Println(infor)
		u = user{Id:index + 1,Name:infor.Title,Age:infor.Age}
		if ! db.NewRecord(&user{}) {
			fmt.Println("insert table users failed")
		}
		db.Create(&u)
		go func() {
			defer wg.Done()
			downLoad(index,infor)
		}()
	}
	wg.Wait()
	fmt.Println("end....")
}

func isAlreadyExist(infors []information, title string) bool {
	for _,infor := range infors {
		if infor.Title == title {
			return true
		}
	}
	return false
}

func downLoad(index int,infor information)  {
	filename := DIR + string(filepath.Separator) + infor.Title + strconv.Itoa(index) + filepath.Ext(infor.URL)  //用户昵称和头像url后缀合成头像图片名称
	fmt.Println(filename)
	resp ,err := http.Get(infor.URL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(filename,data,0666) //读取图片文件写入到filename中
	if err != nil {
		fmt.Println(err)
		return
	}
}