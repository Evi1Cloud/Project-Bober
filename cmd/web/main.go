package main

import (
	"encoding/base64"
	"fmt"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	_ "github.com/lib/pq" // Импортируем драйвер PostgreSQL
	"github.com/gorilla/sessions"
	"github.com/gorilla/securecookie"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	db       *sql.DB // Добавлено поле db
	session  *sessions.CookieStore // Добавлено поле для хранения сессий
}

type Config struct {
	Addr      string
	StaticDir string
}

func main() {
	cfg := new(Config)
	flag.StringVar(&cfg.Addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.StaticDir, "static-dir", "./ui/static", "Path to static assets")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	connStr := "user=postgres password=1111 dbname=clicker sslmode=disable" 
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		db:       db, 
	}
	
	key := securecookie.GenerateRandomKey(32)
	if key == nil {
		fmt.Println("Ошибка при генерации ключа")
		return
	}

	// Кодируем ключ в строку и устанавливаем в переменную окружения
	os.Setenv("SECRET_KEY", base64.StdEncoding.EncodeToString(key))

	// Декодируем ключ из переменной окружения
	secretKey := os.Getenv("SECRET_KEY")
	decodedKey, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		log.Fatal("Ошибка декодирования секретного ключа:", err)
	}

	app.session = sessions.NewCookieStore(decodedKey)

	srv := &http.Server{
		Addr:     cfg.Addr,
		ErrorLog: errorLog,
		Handler:  app.routes(), // Вызов нового метода app.routes()
	}

	infoLog.Printf("Запуск сервера на %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil { // Исправлено: используем if для обработки ошибки
		errorLog.Fatal(err)
	}
}
