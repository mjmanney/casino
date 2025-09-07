package main

import (
    "database/sql"
    "flag"
    "fmt"
    "os"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    dbURL := flag.String("db", os.Getenv("DATABASE_URL"), "Postgres connection string")
    flag.Parse()
    if *dbURL == "" {
        fmt.Println("DATABASE_URL not set and -db not provided")
        os.Exit(1)
    }
    db, err := sql.Open("pgx", *dbURL)
    if err != nil {
        fmt.Println("open error:", err)
        os.Exit(2)
    }
    defer db.Close()
    if err := db.Ping(); err != nil {
        fmt.Println("ping error:", err)
        os.Exit(3)
    }
    var now string
    if err := db.QueryRow("select now()::text").Scan(&now); err != nil {
        fmt.Println("query error:", err)
        os.Exit(4)
    }
    fmt.Println("ok:", now)
}

