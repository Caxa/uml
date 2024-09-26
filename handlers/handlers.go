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

// Структура пользователя
type User struct {
	IDuser   int
	Login    string
	Password string
	Role     string
}

var Db *sql.DB
var Tmpl1 = template.Must(template.ParseFiles("templates/register.html"))
var Tmpl2 = template.Must(template.ParseFiles("templates/upload1.html"))                         // Главная страница
var TmplAdmin = template.Must(template.ParseFiles("templates/admin_page.html"))                  // Шаблон страницы администратора
var TmplChiefEditor = template.Must(template.ParseFiles("templates/chief_editor_page.html"))     // Шаблон страницы главного редактора
var TmplSectionEditor = template.Must(template.ParseFiles("templates/section_editor_page.html")) // Шаблон страницы редактора раздела
var TmplAuthor = template.Must(template.ParseFiles("templates/author_page.html"))                // Шаблон страницы автора

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
		UserID   int
		UserName string
		Role     string
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
	if r.Method == http.MethodGet {
		Tmpl1.Execute(w, nil) // Отображение страницы регистрации
	} else if r.Method == http.MethodPost {
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

// Получаем список всех пользователей
func GetAllUsers() ([]User, error) {
	query := "SELECT id, login, role FROM users"
	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.IDuser, &user.Login, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// Обработчик для страницы администратора
func AdminPage(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "ID пользователя не передан", http.StatusBadRequest)
		return
	}

	// Преобразуем строковый параметр id в целое число
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	// Получаем имя пользователя по его ID
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	// Получаем роль пользователя по его ID
	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении роли пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для передачи в шаблон
	data := struct {
		UserID   int
		UserName string
		Role     string
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
	}

	// Выполняем рендеринг шаблона для страницы администратора
	err = TmplAdmin.Execute(w, data)
	if err != nil {
		// Проверка, были ли уже отправлены заголовки
		if w.Header().Get("Content-Type") == "" {
			log.Printf("привет")
		}
		return
	}
}

// Обработчик для отображения списка сотрудников
func AdminEmployeesPage(w http.ResponseWriter, r *http.Request) {
	// Получаем список всех пользователей
	users, err := GetAllUsers()
	if err != nil {
		http.Error(w, "Ошибка получения списка сотрудников: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Предположим, что текущий пользователь — это администратор, информация о котором нам также нужна
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	// Получаем информацию о текущем пользователе
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка получения имени пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка получения роли пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Подготовка данных для передачи в шаблон
	data := struct {
		UserID   int
		UserName string
		Role     string
		Users    []User
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
		Users:    users,
	}

	// Выполняем рендеринг шаблона с обновленной структурой данных
	err = TmplAdmin.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Обработчик для обновления данных пользователя
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Неправильный метод запроса", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из формы
	userID := r.FormValue("user_id")
	newLogin := r.FormValue("new_login")
	newPassword := r.FormValue("new_password")
	newRole := r.FormValue("new_role")

	// Проверяем ID пользователя
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	// Формируем SQL-запрос для обновления
	query := "UPDATE users SET "
	params := []interface{}{}
	paramIndex := 1

	if newLogin != "" {
		query += "login = $" + strconv.Itoa(paramIndex) + ", "
		params = append(params, newLogin)
		paramIndex++
	}

	if newPassword != "" {
		hashedPassword, err := HashPassword(newPassword)
		if err != nil {
			http.Error(w, "Ошибка при хешировании пароля", http.StatusInternalServerError)
			return
		}
		query += "password = $" + strconv.Itoa(paramIndex) + ", "
		params = append(params, hashedPassword)
		paramIndex++
	}

	if newRole != "" {
		query += "role = $" + strconv.Itoa(paramIndex) + " "
		params = append(params, newRole)
	} else {
		// Убираем последнюю запятую
		query = query[:len(query)-2] + " "
	}

	query += "WHERE id = $" + strconv.Itoa(paramIndex)
	params = append(params, userIDInt)

	// Выполняем запрос
	_, err = Db.Exec(query, params...)
	if err != nil {
		http.Error(w, "Ошибка при обновлении данных пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin_page", http.StatusSeeOther)
}

// Обработчик для страницы главного редактора
func ChiefEditorPage(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "ID пользователя не передан", http.StatusBadRequest)
		return
	}

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
		UserID   int
		UserName string
		Role     string
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
	}

	err = TmplChiefEditor.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Обработчик для страницы редактора раздела
func SectionEditorPage(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "ID пользователя не передан", http.StatusBadRequest)
		return
	}

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
		UserID   int
		UserName string
		Role     string
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
	}

	err = TmplSectionEditor.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Обработчик для страницы автора
func AuthorPage(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "ID пользователя не передан", http.StatusBadRequest)
		return
	}

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
		UserID   int
		UserName string
		Role     string
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
	}

	err = TmplAuthor.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}

// Остальные обработчики
func AssignTopics(w http.ResponseWriter, r *http.Request) {
	// Логика для назначения тем
}

func CheckPublications(w http.ResponseWriter, r *http.Request) {
	// Логика для проверки публикаций
}

func AssignPublications(w http.ResponseWriter, r *http.Request) {
	// Логика для назначения публикаций
}

func EditPublication(w http.ResponseWriter, r *http.Request) {
	// Логика для редактирования публикаций
}

func ApprovePublication(w http.ResponseWriter, r *http.Request) {
	// Логика для утверждения публикаций
}

func CreatePublication(w http.ResponseWriter, r *http.Request) {
	// Логика для создания публикации
}

func FixComments(w http.ResponseWriter, r *http.Request) {
	// Логика для исправления комментариев
}
