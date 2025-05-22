package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
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

func getProduktyByKategoria(db *Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		kategoria := c.Param("kategoria")

		rows, err := db.Query("SELECT id, nazwa, opis, cena, zdj, kategoria FROM produkty WHERE kategoria = ?", kategoria)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Nazwa, &p.Opis, &p.Cena, &p.Zdjecie, &p.Kategoria); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "nie udalo sie sfetchowac"})
				return
			}
			products = append(products, p)
		}

		if len(products) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No products found for this category"})
			return
		}

		c.JSON(http.StatusOK, products)
	}
}

func addToCartHandler(db *Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ProductID int `json:"product_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "zly request"})
			return
		}
		productID := strconv.Itoa(req.ProductID)

		koszyk, err := c.Cookie("koszyk")
		var itemy []string
		if err == nil && koszyk != "" {
			itemy = strings.Split(koszyk, ",")
		}

		itemy = append(itemy, productID)

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "koszyk",
			Value:    strings.Join(itemy, ","),
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		placeholders := strings.Repeat("?,", len(itemy))
		placeholders = strings.TrimRight(placeholders, ",")

		query := fmt.Sprintf("SELECT id, nazwa, opis, cena, zdj, kategoria FROM produkty WHERE id IN (%s)", placeholders)
		args := make([]interface{}, len(itemy))
		for i, v := range itemy {
			args[i] = v
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database failed"})
			return
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Nazwa, &p.Opis, &p.Cena, &p.Zdjecie, &p.Kategoria); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan"})
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
		log.Fatalf("Failed connect: %v", err)
	}
	defer db.Close()
	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/ping", MainHandler)
	r.GET("/produkty/:id", getProdukty(db))
	r.GET("/kategoria/:kategoria", getProduktyByKategoria(db))
	r.POST("/koszyk/dodaj", addToCartHandler(db))
	r.Run()
}
