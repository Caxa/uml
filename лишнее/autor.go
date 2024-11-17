package handlers

import (
	"log"
	"net/http"
	"strconv"
)

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

	// Получаем имя пользователя и роль
	userName, err := GetUserNameByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		log.Printf("Ошибка при получении имени пользователя: %v", err)
		return
	}

	role, err := GetUserRoleByIDFromDB(userID)
	if err != nil {
		http.Error(w, "Ошибка при получении роли пользователя из базы данных", http.StatusInternalServerError)
		log.Printf("Ошибка при получении роли пользователя: %v", err)
		return
	}

	// Временно отключаем получение тем из базы данных
	// topics, err := GetAvailableTopicsForAuthor()
	// if err != nil {
	//     http.Error(w, "Ошибка при получении тем из базы данных", http.StatusInternalServerError)
	//     log.Printf("Ошибка при получении тем из базы данных: %v", err)
	//     return
	// }

	// Временно передаём пустой список тем
	topics := []Topic{}

	// Структура данных для передачи в шаблон
	data := struct {
		UserID   int
		UserName string
		Role     string
		Topics   []Topic
	}{
		UserID:   userID,
		UserName: userName,
		Role:     role,
		Topics:   topics,
	}

	// Рендеринг шаблона
	err = TmplAuthor.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Ошибка выполнения шаблона: %v", err)
		return
	}
}

// Обработчик для исправления замечаний
func FixCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлечение данных из формы
	publicationIDStr := r.FormValue("publication_id")
	corrections := r.FormValue("corrections")

	if publicationIDStr == "" || corrections == "" {
		http.Error(w, "ID публикации и исправления обязательны", http.StatusBadRequest)
		return
	}

	publicationID, err := strconv.Atoi(publicationIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID публикации", http.StatusBadRequest)
		return
	}

	// Запрос на обновление публикации с исправлениями
	query := `UPDATE publications 
              SET corrections = $1, status = 'under_review' 
              WHERE id = $2`

	_, err = Db.Exec(query, corrections, publicationID)
	if err != nil {
		http.Error(w, "Ошибка при обновлении публикации", http.StatusInternalServerError)
		return
	}

	// Перенаправление на страницу с публикациями автора
	http.Redirect(w, r, "/author/publications", http.StatusSeeOther)
}

// Обработчик для страницы создания публикации
func AuthorCreatePublicationHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем список доступных тем, которые созданы редакторами
	rows, err := Db.Query("SELECT id, topic, department FROM user_topics ") // Пример запроса
	if err != nil {
		http.Error(w, "Ошибка при получении списка тем", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var topics []struct {
		ID         int
		Topic      string
		Department string
	}

	// Заполняем массив тем
	for rows.Next() {
		var topic struct {
			ID         int
			Topic      string
			Department string
		}
		if err := rows.Scan(&topic.ID, &topic.Topic, &topic.Department); err != nil {
			http.Error(w, "Ошибка при чтении данных", http.StatusInternalServerError)
			return
		}
		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Ошибка при обработке данных", http.StatusInternalServerError)
		return
	}

	// Передаем данные в шаблон
	data := struct {
		Topics []struct {
			ID         int
			Topic      string
			Department string
		}
	}{
		Topics: topics,
	}

	// Выполняем рендеринг шаблона
	if err := TmplAuthor.Execute(w, data); err != nil {
		http.Error(w, "Ошибка рендеринга страницы", http.StatusInternalServerError)
	}
}

// Обработчик для создания публикации
func CreatePublicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из формы
	topicIDStr := r.FormValue("topic_id")
	title := r.FormValue("title")
	content := r.FormValue("content")

	if topicIDStr == "" || title == "" || content == "" {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}

	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID темы", http.StatusBadRequest)
		return
	}

	// Запрос на добавление публикации в базу данных
	query := `INSERT INTO publications (title, content, topic_id, status) 
              VALUES ($1, $2, $3, 'draft')`

	_, err = Db.Exec(query, title, content, topicID)
	if err != nil {
		http.Error(w, "Ошибка при создании публикации", http.StatusInternalServerError)
		return
	}

	// Перенаправление на страницу с публикациями
	http.Redirect(w, r, "/author/publications", http.StatusSeeOther)
}

// Получение всех тем для автора
func GetAvailableTopicsForAuthor() ([]Topic, error) {
	var topics []Topic

	// Фильтруем темы по статусу 'assigned'
	query := "SELECT id, topic, department FROM topics WHERE status = 'assigned'"

	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var topic Topic

		// Проверка поля ID на корректность
		if err := rows.Scan(&topic.ID, &topic.Topic, &topic.Department); err != nil {
			return nil, err
		}

		// Пропускаем строки с некорректным ID
		if topic.ID == 0 {
			continue
		}

		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}
