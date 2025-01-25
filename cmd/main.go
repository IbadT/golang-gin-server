package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/zhashkevych/todo-app"
	"github.com/zhashkevych/todo-app/pkg/handler"
	"github.com/zhashkevych/todo-app/pkg/repository"
	"github.com/zhashkevych/todo-app/pkg/service"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	// Загружаем переменные окружения из .env файла (если есть)
	if err := godotenv.Load(); err != nil {
		logrus.Warnf("error loading .env file: %s", err.Error())
	}

	// Конфигурация базы данных из переменных окружения
	dbConfig := repository.Config{
		Host:     getEnv("DB_HOST", "db"),           // Имя сервиса в Docker Compose
		Port:     getEnv("DB_PORT", "5432"),         // Порт PostgreSQL
		Username: getEnv("DB_USER", "postgres"),     // Пользователь PostgreSQL
		DBName:   getEnv("DB_NAME", "postgres"),     // Имя базы данных
		Password: getEnv("DB_PASSWORD", "postgres"), // Пароль
		SSLMode:  getEnv("DB_SSLMODE", "disable"),   // Режим SSL
	}

	// Подключение к базе данных с повторными попытками
	db, err := repository.NewPostgresDB(dbConfig)

	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	// Инициализация репозиториев, сервисов и обработчиков
	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	// Запуск HTTP-сервера
	srv := new(todo.Server)
	go func() {
		if err := srv.Run(getEnv("APP_PORT", "8000"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occurred while running http server: %s", err.Error())
		}
	}()

	logrus.Info("TodoApp Started")

	// Ожидание сигнала для завершения работы
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Info("TodoApp Shutting Down")

	// Завершение работы сервера и закрытие соединения с базой данных
	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occurred on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logrus.Errorf("error occurred on db connection close: %s", err.Error())
	}
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
