package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	ID        int     `json:"id"`
	Nazwa     string  `json:"nazwa"`
	Opis      string  `json:"opis"`
	Cena      float64 `json:"cena"`
	Zdjecie   string  `json:"zdj"`
	Kategoria string  `json:"kategoria"`
}

type Database struct {
	*sql.DB
}

func initDB() (*Database, error) {
	db, err := sql.Open("sqlite3", "./baza.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS produkty (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nazwa TEXT,
		opis TEXT,
		cena REAL,
		zdj TEXT,
		kategoria TEXT
	);`)
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}

func getProdukty(db *Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, nazwa, opis, cena, zdj, kategoria FROM produkty")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var p Product
			err := rows.Scan(&p.ID, &p.Nazwa, &p.Opis, &p.Cena, &p.Zdjecie, &p.Kategoria)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
				return
			}
			products = append(products, p)
		}

		c.JSON(http.StatusOK, products)
	}
}

func MainHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()
	r := gin.Default()
	r.GET("/ping", MainHandler)
	r.GET("/produkty", getProdukty(db))
	r.Run()
}
