package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Product_y struct {
	Nazwa     string
	Opis      string
	Cena      float64
	Zdjecie   string
	Kategoria string
}

func OpenDB() (*sql.DB, error) {
	return sql.Open("sqlite3", "baza.db")
}

func InsertProduct(db *sql.DB, product Product_y) error {
	query := `
	INSERT INTO produkty (nazwa, opis, cena, zdj, kategoria)
	VALUES (?, ?, ?, ?, ?);
	`
	_, err := db.Exec(query, product.Nazwa, product.Opis, product.Cena, product.Zdjecie, product.Kategoria)
	return err
}

func main() {
	db, err := OpenDB()
	if err != nil {
		log.Fatal("Could not open DB:", err)
	}
	defer db.Close()

	product := Product_y{
		Nazwa:     "Ametyst",
		Opis:      "zielony kamyk",
		Cena:      30,
		Zdjecie:   "/zdj/ametyst.jpg",
		Kategoria: "kamienie",
	}

	if err := InsertProduct(db, product); err != nil {
		log.Fatal("Insert failed:", err)
	}

	fmt.Println("✅ Product inserted!")
}
