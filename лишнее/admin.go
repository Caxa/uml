// Обработчик для главной страницы редактора
func ChiefEditorPage(w http.ResponseWriter, r *http.Request) {
	// Получение данных редактора
	editorIDStr := r.URL.Query().Get("editor_id")
	editorID, err := strconv.Atoi(editorIDStr)
	if err != nil {
		http.Error(w, "Неверный идентификатор редактора", http.StatusBadRequest)
		return
	}

	// Запрос тем для текущего редактора
	rows, err := Db.Query("SELECT id, topic, department FROM user_topics WHERE editor_id = $1", editorID)
	if err != nil {
		http.Error(w, "Ошибка при получении тем: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Заполняем список тем
	var topics []Topic
	for rows.Next() {
		var topic Topic
		if err := rows.Scan(&topic.ID, &topic.Topic, &topic.Department); err != nil {
			http.Error(w, "Ошибка чтения данных из базы: "+err.Error(), http.StatusInternalServerError)
			return
		}
		topics = append(topics, topic)
	}

	// Данные для передачи в шаблон
	data := ChiefEditorData{
		UserID:   123, // пример значения, необходимо заменить на реальное
		EditorID: editorID,
		UserName: "Главный редактор",
		Role:     "chief_editor",
		Topics:   topics,
	}

	// Рендеринг шаблона
	err = TmplChiefEditor.Execute(w, data)
	if err != nil {
		http.Error(w, "Ошибка выполнения шаблона: "+err.Error(), http.StatusInternalServerError)
	}
}
