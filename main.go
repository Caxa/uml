package main

import (
	"log"
	"net/http"

	"uml/handlers" // Замените на ваше имя модуля
)

func main() {
	handlers.OpenDatabase()   // Открываем подключение к базе данных
	defer handlers.Db.Close() // Закрываем соединение с базой данных при завершении работы

	http.HandleFunc("/", handlers.Home)      // Главная страница
	http.HandleFunc("/main", handlers.Index) // Страница пользователя
	//http.HandleFunc("/create_user", handlers.CreateUserPage) // Страница создания пользователя
	http.HandleFunc("/register", handlers.RegisterHandler) // Обработчик регистрации
	http.HandleFunc("/admin_page", handlers.AdminPage)     // Страница для администратора
	http.HandleFunc("/update_user", handlers.UpdateUser)   // Обработчик обновления данных пользователя
	//http.HandleFunc("/assign_role", handlers.AssignRolePage) // Страница назначения ролей администратора

	log.Println("Сервер запущен на порту :8080")
	log.Fatal(http.ListenAndServe(":8080", nil)) // Запускаем сервер
}
