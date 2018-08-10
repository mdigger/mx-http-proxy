package app

import (
	"expvar"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/mdigger/log"
)

// Info описывает информацию о версии приложения.
type Info struct {
	Name    string     `json:"name"`              // название приложения
	Version string     `json:"version,omitempty"` // версия
	Commit  string     `json:"git,omitempty"`     // идентификатор коммита git
	Date    *time.Time `json:"build,omitempty"`   // дата сборки
	Started time.Time  `json:"-"`                 // время старта сервиса
}

// info описывает информацию о приложении
var info = Info{
	Name:    "TEST-APP",
	Version: "1.0",
	Started: time.Now().UTC(),
}

// инициализируем отдачу информации о версии в expvar
func init() {
	expvar.Publish("app", expvar.Func(func() interface{} {
		return info
	}))
	expvar.Publish("uptime", expvar.Func(func() interface{} {
		return int64(time.Since(info.Started))
	}))
	expvar.Publish("Goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
}

// Agent задает строку, которая может использоваться в качестве имени агента
// при http-запросах.
var Agent = "mdigger/1.0"

// Parse разбирает информацию о версии приложения с заполняет соответствующую
// структуру.
func Parse(name, version, commit, date string) {
	if name != "" {
		info.Name = name
	}
	// избавляемся от префиксов в версии
	if !strings.HasPrefix(version, "SNAPSHOT") {
		if version = strings.TrimPrefix(version, "v"); version != "" {
			if commit != "" {
				info.Commit = commit
				version = strings.TrimSuffix(
					strings.TrimSuffix(version, commit), "-g")
			}
			info.Version = version
		}
	}
	// разбираем дату сборки
	if date != "" {
		for _, dateTemplate := range []string{
			time.RFC3339,
			"2006-01-02T15:04",
			"2006-01-02",
		} {
			if paserdDate, err := time.Parse(dateTemplate, date); err == nil {
				info.Date = &paserdDate
				break
			}
		}
	}
	// формируем строку с агентом
	Agent = fmt.Sprintf("%s/%s", info.Name, info.Version)
	if info.Commit != "" {
		Agent += fmt.Sprintf(" (%s)", commit)
	}
}

// logInfo возвращает информацию о версии приложения в виде списка полей лога.
func (info Info) logInfo() []log.Field {
	var logInfo = []log.Field{log.Field{Name: "name", Value: info.Name}}
	if info.Version != "" {
		logInfo = append(logInfo, log.Field{Name: "version", Value: info.Version})
	}
	if info.Date != nil && !info.Date.IsZero() {
		var dateFormat = "2006-01-02T15:04"
		if hour, min, sec := info.Date.Clock(); hour == 0 && min == 0 && sec == 0 {
			dateFormat = "2006-01-02"
		}
		logInfo = append(logInfo, log.Field{Name: "build",
			Value: info.Date.Local().Format(dateFormat)})
	}
	if info.Commit != "" {
		logInfo = append(logInfo, log.Field{Name: "git", Value: info.Commit})
	}
	return logInfo
}

// LogInfo возвращает информацию о версии приложения в виде списка полей лога.
func LogInfo() []log.Field {
	return info.logInfo()
}

// Get возвращает копию информации о версии приложения.
func Get() Info {
	return info
}

// Uptime возвращает время работы приложения.
func Uptime() time.Duration {
	return time.Since(info.Started)
}
