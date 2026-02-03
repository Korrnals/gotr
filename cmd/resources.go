package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Korrnals/gotr/internal/utils"
	"github.com/Korrnals/gotr/pkg/testrailapi"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// Глобальная инициализация TestRailAPI структур (Инициализируем один раз)
var api = testrailapi.New()

// Определяем свой собственный тип ключа (unexported — только внутри пакета cmd)
type contextKey string

// Константа — наш ключ.
const httpClientKey contextKey = "httpClient"

// ValidResources — динамически генерируемый список всех ресурсов
var ValidResources []string

func init() {
	// Собираем уникальные ресурсы из всех групп
	seen := make(map[string]bool)
	resources := []string{"all"} // "all" — специальный ресурс

	groups := []struct {
		name  string
		paths []testrailapi.APIPath
	}{
		{"cases", api.Cases.Paths()},
		{"casefields", api.CaseFields.Paths()},
		{"casetypes", api.CaseTypes.Paths()},
		{"configurations", api.Configurations.Paths()},
		{"projects", api.Projects.Paths()},
		{"priorities", api.Priorities.Paths()},
		{"runs", api.Runs.Paths()},
		{"tests", api.Tests.Paths()},
		{"suites", api.Suites.Paths()},
		{"sections", api.Sections.Paths()},
		{"statuses", api.Statuses.Paths()},
		{"milestones", api.Milestones.Paths()},
		{"plans", api.Plans.Paths()},
		{"results", api.Results.Paths()},
		{"resultfields", api.ResultFields.Paths()},
		{"reports", api.Reports.Paths()},
		{"attachments", api.Attachments.Paths()},
		{"users", api.Users.Paths()},
		{"roles", api.Roles.Paths()},
		{"templates", api.Templates.Paths()},
		{"groups", api.Groups.Paths()},
		{"sharedsteps", api.SharedSteps.Paths()},
		{"variables", api.Variables.Paths()},
		{"labels", api.Labels.Paths()},
		{"datasets", api.Datasets.Paths()},
		{"bdds", api.BDDs.Paths()},
	}

	for _, g := range groups {
		if len(g.paths) > 0 {
			seen[g.name] = true
		}
	}

	for r := range seen {
		resources = append(resources, r)
	}
	sort.Strings(resources)
	ValidResources = resources
}

// extractGetEndpointName — надёжно извлекает имя после "/get_"
func extractGetEndpointName(uri string) string {
	// Находим позицию "/get_"
	idx := strings.LastIndex(uri, "/get_")
	if idx == -1 {
		return "" // не стандартный GET-эндпоинт TestRail
	}

	name := uri[idx+1:] // всё после "/get_"

	// Отрезаем query-параметры начиная с "&"
	if qIdx := strings.Index(name, "&"); qIdx != -1 {
		name = name[:qIdx]
	}

	// Отрезаем плейсхолдеры начиная с "{"
	if phIdx := strings.Index(name, "{"); phIdx != -1 {
		name = name[:phIdx]
	}

	// Чистим trailing слеши и пробелы
	name = strings.Trim(name, "/ ")

	if name == "" {
		return ""
	}

	return name
}

// getValidGetEndpoints — все чистые имена для автодополнения
func getValidGetEndpoints() []string {
	var names []string
	seen := make(map[string]bool)

	for _, p := range api.Paths() {
		if p.Method != "GET" {
			continue
		}
		name := extractGetEndpointName(p.URI)
		if name != "" && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}

	// Сортируем для красоты
	sort.Strings(names)
	return names
}

// extractGetEndpointName — извлекает чистый эндпоинт с плейсхолдерами (без query и trailing слешей)
func extractAllEndpointName(uri string) string {
	// Убираем префикс "index.php?/api/v2/"
	uri = strings.TrimPrefix(uri, "index.php?/api/v2/")

	// Отрезаем query-параметры начиная с "&"
	if qIdx := strings.Index(uri, "&"); qIdx != -1 {
		uri = uri[:qIdx]
	}

	// Чистим trailing слеши и пробелы
	uri = strings.Trim(uri, "/ ")

	if uri == "" {
		return ""
	}

	return uri
}

