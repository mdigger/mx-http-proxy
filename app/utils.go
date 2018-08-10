package app

import (
	"flag"
	"os"

	"github.com/mdigger/log"
)

func init() {
	// добавляем возможность изменения настроек вывода лога через параметры
	// запуска приложения
	flag.Var(log.Flag(), "log", "log `level`")
}

// IsDocker возвращает true, если приложение запущено в контейнере.
func IsDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

// Env возвращает строку из окружения с заданным именем или значение по
// умолчанию, если окружение не задано.
func Env(name, _default string) string {
	if s, ok := os.LookupEnv(name); ok {
		return s
	}
	return _default
}
