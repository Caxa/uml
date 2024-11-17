package main

import (
	"log"
	"net/http"

	"example.com/myproject/handlers"
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

	http.HandleFunc("/chief_editor/assign_topics", handlers.AssignTopicHandler)
	http.HandleFunc("/chief_editor/delete_topic", handlers.DeleteTopicHandler)
	http.HandleFunc("/chief_editor/check_publications", handlers.CheckPublicationsHandler)
	http.HandleFunc("/chief_editor/edit_draft", handlers.EditDraftHandler)
	http.HandleFunc("/approve_publication", handlers.ApprovePublicationHandler)
	http.HandleFunc("/request_revision", handlers.RequestRevisionHandler)

	http.HandleFunc("/section_editor/assign_publications", handlers.AssignPublications)
	http.HandleFunc("/section_editor/edit_publication", handlers.EditPublication)

	// дейстаивя админа
	http.HandleFunc("/add_user", handlers.AddUserHandler)
	http.HandleFunc("/delete_user", handlers.DeleteUserHandler)

	// автор
	http.HandleFunc("/author/create_publication", handlers.CreatePublicationHandler)
	http.HandleFunc("/author/fix_comments", handlers.FixCommentsHandler)
	http.HandleFunc("/author/create_publication_form", handlers.AuthorCreatePublicationFormHandler)
	http.HandleFunc("/author/edit_publication", handlers.EditPublicationHandler)
	http.HandleFunc("/author/update_publication", handlers.UpdatePublicationHandler)

	http.HandleFunc("/section_editor/publish_publication", handlers.PublishPublicationHandler)

	log.Println("Сервер запущен на порту :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
