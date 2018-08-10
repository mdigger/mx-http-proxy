package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mdigger/log"
)

// Задаем название и метаданные описания проекта по умолчанию.
// Для изменения этих значений на этапе сборки можно использовать возможность
// передачи параметров компилятору:
// 	-ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT) -X main.date=$(DATE)"
var (
	appName = "MX-HTTP-Proxy"
	version string // версия приложения
	commit  string // идентификатор коммита git
	date    string // дата сборки

	appAgent  string             // используется как строка с именем сервиса
	logInfo   []log.Field        // информация о приложении для вывода в лог
	startTime = time.Now().UTC() // время старта сервиса
)

func init() {
	logInfo = []log.Field{log.Field{Name: "name", Value: appName}}
	if !strings.HasPrefix(version, "SNAPSHOT") {
		if version = strings.TrimPrefix(version, "v"); version != "" {
			if commit != "" {
				version = strings.TrimSuffix(
					strings.TrimSuffix(version, commit), "-g")
			}
			logInfo = append(logInfo, log.Field{Name: "version", Value: version})
		}
	}
	if date != "" {
		var dateField = log.Field{Name: "build", Value: date}
		if d, _ := time.Parse(time.RFC3339, date); !d.IsZero() {
			dateField.Value = d.Local().Format("2006-01-02T15:04")
		}
		logInfo = append(logInfo, dateField)
	}
	appAgent = fmt.Sprintf("%s/%s", appName, version)
	if commit != "" {
		appAgent += fmt.Sprintf(" (%s)", commit)
		logInfo = append(logInfo, log.Field{Name: "git", Value: commit})
	}
	// добавляем возможность изменения настроек вывода лога через параметры
	// запуска приложения
	flag.Var(log.Flag(), "log", "log `level`")
}

// isDocker возвращает true, если приложение запущено в контейнере.
func isDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}
