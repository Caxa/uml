package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Обработчик для регистрации
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Отображение страницы регистрации
		Tmpl1.Execute(w, nil)
	} else if r.Method == http.MethodPost {
		// Обработка формы регистрации
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Ошибка при обработке формы", http.StatusInternalServerError)
			return
		}
		login := r.Form.Get("login")
		password := r.Form.Get("password")

		hashedPassword, err := HashPassword(password)
		if err != nil {
			http.Error(w, "Ошибка при хешировании пароля", http.StatusInternalServerError)
			return
		}

		query := "INSERT INTO users (login, password, role) VALUES ($1, $2, 'user')" // Роль по умолчанию
		_, err = Db.Exec(query, login, hashedPassword)
		if err != nil {
			http.Error(w, "Ошибка при вставке пользователя", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
}

// Функция аутентификации
func AuthenticateUser(ctx context.Context, login, password string) int {
	query := "SELECT id, password, role FROM users WHERE login = $1"
	var currentUser User
	var hashedPassword string
	err := Db.QueryRow(query, login).Scan(&currentUser.IDuser, &hashedPassword, &currentUser.Role)
	if err != nil {
		log.Println("Ошибка при выполнении запроса:", err)
		return 0 // Если пользователя не найдено, возвращаем 0
	}
	if CheckPasswordHash(password, hashedPassword) {
		return currentUser.IDuser // Возвращаем ID пользователя при успешной аутентификации
	} else {
		log.Println("Неверный пароль для пользователя:", login)
		return -1 // Возвращаем -1 при неверном пароле
	}
}

// Хеширование пароля
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Проверка пароля
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Функция для открытия базы данных
func OpenDatabase() {
	var err error
	Db, err = sql.Open("postgres", "user=postgres password=1234 dbname=map sslmode=disable")
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}
	err = Db.Ping()
	if err != nil {
		log.Fatal("Не удалось выполнить ping базы данных:", err)
	}
	log.Println("Успешно подключено к базе данных")
}
