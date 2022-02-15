package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("Now Listening on 8080")
	gin.SetMode(gin.ReleaseMode)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := gin.Default()
	server.LoadHTMLGlob("public/*.html")
	server.Static("/assets", "./public/assets")
	server.GET("/", index)
	server.GET("/aboutme", about)
	server.GET("/contact", contact)
	server.GET("/post", post)
	server.GET("/admin", admin)
	server.GET("/login", login)
	server.GET("/logout", logout)
	server.POST("/send", send)
	server.POST("/newpost", newpost)
	server.POST("/log", log)
	if err := server.Run(":" + port); err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}

type Res struct {
	content string
	title   string
	id      int
	date    string
}

func index(c *gin.Context) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/blog")
	if err != nil {
		panic(err.Error())
	}
	results, err := db.Query("SELECT id FROM blogs")
	if err != nil {
		panic(err.Error())
	}
	var id []int
	var res Res
	var largerNumber, temp int
	for results.Next() {
		err = results.Scan(&res.id)
		if err != nil {
			panic(err.Error())
		}
		id = append(id, res.id)
	}
	for _, element := range id {
		if element > temp {
			temp = element
			largerNumber = temp
		}
	}
	type Result struct {
		Title   string
		Content string
		Date    string
	}
	err = db.QueryRow("SELECT title, content, date FROM blogs WHERE id = ?", largerNumber).Scan(&res.title, &res.content, &res.date)
	if err != nil {
		panic(err.Error())
	}
	val, _ := c.Cookie("admin")
	if val == "" {
		c.HTML(200, "index.html", gin.H{
			"result": Result{
				Title:   res.title,
				Content: res.content,
				Date:    res.date,
			},
			"login": "nil",
		})
	} else {
		c.HTML(200, "index.html", gin.H{
			"result": Result{
				Title:   res.title,
				Content: res.content,
				Date:    res.date,
			},
			"login": "true",
		})
	}
}

func about(c *gin.Context) {
	val, _ := c.Cookie("admin")
	if val == "" {
		c.HTML(200, "about.html", gin.H{
			"login": "nil",
		})
	} else {
		c.HTML(200, "about.html", gin.H{
			"login": "true",
		})
	}
}

func contact(c *gin.Context) {
	val, _ := c.Cookie("admin")
	if val == "" {
		c.HTML(200, "contact.html", gin.H{
			"login": "nil",
		})
	} else {
		c.HTML(200, "contact.html", gin.H{
			"login": "true",
		})
	}
}

func post(c *gin.Context) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/blog")
	if err != nil {
		panic(err.Error())
	}
	results, err := db.Query("SELECT title, content, date FROM blogs")
	if err != nil {
		panic(err.Error())
	}
	type Post struct {
		Title   string
		Content string
		Date    string
	}

	type PostCollection struct {
		Posts []Post
	}
	ress := PostCollection{}
	posts := Post{}
	for results.Next() {
		err = results.Scan(&posts.Title, &posts.Content, &posts.Date)
		if err != nil {
			panic(err.Error())
		}
		ress.Posts = append(ress.Posts, posts)
	}
	val, _ := c.Cookie("admin")
	if val == "" {
		c.HTML(200, "post.html", gin.H{
			"ress":  ress.Posts,
			"login": "nil",
		})
	} else {
		c.HTML(200, "post.html", gin.H{
			"ress":  ress.Posts,
			"login": "true",
		})
	}
}

func send(c *gin.Context) {
	r := c.Request
	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	type jsonStruct struct {
		Name    string
		Message string
		Email   string
	}
	jsondata, _ := json.Marshal(jsonStruct{
		Name:    name,
		Message: message,
		Email:   email,
	})

	postBody, _ := json.Marshal(map[string]string{
		"username":   "Reviath",
		"content":    string(jsondata),
		"avatar_url": "https://cdn.discordapp.com/avatars/894273903600484384/a_468b7aea9b62309e1afcd7849011b3d6.gif",
	})
	resp, err := http.Post("https://discord.com/api/webhooks/942685779468116018/q4TT9TqKV27sBSiEZIQHQQPECCmjqqi47UjfF7Nb7DH5SmXd7NkZSLcsfY-cVsST3mzh", "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		fmt.Printf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	c.Redirect(http.StatusMovedPermanently, "/contact")
}

func admin(c *gin.Context) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/blog")
	if err != nil {
		panic(err.Error())
	}
	type jsonStruct struct {
		Username string
		Password string
	}
	var jsonstruct jsonStruct
	val, _ := c.Cookie("admin")
	if val == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	json.Unmarshal([]byte(val), &jsonstruct)
	row, _ := db.Query("SELECT id FROM admin WHERE username = ? AND password = ?", jsonstruct.Username, jsonstruct.Password)

	c.HTML(200, "admin.html", gin.H{
		"cookie": jsonStruct{
			Username: jsonstruct.Username,
			Password: jsonstruct.Password,
		},
	})
	defer row.Close()
}

func newpost(c *gin.Context) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/blog")
	if err != nil {
		panic(err.Error())
	}
	r := c.Request
	title := r.FormValue("title")
	content := r.FormValue("content")
	current_time := time.Now()
	currenttime := fmt.Sprintf("%s, %d %s %d", current_time.Weekday().String(), current_time.Day(), current_time.Month().String(), current_time.Year())
	insert, err := db.Query("INSERT INTO blogs (title, content, date) VALUES (?, ?, ?)", title, content, currenttime)

	if err != nil {
		panic(err.Error())
	}
	defer insert.Close()
	c.Redirect(http.StatusMovedPermanently, "/admin")
}

func login(c *gin.Context) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/blog")
	if err != nil {
		panic(err.Error())
	}
	type jsonStruct struct {
		Username string
		Password string
	}
	var jsonstruct jsonStruct
	val, _ := c.Cookie("admin")
	if val != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/admin")
		return
	}
	json.Unmarshal([]byte(val), &jsonstruct)
	row, _ := db.Query("SELECT id FROM admin WHERE username = ? AND password = ?", jsonstruct.Username, jsonstruct.Password)
	c.HTML(200, "login.html", gin.H{})
	defer row.Close()
}

func log(c *gin.Context) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/blog")
	if err != nil {
		panic(err.Error())
	}
	r := c.Request
	username := r.FormValue("username")
	password := r.FormValue("password")
	type jsonStruct struct {
		Username string
		Password string
	}

	row, err := db.Query("SELECT id FROM admin WHERE username = ? AND password = ?", username, password)
	if err == nil {
		jsondata, _ := json.Marshal(jsonStruct{
			Username: username,
			Password: password,
		})
		c.SetCookie("admin", string(jsondata), 60*60*24*7, "/", "localhost", false, false)
		c.Redirect(http.StatusMovedPermanently, "/admin")
	} else {
		c.Redirect(http.StatusMovedPermanently, "/login")
	}
	defer row.Close()
}

func logout(c *gin.Context) {
	val, _ := c.Cookie("admin")
	if val == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	} else {
		c.SetCookie("admin", "nil", -1, "/", "localhost", false, false)
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}
