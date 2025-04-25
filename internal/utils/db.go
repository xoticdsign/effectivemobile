package utils

import "log/slog"

func LogDBDebug(log *slog.Logger, source string, op string, err error) {
	log.Debug(
		"транзакция прервана, rollback",
		slog.String("source", source),
		slog.String("op", op),
		slog.Any("error", err),
	)
}
