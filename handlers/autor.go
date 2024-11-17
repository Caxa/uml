package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"
)

// Получение всех тем для автора
func GetAvailableTopicsForAuthor() ([]Topic, error) {
	var topics []Topic

	// Фильтруем темы по статусу 'assigned'
	query := "SELECT id, topic, department FROM user_topics "

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

// GetTopicNameByID получает название темы по её ID из базы данных
func GetTopicNameByID(topicID int) (string, error) {
	var topicName string

	query := `SELECT topic FROM user_topics WHERE id = $1`
	err := Db.QueryRow(query, topicID).Scan(&topicName)
	if err != nil {
		return "", err
	}

	return topicName, nil
}

func GetAuthorPublications(authorID int) ([]Publication, error) {
	var publications []Publication
	query := "SELECT id, title, status FROM publications WHERE author_id = $1"
	rows, err := Db.Query(query, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var publication Publication
		if err := rows.Scan(&publication.ID, &publication.Title, &publication.Status); err != nil {
			return nil, err
		}
		publications = append(publications, publication)
	}
	return publications, nil
}

// GetTopicsByAuthorID получает список доступных тем для автора
func GetTopicsByAuthorID(authorID int) ([]Topic, error) {
	var topics []Topic

	// Запрос для получения списка тем, связанных с автором
	query := "SELECT id, topic, department FROM user_topics "
	rows, err := Db.Query(query, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.ID, &topic.Topic, &topic.Department); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}

	// Проверка ошибок после завершения итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

// Функция для получения всех доступных тем
func GetAllTopics() ([]Topic, error) {
	var topics []Topic
	query := "SELECT id, topic, department FROM user_topics"
	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.ID, &topic.Topic, &topic.Department); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return topics, nil
}

func FixCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	publicationIDStr := r.FormValue("publication_id")
	corrections := r.FormValue("corrections")

	if publicationIDStr == "" || corrections == "" {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}

	publicationID, err := strconv.Atoi(publicationIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID публикации", http.StatusBadRequest)
		return
	}

	query := `UPDATE publications 
              SET corrections = $1, status = 'under_review' 
              WHERE id = $2`

	_, err = Db.Exec(query, corrections, publicationID)
	if err != nil {
		http.Error(w, "Ошибка при обновлении публикации", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/author/publications", http.StatusSeeOther)
}

func AuthorCreatePublicationFormHandler(w http.ResponseWriter, r *http.Request) {
	authorIDStr := r.URL.Query().Get("author_id")
	authorID, err := strconv.Atoi(authorIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор автора", http.StatusBadRequest)
		return
	}

	topicIDStr := r.URL.Query().Get("topic_id")
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		http.Error(w, "Неверный формат topic_id", http.StatusBadRequest)
		return
	}

	topicName, err := GetTopicNameByID(topicID)
	if err != nil {
		http.Error(w, "Ошибка получения названия темы: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		TopicID   int
		TopicName string
		AuthorID  int
	}{
		TopicID:   topicID,
		TopicName: topicName,
		AuthorID:  authorID,
	}

	err = TmplCreatePublication.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Обработчик для создания публикации
func AuthorCreatePublicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получение ID автора из сессии
	session, _ := store.Get(r, "session-name")
	authorID, ok := session.Values["author_id"].(int)
	if !ok || authorID == 0 {
		http.Error(w, "Автор не авторизован", http.StatusUnauthorized)
		return
	}

	// Получение данных из формы
	topicIDStr := r.FormValue("topic_id")
	title := r.FormValue("title")
	content := r.FormValue("content")

	// Проверка обязательных полей
	if title == "" || content == "" {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}

	// Преобразование topic_id
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID темы", http.StatusBadRequest)
		return
	}

	// Текущее время для полей created_at и updated_at
	currentTime := time.Now()

	// SQL-запрос для вставки записи, включая created_at и updated_at
	query := `INSERT INTO publications 
                (title, content, topic_id, author_id, status, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, 'draft', $5, $5)`

	// Выполнение запроса с параметрами
	result, err := Db.Exec(query, title, content, topicID, authorID, currentTime)
	if err != nil {
		log.Printf("Ошибка при выполнении запроса вставки: %v", err)
		http.Error(w, "Ошибка при создании публикации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получение ID созданной публикации для проверки результата
	pubID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Ошибка получения ID вставленной публикации: %v", err)
	} else {
		log.Printf("Публикация успешно создана с ID: %d", pubID)
	}

	// Перенаправление на страницу автора
	http.Redirect(w, r, "/author?id="+strconv.Itoa(authorID), http.StatusSeeOther)
}
func CreatePublicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получение ID автора из формы
	authorIDStr := r.FormValue("author_id")
	authorID, err := strconv.Atoi(authorIDStr)
	if err != nil || authorID == 0 {
		http.Error(w, "Неверный идентификатор автора", http.StatusBadRequest)
		return
	}

	topicIDStr := r.FormValue("topic_id")
	title := r.FormValue("title")
	content := r.FormValue("content")

	// Проверка обязательных полей
	if title == "" || content == "" || topicIDStr == "" {
		http.Error(w, "Все поля обязательны", http.StatusBadRequest)
		return
	}

	// Преобразование topic_id в число
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор темы", http.StatusBadRequest)
		return
	}

	// Проверка существования `topic_id`
	exists, err := CheckTopicExists(topicID)
	if err != nil {
		http.Error(w, "Ошибка проверки темы", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Указанная тема не существует", http.StatusBadRequest)
		return
	}

	// Создание публикации
	query := `INSERT INTO publications (title, content, topic_id, author_id, status, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, 'draft', $5, $5)`
	currentTime := time.Now()
	_, err = Db.Exec(query, title, content, topicID, authorID, currentTime)
	if err != nil {
		http.Error(w, "Ошибка при создании публикации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Перенаправление на страницу автора после создания публикации
	redirectURL := "/author_page?id=" + strconv.Itoa(authorID)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// CheckTopicExists проверяет, существует ли тема с данным ID в таблице user_topics.
func CheckTopicExists(topicID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_topics WHERE id = $1)`

	// Выполняем запрос в базу данных, который возвращает true, если запись существует
	err := Db.QueryRow(query, topicID).Scan(&exists)
	if err != nil {
		return false, err // возвращаем ошибку, если возникла проблема с запросом
	}

	return exists, nil // возвращаем результат проверки
}
func AuthorPage(w http.ResponseWriter, r *http.Request) {
	authorIDStr := r.URL.Query().Get("id")
	authorID, err := strconv.Atoi(authorIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор автора", http.StatusBadRequest)
		return
	}

	topics, err := GetAllTopics()
	if err != nil {
		http.Error(w, "Ошибка при получении тем", http.StatusInternalServerError)
		return
	}

	// Получение публикаций автора
	publications, err := GetAuthorPublications(authorID)
	if err != nil {
		http.Error(w, "Ошибка при получении публикаций автора", http.StatusInternalServerError)
		return
	}

	data := struct {
		AuthorID     int
		Topics       []Topic
		Publications []Publication
	}{
		AuthorID:     authorID,
		Topics:       topics,
		Publications: publications,
	}

	if err := TmplAuthor.Execute(w, data); err != nil {
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}
