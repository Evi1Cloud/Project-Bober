package main

import (
	"html/template"
	"net/http"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        app.notFound(w)
        return
    }

    session, _ := app.session.Get(r, "session-name")
    username := session.Values["username"]

    files := []string{
        "./ui/html/index.html",
    }

    ts, err := template.ParseFiles(files...)
    if err != nil {
        app.serverError(w, err)
        return
    }

    err = ts.Execute(w, map[string]interface{}{
        "Username": username,
    })
    if err != nil {
        app.serverError(w, err)
    }
}



func (app *application) register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			app.serverError(w, err) // Использование помощника serverError()
			return
		}

		_, err = app.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, hashedPassword)
		if err != nil {
			// Проверка на уникальность имени пользователя
			if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
				http.Error(w, "Username already taken. Please choose another one.", http.StatusConflict)
				return
			}
			app.serverError(w, err) // Использование помощника serverError()
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Отображение формы регистрации
	tmpl, err := template.ParseFiles("./ui/html/register.html")
	if err != nil {
		app.serverError(w, err) // Использование помощника serverError()
		return
	}
	tmpl.Execute(w, nil)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        username := r.FormValue("username")
        password := r.FormValue("password")

        var hashedPassword string
        err := app.db.QueryRow("SELECT password FROM users WHERE username=$1", username).Scan(&hashedPassword)
        if err != nil {
            http.Error(w, "User not found", http.StatusUnauthorized)
            return
        }

        err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
        if err != nil {
            http.Error(w, "Invalid password", http.StatusUnauthorized)
            return
        }

        // Успешный вход, сохраняем имя пользователя в сессии
        session, _ := app.session.Get(r, "session-name")
        session.Values["username"] = username
        session.Save(r, w)

        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    // Отображение формы входа
    tmpl, err := template.ParseFiles("./ui/html/login.html")
    if err != nil {
        app.serverError(w, err)
        return
    }
    tmpl.Execute(w, nil)
}
