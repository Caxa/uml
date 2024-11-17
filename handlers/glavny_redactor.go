package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Добавление новой темы
func AssignTopicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	editorIDStr := r.FormValue("editor_id")
	topic := r.FormValue("topic")
	department := r.FormValue("department")
	editorID, err := strconv.Atoi(editorIDStr)
	if err != nil {
		http.Error(w, "Неверный ID главного редактора", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO user_topics (editor_id, topic, department, assigned_at) VALUES ($1, $2, $3, $4)`
	_, err = Db.Exec(query, editorID, topic, department, time.Now())
	if err != nil {
		http.Error(w, "Ошибка при назначении темы", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/chief_editor_page?id="+editorIDStr, http.StatusSeeOther)
}

// Просмотр текущих тем
func ViewTopicsHandler(w http.ResponseWriter, r *http.Request) {
	editorIDStr := r.URL.Query().Get("editor_id")
	editorID, err := strconv.Atoi(editorIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор главного редактора", http.StatusBadRequest)
		return
	}

	topics, err := GetTopicsByEditorID(editorID)
	if err != nil {
		http.Error(w, "Ошибка при получении тем", http.StatusInternalServerError)
		return
	}

	data := struct {
		EditorID int
		Topics   []Topic
	}{
		EditorID: editorID,
		Topics:   topics,
	}

	if err := TmplChiefEditor.Execute(w, data); err != nil {
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}
func DeleteTopicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	topicIDStr := r.FormValue("topic_id")
	editorIDStr := r.FormValue("editor_id")

	// Преобразование идентификаторов
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор темы", http.StatusBadRequest)
		return
	}

	editorID, err := strconv.Atoi(editorIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор редактора", http.StatusBadRequest)
		return
	}

	// Выполнение запроса на удаление
	query := "DELETE FROM user_topics WHERE id = $1 AND editor_id = $2"
	result, err := Db.Exec(query, topicID, editorID)
	if err != nil {
		http.Error(w, "Ошибка при удалении темы", http.StatusInternalServerError)
		return
	}

	// Проверка количества удаленных записей
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Тема с ID %d для редактора %d не найдена", topicID, editorID)
		http.Error(w, "Тема не найдена", http.StatusNotFound)
		return
	}

	log.Printf("Тема с ID %d для редактора %d успешно удалена", topicID, editorID)

	// Перенаправление на страницу главного редактора
	http.Redirect(w, r, "/chief_editor_page?id="+editorIDStr, http.StatusSeeOther)
}

// Вспомогательная функция для получения тем
func GetTopicsByEditorID(editorID int) ([]Topic, error) {
	rows, err := Db.Query("SELECT id, topic, department FROM user_topics WHERE editor_id = $1", editorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topics := []Topic{}
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.ID, &topic.Topic, &topic.Department); err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return topics, nil
}

// Вспомогательная функция для получения подготовленных публикаций
func GetPreparedPublications() ([]Publication, error) {
	query := "SELECT id, title, content, department, created_at FROM publications WHERE status = 'draft'"
	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	publications := []Publication{}
	for rows.Next() {
		var pub Publication
		if err := rows.Scan(&pub.ID, &pub.Title, &pub.Content, &pub.Department, &pub.CreatedAt); err != nil {
			return nil, err
		}
		pub.Status = "draft" // Устанавливаем статус по умолчанию
		publications = append(publications, pub)
	}
	return publications, nil
}

// ChiefEditorPage обрабатывает отображение страницы главного редактора
func ChiefEditorPage(w http.ResponseWriter, r *http.Request) {
	// Получаем ID редактора из параметров запроса
	editorIDStr := r.URL.Query().Get("id")
	if editorIDStr == "" {
		http.Error(w, "ID пользователя не передан", http.StatusBadRequest)
		return
	}

	// Конвертируем editorID в целое число
	editorID, err := strconv.Atoi(editorIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя", http.StatusBadRequest)
		return
	}

	// Получаем имя и роль главного редактора
	userName, err := GetUserNameByIDFromDB(editorID)
	if err != nil {
		http.Error(w, "Ошибка при получении имени пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	role, err := GetUserRoleByIDFromDB(editorID)
	if err != nil {
		http.Error(w, "Ошибка при получении роли пользователя из базы данных", http.StatusInternalServerError)
		return
	}

	// Получаем темы для отображения на странице
	topics, err := GetTopicsByEditorID(editorID)
	if err != nil {
		http.Error(w, "Ошибка при получении списка тем из базы данных", http.StatusInternalServerError)
		return
	}

	// Получаем подготовленные публикации
	publications, err := GetPublications()
	if err != nil {
		http.Error(w, "Ошибка при получении публикаций", http.StatusInternalServerError)
		return
	}

	// Получаем черновики публикаций
	draftPublications, err := GetDraftPublications()
	if err != nil {
		http.Error(w, "Ошибка при получении черновиков", http.StatusInternalServerError)
		return
	}

	// Подготавливаем данные для передачи в шаблон
	data := struct {
		UserID            int
		EditorID          int
		UserName          string
		Role              string
		Topics            []Topic
		Publications      []Publication
		DraftPublications []Publication
	}{
		UserID:            editorID,
		EditorID:          editorID,
		UserName:          userName,
		Role:              role,
		Topics:            topics,
		Publications:      publications,
		DraftPublications: draftPublications,
	}

	// Выполняем рендеринг шаблона с передачей данных
	if err := TmplChiefEditor.Execute(w, data); err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
	}
}

// Проверка публикаций
func CheckPublicationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Неправильный метод запроса", http.StatusMethodNotAllowed)
		return
	}

	query := "SELECT id, title, content, created_at FROM publications WHERE status = 'draft'"
	rows, err := Db.Query(query)
	if err != nil {
		http.Error(w, "Ошибка при получении публикаций", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	publications := []struct {
		ID        int
		Title     string
		Content   string
		CreatedAt time.Time
	}{}

	for rows.Next() {
		var pub struct {
			ID        int
			Title     string
			Content   string
			CreatedAt time.Time
		}
		if err := rows.Scan(&pub.ID, &pub.Title, &pub.Content, &pub.CreatedAt); err != nil {
			http.Error(w, "Ошибка обработки публикаций", http.StatusInternalServerError)
			return
		}
		publications = append(publications, pub)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h2>Подготовленные публикации:</h2><ul>")
	for _, pub := range publications {
		fmt.Fprintf(w, "<li>%d - %s - %s</li>", pub.ID, pub.Title, pub.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Fprintln(w, "</ul>")
}

// Получение черновиков публикаций
func GetDraftPublications() ([]Publication, error) {
	query := "SELECT id, title, content, created_at FROM publications WHERE status = 'draft'"
	rows, err := Db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	publications := []Publication{}
	for rows.Next() {
		var pub Publication
		if err := rows.Scan(&pub.ID, &pub.Title, &pub.Content, &pub.CreatedAt); err != nil {
			return nil, err
		}
		publications = append(publications, pub)
	}
	return publications, nil
}

// Обработчик для одобрения публикации
func ApprovePublicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.FormValue("article_id")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор статьи", http.StatusBadRequest)
		return
	}

	query := "UPDATE publications SET status = 'approved', updated_at = $1 WHERE id = $2"
	_, err = Db.Exec(query, time.Now(), articleID)
	if err != nil {
		http.Error(w, "Ошибка при обновлении статуса публикации", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/chief_editor_page?id="+r.FormValue("editor_id"), http.StatusSeeOther)
}

// Обработчик для отправки публикации на доработку
func RequestRevisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.FormValue("article_id")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор статьи", http.StatusBadRequest)
		return
	}

	query := "UPDATE publications SET status = 'revision', updated_at = $1 WHERE id = $2"
	_, err = Db.Exec(query, time.Now(), articleID)
	if err != nil {
		http.Error(w, "Ошибка при обновлении статуса публикации", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/chief_editor_page?id="+r.FormValue("editor_id"), http.StatusSeeOther)
}

// Изменение черновика публикации
func EditDraftHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.FormValue("article_id")
	title := r.FormValue("title")
	content := r.FormValue("content")

	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор статьи", http.StatusBadRequest)
		return
	}

	query := `UPDATE publications SET title = $1, content = $2, updated_at = $3 WHERE id = $4 AND status = 'draft'`
	_, err = Db.Exec(query, title, content, time.Now(), articleID)
	if err != nil {
		log.Printf("Ошибка обновления черновика: %v", err)
		http.Error(w, "Ошибка при обновлении черновика", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/chief_editor_page?id="+r.FormValue("editor_id"), http.StatusSeeOther)
}

func GetPublications() ([]Publication, error) {
	var publications []Publication

	query := `SELECT id, title, content, topic_id, author_id, status, created_at, updated_at FROM publications`
	rows, err := Db.Query(query)
	if err != nil {
		log.Printf("Ошибка выполнения SQL-запроса: %v", err)
		return nil, fmt.Errorf("ошибка при выполнении запроса к базе данных")
	}
	defer rows.Close()

	for rows.Next() {
		var pub Publication
		if err := rows.Scan(&pub.ID, &pub.Title, &pub.Content, &pub.TopicID, &pub.AuthorID, &pub.Status, &pub.CreatedAt, &pub.UpdatedAt); err != nil {
			log.Printf("Ошибка сканирования данных публикации: %v", err)
			return nil, fmt.Errorf("ошибка при чтении данных")
		}
		publications = append(publications, pub)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Ошибка после завершения rows.Next(): %v", err)
		return nil, fmt.Errorf("ошибка после завершения чтения данных")
	}

	return publications, nil
}
