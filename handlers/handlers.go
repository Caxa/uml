package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"text/template"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	IDuser int
	Role   string // Поле для роли пользователя
}

var Db *sql.DB
var Tmpl1 = template.Must(template.ParseFiles("templates/register.html"))
var Tmpl2 = template.Must(template.ParseFiles("templates/upload1.html")) // Главная страница
var Files []string

// Функция для открытия базы данных
func OpenDatabase() {
	var err error
	Db, err = sql.Open("postgres", "user=postgres password=1234 dbname=map sslmode=disable")
	if err != nil {
		log.Println("Не удалось подключиться к базе данных:", err)
	} else {
		err = Db.Ping()
		if err != nil {
			log.Println("Не удалось выполнить ping базы данных:", err)
		} else {
			log.Println("Успешно подключено к базе данных")
		}
	}
}

// Получаем имя пользователя по ID из базы данных
func GetUserNameByIDFromDB(userID int) (string, error) {
	var userName string
	query := "SELECT login FROM users WHERE id = $1"
	err := Db.QueryRow(query, userID).Scan(&userName)
	if err != nil {
		log.Println("Ошибка при получении имени пользователя из базы данных:", err)
		return "", err
	}
	return userName, nil
}

// Получаем роль пользователя по ID из базы данных
func GetUserRoleByIDFromDB(userID int) (string, error) {
	var role string
	query := "SELECT role FROM users WHERE id = $1"
	err := Db.QueryRow(query, userID).Scan(&role)
	if err != nil {
		log.Println("Ошибка при получении роли пользователя из базы данных:", err)
		return "", err
	}
	return role, nil
}

// Обработчик главной страницы
func Index(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении роли пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	data := struct {
		UserID   int    // ID пользователя
		UserName string // Имя пользователя
		Role     string // Роль пользователя
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
	}

	err = Tmpl2.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка при выполнении шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Обработчик главной страницы
func Home(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		Tmpl1.Execute(w, nil) // Отображение страницы регистрации
	} else if r.Method == "POST" {
		login := r.FormValue("login")
		password := r.FormValue("password")
		id := AuthenticateUser(r.Context(), login, password)

		// Проверяем результат аутентификации
		if id == -1 {
			log.Println("Ошибка аутентификации: неверный пароль")
			http.ServeFile(w, r, "templates/errorModal.html") // Ошибка при аутентификации
			return
		} else if id == 0 {
			log.Println("Регистрация пользователя, так как id = 0")
			RegisterHandler(w, r) // Если id == 0, направляем на регистрацию
			return
		} else {
			// Успешная аутентификация, перенаправляем на главную страницу
			log.Printf("Успешная аутентификация для пользователя с ID: %d", id)
			http.Redirect(w, r, "/main?id="+strconv.Itoa(id), http.StatusFound)
			return
		}
	}
}

// Обработчик для регистрации
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Отображение страницы регистрации
		Tmpl1.Execute(w, nil)
	} else if r.Method == "POST" {
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

// Обработчик для отображения страницы создания пользователя
func CreateUserPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/create_user.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Ошибка при выполнении шаблона: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// Обработчик для создания пользователя
func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Ошибка при обработке формы", http.StatusInternalServerError)
			return
		}
		login := r.FormValue("login")
		password := r.FormValue("password")
		role := r.FormValue("role")

		hashedPassword, err := HashPassword(password)
		if err != nil {
			http.Error(w, "Ошибка при хешировании пароля", http.StatusInternalServerError)
			return
		}

		query := "INSERT INTO users (login, password, role) VALUES ($1, $2, $3)"
		_, err = Db.Exec(query, login, hashedPassword, role)
		if err != nil {
			http.Error(w, "Ошибка при создании пользователя", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
}

// Обработчик страницы администратора
func AdminPage(w http.ResponseWriter, r *http.Request) {
	// Здесь можно добавить логику для отображения страницы администратора
	tmpl := template.Must(template.ParseFiles("templates/admin_page.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Ошибка при выполнении шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Обработчик для обновления данных пользователя
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Получаем данные из формы
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Ошибка при обработке формы", http.StatusInternalServerError)
			return
		}

		userID := r.FormValue("user_id")
		newLogin := r.FormValue("new_login")
		newPassword := r.FormValue("new_password")
		newRole := r.FormValue("new_role")

		// Проверка и обновление логина
		if newLogin != "" {
			query := "UPDATE users SET login = $1 WHERE id = $2"
			_, err := Db.Exec(query, newLogin, userID)
			if err != nil {
				http.Error(w, "Ошибка при обновлении логина", http.StatusInternalServerError)
				return
			}
		}

		// Проверка и обновление пароля
		if newPassword != "" {
			hashedPassword, err := HashPassword(newPassword)
			if err != nil {
				http.Error(w, "Ошибка при хешировании пароля", http.StatusInternalServerError)
				return
			}

			query := "UPDATE users SET password = $1 WHERE id = $2"
			_, err = Db.Exec(query, hashedPassword, userID)
			if err != nil {
				http.Error(w, "Ошибка при обновлении пароля", http.StatusInternalServerError)
				return
			}
		}

		// Обновление роли
		if newRole != "" {
			query := "UPDATE users SET role = $1 WHERE id = $2"
			_, err := Db.Exec(query, newRole, userID)
			if err != nil {
				http.Error(w, "Ошибка при обновлении роли", http.StatusInternalServerError)
				return
			}
		}

		// Перенаправление на страницу администратора
		http.Redirect(w, r, "/admin_page", http.StatusFound)
	}
}
