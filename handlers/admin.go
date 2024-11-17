package handlers

import (
	"net/http"
	"strconv"
)

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

// Обработчик для добавления нового пользователя
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Неправильный метод запроса", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из формы
	login := r.FormValue("login")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// Проверяем, что логин, пароль и роль не пустые
	if login == "" || password == "" || role == "" {
		http.Error(w, "Поля логин, пароль и роль обязательны", http.StatusBadRequest)
		return
	}

	// Хешируем пароль
	hashedPassword, err := HashPassword(password)
	if err != nil {
		http.Error(w, "Ошибка при хешировании пароля", http.StatusInternalServerError)
		return
	}

	// Выполняем SQL-запрос для добавления пользователя
	query := "INSERT INTO users (login, password, role) VALUES ($1, $2, $3)"
	_, err = Db.Exec(query, login, hashedPassword, role)
	if err != nil {
		http.Error(w, "Ошибка добавления пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем ID текущего администратора из формы
	adminIDStr := r.FormValue("admin_id")
	adminID, err := strconv.Atoi(adminIDStr)
	if err != nil || adminID <= 0 {
		http.Error(w, "Неверный идентификатор администратора", http.StatusBadRequest)
		return
	}

	// Перенаправляем администратора на страницу с обновленным списком пользователей
	http.Redirect(w, r, "/admin_page?id="+strconv.Itoa(adminID), http.StatusSeeOther)
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

	// Получаем список сотрудников
	users, err := GetAllUsers()
	if err != nil {
		http.Error(w, "Ошибка получения списка сотрудников", http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для передачи в шаблон
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

	// Выполняем рендеринг шаблона для страницы администратора
	err = TmplAdmin.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		return
	}
}


// Обработчик для удаления пользователя
// Обработчик для удаления пользователя
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Неправильный метод запроса", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID пользователя, которого нужно удалить
	userIDStr := r.FormValue("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	// Удаляем пользователя из базы данных
	query := "DELETE FROM users WHERE id = $1"
	_, err = Db.Exec(query, userID)
	if err != nil {
		http.Error(w, "Ошибка при удалении пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем ID текущего администратора
	adminIDStr := r.FormValue("admin_id")
	adminID, err := strconv.Atoi(adminIDStr)
	if err != nil || adminID <= 0 {
		http.Error(w, "Неверный идентификатор администратора", http.StatusBadRequest)
		return
	}

	// Перенаправляем администратора на страницу с обновленным списком пользователей
	http.Redirect(w, r, "/admin_page?id="+strconv.Itoa(adminID), http.StatusSeeOther)
}
