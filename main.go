package main

import (
	"log"
	"net/http"
	"uml/handlers"
)

func main() {
	handlers.OpenDatabase()   // Открываем подключение к базе данных
	defer handlers.Db.Close() // Закрываем соединение с базой данных при завершении работы

	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/main", handlers.Index)

	// Страницы для ролей
	http.HandleFunc("/admin_page", handlers.AdminPage)
	http.HandleFunc("/chief_editor_page", handlers.ChiefEditorPage)
	http.HandleFunc("/section_editor_page", handlers.SectionEditorPage)
	http.HandleFunc("/author_page", handlers.AuthorPage)

	// Дополнительные маршруты для функционала
	http.HandleFunc("/admin/employees", handlers.AdminEmployeesPage)
	http.HandleFunc("/update_user", handlers.UpdateUserHandler)
	http.HandleFunc("/chief_editor/assign_topics", handlers.AssignTopics)
	http.HandleFunc("/chief_editor/check_publications", handlers.CheckPublications)
	http.HandleFunc("/section_editor/assign_publications", handlers.AssignPublications)
	http.HandleFunc("/section_editor/edit_publication", handlers.EditPublication)
	http.HandleFunc("/section_editor/approve_publication", handlers.ApprovePublication)
	http.HandleFunc("/author/create_publication", handlers.CreatePublication)
	http.HandleFunc("/author/fix_comments", handlers.FixComments)

	log.Println("Сервер запущен на порту :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
