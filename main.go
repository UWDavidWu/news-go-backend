package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"fmt"
)

var (
	db       *sql.DB
	category = [...]string{"business", "entertainment", "general", "health", "science", "sports", "technology"}
	country  = [...]string{"ca", "us", "au"}
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
	Intensity   int    `json:"intensity"`
}

func main() {

	//loadConfig()

	// connect to db
	connectDB(os.Getenv("DATABASE_URL"))

	port := os.Getenv("PORT")
	// port := "8080"

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	r := gin.New()
	r.GET("/news/:country/:category", getHomepageNews)
	go getNewsEvery30Minutes()
	r.Run(":" + port)
}

// func loadConfig() {
// 	config, err := util.LoadConfig(".")
// 	if err != nil {
// 		panic(err)
// 	}

// }

func connectDB(conn string) {

	var err error
	db, err = sql.Open("postgres", conn)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}

// go routine to get news every 30 minutes
func getNewsEvery30Minutes() {
	for {
		for _, cate := range category {
			for _, coun := range country {
				getNews(coun, cate)
				//log
				fmt.Println("Get news from " + coun + " " + cate + time.Now().String())
			}
		}
		time.Sleep(30 * time.Minute)
	}
}

func getHomepageNews(c *gin.Context) {
	country := c.Param("country")
	category := c.Param("category")

	// query for dv where country is country and category is category and publishedAt is the latest and limit 12
	rows, err := db.Query("SELECT * FROM news WHERE country = $1 AND category = $2 ORDER BY publishedAt DESC LIMIT 60", country, category)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	//return result
	var news []News
	for rows.Next() {
		var n News
		err := rows.Scan(&n.Id, &n.Title, &n.Country, &n.Category, &n.Url, &n.UrlToImage, &n.PublishedAt, &n.Intensity)
		if err != nil {
			panic(err)
		}
		news = append(news, n)
	}
	c.JSON(http.StatusOK, gin.H{
		"news": news,
	})

}

func getNews(country string, category string) {
	url := "https://newsapi.org/v2/top-headlines?country=" + country + "&category=" + category + "&apiKey=" + os.Getenv("API_KEY")
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
