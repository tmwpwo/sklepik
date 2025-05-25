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

func DeleteProductByID(db *sql.DB, id int) error {
	query := `DELETE FROM produkty WHERE id = ?;`
	_, err := db.Exec(query, id)
	return err
}

func main() {
	db, err := OpenDB()
	if err != nil {
		log.Fatal("Could not open DB:", err)
	}
	defer db.Close()

	for i, product := range product {
		if err := InsertProduct(db, product); err != nil {
			log.Fatalf("nie dodalo %d: %v", i+1, err)
		}
		fmt.Printf("dodalo %d.\n", i+1)
	}

	// if err := DeleteProductByID(db, 1); err != nil {
	// 	log.Fatalf("nie dodalo: %v", err)
	// }
	// fmt.Println("wyjebane.")
}