// Вспомогательня функция 'getResourceEndpoints' - получения списка 'endpoints' соответствующего ресурса
func getResourceEndpoints(resource string, outputType string) ([]string, error) {
	var paths []testrailapi.APIPath
	switch resource {
	case "all":
		paths = api.Paths()
	case "cases":
		paths = api.Cases.Paths()
	case "casefields":
		paths = api.CaseFields.Paths()
	case "casetypes":
		paths = api.CaseTypes.Paths()
	case "configurations":
		paths = api.Configurations.Paths()
	case "projects":
		paths = api.Projects.Paths()
	case "priorities":
		paths = api.Priorities.Paths()
	case "runs":
		paths = api.Runs.Paths()
	case "tests":
		paths = api.Tests.Paths()
	case "suites":
		paths = api.Suites.Paths()
	case "sections":
		paths = api.Sections.Paths()
	case "statuses":
		paths = api.Statuses.Paths()
	case "milestones":
		paths = api.Milestones.Paths()
	case "plans":
		paths = api.Plans.Paths()
	case "results":
		paths = api.Results.Paths()
	case "resultfields":
		paths = api.ResultFields.Paths()
	case "reports":
		paths = api.Reports.Paths()
	case "attachments":
		paths = api.Attachments.Paths()
	case "users":
		paths = api.Users.Paths()
	case "roles":
		paths = api.Roles.Paths()
	case "templates":
		paths = api.Templates.Paths()
	case "groups":
		paths = api.Groups.Paths()
	case "sharedsteps":
		paths = api.SharedSteps.Paths()
	case "variables":
		paths = api.Variables.Paths()
	case "labels":
		paths = api.Labels.Paths()
	case "datasets":
		paths = api.Datasets.Paths()
	case "bdds":
		paths = api.BDDs.Paths()
	default:
		fmt.Printf("Неизвестный ресурс: %s\n\nДоступные ресурсы:\n", resource)
		fmt.Println("  all, cases, casefields, casetypes, configurations, projects, priorities,")
		fmt.Println("  runs, tests, suites, sections, statuses, milestones, plans, results,")
		fmt.Println("  resultfields, reports, attachments, users, roles, templates, groups,")
		fmt.Println("  sharedsteps, variables, labels, datasets, bdds")
		return nil, nil
	}

	// Сортируем для красоты
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].URI < paths[j].URI
	})

	var endpoints []string
	switch outputType {
	// Вывод в JSON — красиво и удобно для скриптов
	case "json":
		data, err := json.MarshalIndent(paths, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("ошибка формирования JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil, err
	// Вывод 'Method + Endpoints'
	case "short":
		for _, p := range paths {
			fmt.Printf("%s %s\n", p.Method, p.URI)
		}
		return nil, fmt.Errorf("ошибка формирования короткого списка ресурсов")
	// Краткий вывод — только URI
	case "list":
		for _, p := range paths {
			name := extractGetEndpointName(p.URI)
			endpoints = append(endpoints, name)
		}
		return endpoints, fmt.Errorf("ошибка формирования списка ресурсов")
	default:
		fmt.Printf("Эндпоинты для %s (%d):\n\n", resource, len(paths))
		for _, p := range paths {
			fmt.Printf("  %s %s\n      %s\n", p.Method, p.URI, p.Description)
			if len(p.Params) > 0 {
				fmt.Print("      Параметры:\n")
				for name, desc := range p.Params {
					fmt.Printf("        - %s: %s\n", name, desc)
				}
			}
			fmt.Println()
		}
	}

	return endpoints, nil
}

// getAllShortEndpoints — возвращает список всех коротких эндпоинтов для ресурса (GET, POST, DELETE)
func getAllShortEndpoints(resource string) []string {
	var paths []testrailapi.APIPath
	paths = getResourcePaths(resource)
	if paths == nil {
		return nil
	}

	var endpoints []string
	seen := make(map[string]bool)
	for _, p := range paths {
		name := extractAllEndpointName(p.URI)
		if name != "" && !seen[name] {
			seen[name] = true
			endpoints = append(endpoints, name)
		}
	}

	sort.Strings(endpoints)
	return endpoints
}

// Внешняя функция-обертка, которая возвращает ВСЕ ендпоинты конкретного ресурса
var endpointsCache = make(map[string][]string)

// GetEndpoints — возвращает все короткие эндпоинты для ресурса (с кэшированием)
func GetEndpoints(resource string) []string {
	if cached, ok := endpointsCache[resource]; ok {
		return cached
	}

	var endpoints []string
	if resource == "all" {
		seen := make(map[string]bool)
		for _, r := range ValidResources {
			if r == "all" {
				continue
			}
			resEndpoints := GetEndpoints(r) // рекурсия, но кэш спасает
			for _, e := range resEndpoints {
				if !seen[e] {
					seen[e] = true
					endpoints = append(endpoints, e)
				}
			}
		}
		sort.Strings(endpoints)
	} else {
		endpoints = getAllShortEndpoints(resource)
		if endpoints == nil {
			// Можно залогировать, но не возвращать ошибку — просто пустой срез
			return nil
		}
	}

	endpointsCache[resource] = endpoints
	return endpoints
}

// Вспомогательная приватная функция 'replaceAllPlaceholders' - которая выполняет замену плейсхолдеров на соответствующие `_id“
func replaceAllPlaceholders(uri, id string) string {
	placeholders := []string{
		"{project_id}", "{case_id}", "{run_id}", "{test_id}", "{section_id}",
		"{suite_id}", "{milestone_id}", "{plan_id}", "{user_id}", "{role_id}",
		"{group_id}", "{dataset_id}", "{shared_step_id}", "{report_template_id}",
		"{email}",
	}
	for _, ph := range placeholders {
		uri = strings.ReplaceAll(uri, ph, id)
	}
	return uri
}

// buildRequestParams — собирает полный эндпоинт и query-параметры из флагов и позиционного ID
// Приватная функция (маленькая буква) — используется только внутри пакета cmd
func buildRequestParams(endpoint string, mainID string, cmd *cobra.Command) (string, map[string]string, error) {
	fullEndpoint := endpoint
	queryParams := make(map[string]string)

	// Подставляем основной ID (project_id, run_id и т.д.)
	if mainID != "" {
		fullEndpoint = replaceAllPlaceholders(fullEndpoint, mainID)
		if !strings.Contains(fullEndpoint, mainID) {
			fullEndpoint += "/" + mainID
		}
		utils.DebugPrint("{resources} - fullEndpoint после ID: %s", fullEndpoint)
	}

	// Query-параметры — только если значение не пустое
	flags := []struct {
		flagName string // имя флага в Cobra
		queryKey string // имя в TestRail API
	}{
		{"suite-id", "suite_id"},
		{"section-id", "section_id"},
		{"milestone-id", "milestone_id"},
		{"assignedto-id", "assignedto_id"},
		{"status-id", "status_id"},
		{"priority-id", "priority_id"},
		{"type-id", "type_id"},
		{"created-by", "created_by"},
		{"updated-by", "updated_by"},
		// Добавляй новые по мере надобности
	}

	for _, f := range flags {
		if val, err := cmd.Flags().GetString(f.flagName); err == nil && val != "" {
			queryParams[f.queryKey] = val
			utils.DebugPrint("{resources} - Добавлен параметр: %s = %s", f.queryKey, val)
		}
	}

	return fullEndpoint, queryParams, nil
}

// getResourcePaths — возвращает пути для указанного ресурса
// resource — имя ресурса ("projects", "cases" и т.д.)
func getResourcePaths(resource string) []testrailapi.APIPath {
	switch resource {
	case "all":
		return api.Paths()
	case "cases":
		return api.Cases.Paths()
	case "casefields":
		return api.CaseFields.Paths()
	case "casetypes":
		return api.CaseTypes.Paths()
	case "configurations":
		return api.Configurations.Paths()
	case "projects":
		return api.Projects.Paths()
	case "priorities":
		return api.Priorities.Paths()
	case "runs":
		return api.Runs.Paths()
	case "tests":
		return api.Tests.Paths()
	case "suites":
		return api.Suites.Paths()
	case "sections":
		return api.Sections.Paths()
	case "statuses":
		return api.Statuses.Paths()
	case "milestones":
		return api.Milestones.Paths()
	case "plans":
		return api.Plans.Paths()
	case "results":
		return api.Results.Paths()
	case "resultfields":
		return api.ResultFields.Paths()
	case "reports":
		return api.Reports.Paths()
	case "attachments":
		return api.Attachments.Paths()
	case "users":
		return api.Users.Paths()
	case "roles":
		return api.Roles.Paths()
	case "templates":
		return api.Templates.Paths()
	case "groups":
		return api.Groups.Paths()
	case "sharedsteps":
		return api.SharedSteps.Paths()
	case "variables":
		return api.Variables.Paths()
	case "labels":
		return api.Labels.Paths()
	case "datasets":
		return api.Datasets.Paths()
	case "bdds":
		return api.BDDs.Paths()
	default:
		return nil // или return []APIPath{} — пустой срез
	}
}
