package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Обработчик для страницы редактора отдела
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

	// Получаем список публикаций
	publications, err := GetPublications()
	if err != nil {
		http.Error(w, "Ошибка при получении публикаций", http.StatusInternalServerError)
		return
	}

	// Создаем данные для шаблона
	data := struct {
		UserID       int
		UserName     string
		Role         string
		Publications []Publication
	}{
		UserID:       userID,
		UserName:     userName,
		Role:         role,
		Publications: publications,
	}

	// Выполняем шаблон
	if err = TmplSectionEditor.Execute(w, data); err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		return
	}
}

// Обработчик разрешения выкладки публикации
func AllowPublicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.FormValue("article_id")
	editorIDStr := r.FormValue("editor_id")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор статьи", http.StatusBadRequest)
		return
	}

	// Проверка статуса публикации
	var status string
	query := `SELECT status FROM publications WHERE id = $1`
	err = Db.QueryRow(query, articleID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Публикация не найдена", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка при проверке статуса публикации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем, одобрена ли публикация
	if status != "approved" {
		http.Error(w, "Публикация должна быть одобрена перед выкладкой", http.StatusBadRequest)
		return
	}

	// Обновляем статус на "ready_for_publication" и устанавливаем is_published = TRUE
	updateQuery := `UPDATE publications SET status = 'ready_for_publication', is_published = TRUE, updated_at = $1 WHERE id = $2`
	_, err = Db.Exec(updateQuery, time.Now(), articleID)
	if err != nil {
		http.Error(w, "Ошибка при обновлении публикации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Перенаправление обратно на страницу редактора отдела
	http.Redirect(w, r, "/section_editor_page?id="+editorIDStr, http.StatusSeeOther)
}

// Обработчик для выкладки публикации
func PublishPublicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем идентификатор статьи и редактора
	articleIDStr := r.FormValue("article_id")
	editorIDStr := r.FormValue("editor_id")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор статьи", http.StatusBadRequest)
		return
	}

	// Проверка статуса публикации
	var status string
	query := `SELECT status FROM publications WHERE id = $1`
	err = Db.QueryRow(query, articleID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Публикация не найдена", http.StatusNotFound)
			return
		}
		http.Error(w, "Ошибка при проверке статуса публикации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем, имеет ли публикация статус "approved"
	if status != "approved" {
		http.Error(w, "Публикация должна быть одобрена перед выкладкой", http.StatusBadRequest)
		return
	}

	// Обновляем статус публикации на "published" и устанавливаем флаг is_published = TRUE
	updateQuery := `UPDATE publications SET status = 'published', is_published = TRUE, updated_at = $1 WHERE id = $2`
	_, err = Db.Exec(updateQuery, time.Now(), articleID)
	if err != nil {
		http.Error(w, "Ошибка при выкладке публикации: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Перенаправление обратно на страницу редактора отдела
	http.Redirect(w, r, "/section_editor_page?id="+editorIDStr, http.StatusSeeOther)
}

// Обработчик для страницы редактора отдела
func SectionEditorPageHandler(w http.ResponseWriter, r *http.Request) {
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

	// Получаем публикации
	publications, err := GetPublications()
	if err != nil {
		http.Error(w, "Ошибка при получении публикаций", http.StatusInternalServerError)
		return
	}

	// Создаем данные для шаблона
	data := struct {
		UserID       int
		Publications []Publication
	}{
		UserID:       userID,
		Publications: publications,
	}

	// Выполняем шаблон
	if err := TmplSectionEditor.Execute(w, data); err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		return
	}
}
