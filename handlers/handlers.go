package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

// Структура пользователя
type User struct {
	IDuser   int
	Login    string
	Password string
	Role     string
}

type Publication struct {
	ID         int
	Title      string
	Content    string
	TopicID    int
	AuthorID   int
	Status     string
	Department string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Topic struct {
	ID         int
	Title      string
	Topic      string
	Department string
}

type ChiefEditorData struct {
	UserID   int
	EditorID int
	UserName string
	Role     string
	Topics   []Topic
}

var Db *sql.DB
var Tmpl1 = template.Must(template.ParseFiles("templates/register.html"))
var TmplCatalog = template.Must(template.ParseFiles("templates/upload1.html"))                   // Главная страница
var TmplAdmin = template.Must(template.ParseFiles("templates/admin_page.html"))                  // Шаблон страницы администратора
var TmplChiefEditor = template.Must(template.ParseFiles("templates/chief_editor_page.html"))     // Шаблон страницы главного редактора
var TmplSectionEditor = template.Must(template.ParseFiles("templates/section_editor_page.html")) // Шаблон страницы редактора раздела
var TmplAuthor = template.Must(template.ParseFiles("templates/author_page.html"))
var TmplCreatePublication = template.Must(template.ParseFiles("templates/сreate_publication.html"))

var store = sessions.NewCookieStore([]byte("секретный-ключ"))

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

	err = TmplCatalog.Execute(w, data)
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

// Назначение публикаций автору
func AssignPublications(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.FormValue("user_id")
	publicationIDStr := r.FormValue("publication_id")

	userID, err := strconv.Atoi(userIDStr)
	publicationID, err := strconv.Atoi(publicationIDStr)

	query := "INSERT INTO user_publications (user_id, publication_id) VALUES ($1, $2)"
	_, err = Db.Exec(query, userID, publicationID)
	if err != nil {
		http.Error(w, "Ошибка при назначении публикации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/chief_editor_page", http.StatusSeeOther)
}

// Отображение каталога публикаций
func CatalogPage(w http.ResponseWriter, r *http.Request) {
	rows, err := Db.Query("SELECT id, title, content FROM publications")
	if err != nil {
		http.Error(w, "Ошибка при получении публикаций: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var publications []Publication
	for rows.Next() {
		var pub Publication
		err := rows.Scan(&pub.ID, &pub.Title, &pub.Content)
		if err != nil {
			http.Error(w, "Ошибка при обработке публикации: "+err.Error(), http.StatusInternalServerError)
			return
		}
		publications = append(publications, pub)
	}

	data := struct {
		Publications []Publication
	}{
		Publications: publications,
	}

	err = TmplCatalog.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона каталога: "+err.Error(), http.StatusInternalServerError)
	}
}

// Редактирование публикации
func EditPublication(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		pubID := r.FormValue("id")
		title := r.FormValue("title")
		content := r.FormValue("content")

		query := "UPDATE publications SET title = $1, content = $2 WHERE id = $3"
		_, err := Db.Exec(query, title, content, pubID)
		if err != nil {
			http.Error(w, "Ошибка при редактировании публикации: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/catalog_page", http.StatusSeeOther)
	}
}

// Удаление публикации
func DeletePublication(w http.ResponseWriter, r *http.Request) {
	pubID := r.URL.Query().Get("id")
	query := "DELETE FROM publications WHERE id = $1"
	_, err := Db.Exec(query, pubID)
	if err != nil {
		http.Error(w, "Ошибка при удалении публикации: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/catalog_page", http.StatusSeeOther)
}

// Проверка роли пользователя
func CheckUserRole(userID int) (string, error) {
	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		return "", err
	}
	return role, nil
}

func FixComments() {}
