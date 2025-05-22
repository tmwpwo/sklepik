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
		id := c.Param("id")

		query := "SELECT id, nazwa, opis, cena, zdj, kategoria FROM produkty WHERE id = ?"
		row := db.QueryRow(query, id)

		var p Product
		err := row.Scan(&p.ID, &p.Nazwa, &p.Opis, &p.Cena, &p.Zdjecie, &p.Kategoria)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "niema"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "nie udalo sie sfetchowac"})
			}
			return
		}

		c.JSON(http.StatusOK, p)
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
		log.Fatalf("Failed connect: %v", err)
	}
	defer db.Close()
	r := gin.Default()
	r.GET("/ping", MainHandler)
	r.GET("/produkty/:id", getProdukty(db))
	r.GET("/kategorie")
	r.Run()
}
