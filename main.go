package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var (
	db *sql.DB
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
	Source      Source `json:"source"`
	Title       string `json:"title"`
	Country     string `json:"country"`
	Category    string `json:"category"`
	Url         string `json:"url"`
	UrlToImage  string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "root"
	password = "secret"
	dbname   = "root"
	apiKey   = "7af370dcc040451c8363ff120e0e2478"
)

func main() {
	connectDB()
	startServer()
	go print()
}

func connectDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	// defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}

func startServer() {
	r := gin.Default()
	r.GET("/ping", pong)
	r.GET("/news/home/:country/:category", getHomepageNews)
	r.GET("/news/section/:country/:category", getCategoryNews)
	r.GET("/news/subscribed/:id", getSubscribedNews)
	r.POST("/news", createNews)

	getNews("us", "business")
	r.Run("localhost:8080")
}

func print() {
	fmt.Println("I'm a new routine")
}

func pong(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func getHomepageNews(c *gin.Context) {
	country := c.Param("country")
	category := c.Param("category")
	c.JSON(http.StatusOK, gin.H{
		"country":  country,
		"category": category,
	})
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

func createNews(c *gin.Context) {
	var news News
	c.BindJSON(&news)
	c.JSON(http.StatusOK, gin.H{
		"news": news,
	})
}

func getNews(country string, category string) {
	url := "https://newsapi.org/v2/top-headlines?country=" + country + "&category=" + category + "&apiKey=" + apiKey
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	//decode twice

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
		// fmt.Println(article)
	}

}

// func BulkInsert(unsavedRows []*ExampleRowStruct) error {
// 	valueStrings := make([]string, 0, len(unsavedRows))
// 	valueArgs := make([]interface{}, 0, len(unsavedRows)*3)
// 	i := 0
// 	for _, post := range unsavedRows {
// 		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
// 		valueArgs = append(valueArgs, post.Column1)
// 		valueArgs = append(valueArgs, post.Column2)
// 		valueArgs = append(valueArgs, post.Column3)
// 		i++
// 	}
// 	stmt := fmt.Sprintf("INSERT INTO my_sample_table (column1, column2, column3) VALUES %s", strings.Join(valueStrings, ","))
// 	_, err := db.Exec(stmt, valueArgs...)
// 	return err
// }
