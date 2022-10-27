package main

import (
	"database/sql"
	"encoding/json"

	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/uwdavidwu/News_backend_go/util"

	"fmt"
)

var (
	db       *sql.DB
	category = [...]string{"business", "entertainment", "general", "health", "science", "sports", "technology"}
)

type Object struct {
	Articles []Article `json:"articles"`
}

type Article struct {
	Source      Source `json:"source"`
	Title       string `json:"title"`
	Url         string `json:"url"`
	UrlToImage  string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
}

type Source struct {
	Name string `json:"name"`
}

type News struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Country     string `json:"country"`
	Category    string `json:"category"`
	Url         string `json:"url"`
	UrlToImage  string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
}

const (
	apiKey = "7af370dcc040451c8363ff120e0e2478"
)

func main() {

	config, err := util.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	connectDB(config.DB_SOURCE)
	startServer()
	go print()
}

func connectDB(connStr string) {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}

func startServer() {
	r := gin.Default()
	r.GET("/ping", pong)
	r.GET("/news/home/:country/:category", getHomepageNews)
	r.GET("/news/section/:country/:category", getCategoryNews)
	// getNews("ca", "business")
	r.Run()
}

func pong(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func getHomepageNews(c *gin.Context) {
	country := c.Param("country")
	category := c.Param("category")
	rows, err := db.Query("SELECT * FROM news WHERE country = $1 AND category = $2 ORDER BY publishedAt DESC LIMIT 12", country, category)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var news []News
	for rows.Next() {
		var n News
		err := rows.Scan(&n.Id, &n.Title, &n.Country, &n.Category, &n.Url, &n.UrlToImage, &n.PublishedAt)
		if err != nil {
			panic(err)
		}
		news = append(news, n)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, news)

}

func getCategoryNews(c *gin.Context) {
	country := c.Param("country")
	category := c.Param("category")
	c.JSON(http.StatusOK, gin.H{
		"country":  country,
		"category": category,
	})
}

func getSubscribedNews(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func getNews(country string, category string) {
	url := "https://newsapi.org/v2/top-headlines?country=" + country + "&category=" + category + "&apiKey=" + apiKey
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var obj Object
	if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
		panic(err)
	}

	for _, article := range obj.Articles {
		sqlStatement := `
		INSERT INTO news (title, country, category, url, urlToImage, publishedAt) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (url) DO NOTHING`
		_, err = db.Exec(sqlStatement, article.Title, country, category, article.Url, article.UrlToImage, article.PublishedAt)
		if err != nil {
			panic(err)
		}
	}

}
