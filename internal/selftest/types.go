// internal/selftest/types.go
// Типы и интерфейсы для self-test команды
package selftest

import (
	"fmt"
	"time"
)

// Result статус результата проверки
type Result string

const (
	ResultPass Result = "PASS" // Успешно
	ResultFail Result = "FAIL" // Ошибка
	ResultWarn Result = "WARN" // Предупреждение
	ResultSkip Result = "SKIP" // Пропущено
)

// String возвращает emoji представление результата
func (r Result) String() string {
	switch r {
	case ResultPass:
		return "✓"
	case ResultFail:
		return "✗"
	case ResultWarn:
		return "⚠"
	case ResultSkip:
		return "⊘"
	default:
		return "?"
	}
}

// Color возвращает ANSI цвет для результата
func (r Result) Color() string {
	switch r {
	case ResultPass:
		return "\033[32m" // Зелёный
	case ResultFail:
		return "\033[31m" // Красный
	case ResultWarn:
		return "\033[33m" // Жёлтый
	case ResultSkip:
		return "\033[90m" // Серый
	default:
		return "\033[0m"
	}
}

// ResetColor сбрасывает цвет
func (r Result) ResetColor() string {
	return "\033[0m"
}

// CheckResult результат одной проверки
type CheckResult struct {
	Name        string        `json:"name"`        // Название проверки
	Category    string        `json:"category"`    // Категория (config, api, tests, etc)
	Result      Result        `json:"result"`      // Статус
	Message     string        `json:"message"`     // Человекочитаемое сообщение
	Details     string        `json:"details"`     // Детали (опционально)
	Duration    time.Duration `json:"duration"`    // Время выполнения
	Error       error         `json:"-"`           // Ошибка (не сериализуется)
	CanFix      bool          `json:"can_fix"`     // Можно ли автоматически исправить
	FixCommand  string        `json:"fix_command"` // Команда для исправления
}

// Report полный отчёт о self-test
type Report struct {
	Timestamp   time.Time     `json:"timestamp"`   // Время запуска
	Version     string        `json:"version"`     // Версия gotr
	Commit      string        `json:"commit"`      // Git commit
	GoVersion   string        `json:"go_version"`  // Версия Go
	Platform    string        `json:"platform"`    // OS/Arch
	Checks      []CheckResult `json:"checks"`      // Все проверки
	Duration    time.Duration `json:"duration"`    // Общее время
	Health      Result        `json:"health"`      // Общий статус
	TotalPassed int           `json:"total_passed"`
	TotalFailed int           `json:"total_failed"`
	TotalWarn   int           `json:"total_warn"`
	TotalSkip   int           `json:"total_skip"`
}

// CalculateHealth вычисляет общий статус здоровья
func (r *Report) CalculateHealth() {
	hasFail := false
	hasWarn := false

	for _, c := range r.Checks {
		switch c.Result {
		case ResultPass:
			r.TotalPassed++
		case ResultFail:
			r.TotalFailed++
			hasFail = true
		case ResultWarn:
			r.TotalWarn++
			hasWarn = true
		case ResultSkip:
			r.TotalSkip++
		}
	}

	if hasFail {
		r.Health = ResultFail
	} else if hasWarn {
		r.Health = ResultWarn
	} else {
		r.Health = ResultPass
	}
}

// OverallStatus строка статуса
func (r *Report) OverallStatus() string {
	switch r.Health {
	case ResultPass:
		return "Healthy"
	case ResultWarn:
		return "Degraded"
	case ResultFail:
		return "Unhealthy"
	default:
		return "Unknown"
	}
}

// Checker интерфейс для проверок
type Checker interface {
	Name() string
	Category() string
	Check() CheckResult
}

// Runner запускает все проверки
type Runner struct {
	checkers []Checker
}

// NewRunner создаёт новый Runner
func NewRunner() *Runner {
	return &Runner{
		checkers: make([]Checker, 0),
	}
}

// Register добавляет проверку
func (r *Runner) Register(c Checker) {
	r.checkers = append(r.checkers, c)
}

// Run выполняет все проверки
func (r *Runner) Run() *Report {
	report := &Report{
		Timestamp: time.Now(),
		Checks:    make([]CheckResult, 0, len(r.checkers)),
	}

	start := time.Now()

	for _, checker := range r.checkers {
		checkStart := time.Now()
		result := checker.Check()
		result.Duration = time.Since(checkStart)
		result.Name = checker.Name()
		result.Category = checker.Category()
		report.Checks = append(report.Checks, result)
	}

	report.Duration = time.Since(start)
	report.CalculateHealth()

	return report
}

// PrintHuman выводит отчёт в человекочитаемом формате
func (r *Report) PrintHuman() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    gotr Self-Test Report                     ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Version:  %-50s║\n", r.Version)
	fmt.Printf("║  Commit:   %-50s║\n", r.Commit)
	fmt.Printf("║  Go:       %-50s║\n", r.GoVersion)
	fmt.Printf("║  Platform: %-50s║\n", r.Platform)
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	currentCategory := ""
	for _, check := range r.Checks {
		if check.Category != currentCategory {
			currentCategory = check.Category
			fmt.Printf("\n[%s]\n", currentCategory)
		}
		fmt.Printf("  %s %s %s\n",
			check.Result.Color()+check.Result.String()+check.Result.ResetColor(),
			check.Name,
			formatDetails(check),
		)
		if check.Details != "" {
			fmt.Printf("      %s\n", check.Details)
		}
	}

	fmt.Println()
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Results:  %s %d passed  %s %d failed  %s %d warn  %s %d skip   ║\n",
		ResultPass.String(), r.TotalPassed,
		ResultFail.String(), r.TotalFailed,
		ResultWarn.String(), r.TotalWarn,
		ResultSkip.String(), r.TotalSkip,
	)
	fmt.Printf("║  Overall:  %s%-56s%s║\n",
		r.Health.Color(), r.OverallStatus(), r.Health.ResetColor())
	fmt.Printf("║  Duration: %-50s║\n", r.Duration.Round(time.Millisecond))
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func formatDetails(c CheckResult) string {
	if c.Error != nil {
		return "— " + c.Error.Error()
	}
	if c.Message != "" {
		return "— " + c.Message
	}
	return ""
}
