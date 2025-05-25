package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

func ensureTokenCookie() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := c.Request.Cookie("token")
		if err != nil {
			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "token",
				Value:    "",
				Path:     "/",
				MaxAge:   3600,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		}
		c.Next()
	}
}

var tokenToCategory = map[string]string{
	"abc123": "special",
}

func specialProdukty(db *Database) gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenCookie, err := c.Request.Cookie("token")
		if err != nil || tokenCookie.Value == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			return
		}

		category, ok := tokenToCategory[tokenCookie.Value]
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
			return
		}

		rows, err := db.Query(`
			SELECT id, nazwa, opis, cena, zdj, kategoria
			FROM produkty
			WHERE kategoria = ?
		`, category)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Nazwa, &p.Opis, &p.Cena, &p.Zdjecie, &p.Kategoria); err != nil {
				continue
			}
			products = append(products, p)
		}

		c.JSON(http.StatusOK, gin.H{"special": products})
	}
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

func addToCartHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID       int `json:"id"`
			Quantity int `json:"quantity"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		cart := make(map[int]int)
		if cookie, err := c.Request.Cookie("koszyk"); err == nil && cookie.Value != "" {
			data, _ := url.QueryUnescape(cookie.Value)
			pairs := strings.Split(data, ",")
			for _, pair := range pairs {
				parts := strings.Split(pair, ":")
				if len(parts) != 2 {
					continue
				}
				id, _ := strconv.Atoi(parts[0])
				qty, _ := strconv.Atoi(parts[1])
				if id > 0 {
					cart[id] = qty
				}
			}
		}
		cart[req.ID] += req.Quantity
		var cookieParts []string
		for id, qty := range cart {
			cookieParts = append(cookieParts, fmt.Sprintf("%d:%d", id, qty))
		}
		cookieValue := url.QueryEscape(strings.Join(cookieParts, ","))

		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "koszyk",
			Value:    cookieValue,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		c.JSON(http.StatusOK, gin.H{"message": "dodano"})
	}
}

func viewCartHandler(db *Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		cart := make(map[int]int)
		if cookie, err := c.Request.Cookie("koszyk"); err == nil && cookie.Value != "" {
			data, _ := url.QueryUnescape(cookie.Value)
			pairs := strings.Split(data, ",")
			for _, pair := range pairs {
				parts := strings.Split(pair, ":")
				if len(parts) != 2 {
					continue
				}
				id, _ := strconv.Atoi(parts[0])
				qty, _ := strconv.Atoi(parts[1])
				if id > 0 {
					cart[id] = qty
				}
			}
		}

		if len(cart) == 0 {
			c.JSON(http.StatusOK, gin.H{"koszyk": []any{}})
			return
		}

		var ids []string
		var args []any
		for id := range cart {
			ids = append(ids, "?")
			args = append(args, id)
		}
		query := fmt.Sprintf("SELECT id, nazwa, opis, cena, zdj, kategoria FROM produkty WHERE id IN (%s)", strings.Join(ids, ","))

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
			return
		}
		defer rows.Close()

		var response []map[string]any
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Nazwa, &p.Opis, &p.Cena, &p.Zdjecie, &p.Kategoria); err != nil {
				continue
			}
			response = append(response, gin.H{
				"product":  p,
				"quantity": cart[p.ID],
			})
		}

		c.JSON(http.StatusOK, gin.H{"koszyk": response})
	}
}

func MainHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "5DjHs",
	})
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed connect: %v", err)
	}
	defer db.Close()
	r := gin.Default()

	r.Use(ensureTokenCookie())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))
	r.GET("/clue", MainHandler)
	r.GET("/produkty/:id", getProdukty(db))
	r.GET("/kategoria/:kategoria", getProduktyByKategoria(db))
	r.POST("/koszyk", addToCartHandler())
	r.GET("/koszyk", viewCartHandler(db))
	r.GET("/special", specialProdukty(db))

	r.Run()
}
