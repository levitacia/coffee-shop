package main

import (
	"log/slog"
	"os"
	"sso/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// TODO: инициализировать объект конфига
	cfg := config.MustLoad()

	// TODO: инициализировать логгер (slog?)
	log := setupLogger(cfg.Env)
	log.Info("starting application", slog.String("env", cfg.Env), slog.Any("cfg", cfg))

	// TODO: инициализация приложения (app)

	// слои: транспортный слой (обработчик запросов, примиает запросы и обращается к сервисному слою),
	// сервисный слой (бизнес логика приложения (можно назвать как бинарник проекта, так и объекты из сервисного слоя)),
	// слой работы с данными

	// TODO: запустить gRPC-сервер приложения
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log

}
