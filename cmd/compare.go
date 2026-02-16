package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/utils"
	"github.com/spf13/cobra"
)

// compareCmd — команда для сравнения данных между проектами
var compareCmd = &cobra.Command{
	Use:   "compare <resource> [args...]",
	Short: "Сравнение данных между проектами",
	Long: `Выполнение сравнения ресурсов между двумя проектами по указанному полю.

Параметры:
	--pid1 <id>   ID первого проекта
	--pid2 <id>   ID второго проекта
	--field <name> Поле для сравнения (по умолчанию 'title')

Примеры:
	gotr compare cases --pid1 30 --pid2 31 --field title
	gotr compare cases --pid1 30 --pid2 31 --field priority_id
	gotr compare all --pid1 30 --pid2 31
`,
	Args: cobra.MinimumNArgs(1), // resource обязателен
	RunE: func(cmd *cobra.Command, args []string) error {
		cli := GetClientInterface(cmd)

		resource := args[0]

		pid1Str, _ := cmd.Flags().GetString("pid1")
		pid1, err := strconv.ParseInt(pid1Str, 10, 64)
		if err != nil || pid1 <= 0 {
			return fmt.Errorf("укажите корректный pid1 (--pid1)")
		}

		pid2Str, _ := cmd.Flags().GetString("pid2")
		pid2, err := strconv.ParseInt(pid2Str, 10, 64)
		if err != nil || pid2 <= 0 {
			return fmt.Errorf("укажите корректный pid2 (--pid2)")
		}

		field, _ := cmd.Flags().GetString("field")
		if field == "" {
			field = "title" // default
		}

		switch resource {
		case "all":
			return compareAllResources(cli, pid1, pid2, field)
		case "cases":
			return compareCases(cli, pid1, pid2, field)
		case "suites":
			return compareNamedResource(cli, pid1, pid2, "suites", fetchSuiteNames)
		case "sections":
			return compareNamedResource(cli, pid1, pid2, "sections", fetchSectionNames)
		case "sharedsteps":
			return compareNamedResource(cli, pid1, pid2, "sharedsteps", fetchSharedStepNames)
		case "runs":
			return compareNamedResource(cli, pid1, pid2, "runs", fetchRunNames)
		case "plans":
			return compareNamedResource(cli, pid1, pid2, "plans", fetchPlanNames)
		case "milestones":
			return compareNamedResource(cli, pid1, pid2, "milestones", fetchMilestoneNames)
		case "datasets":
			return compareNamedResource(cli, pid1, pid2, "datasets", fetchDatasetNames)
		case "groups":
			return compareNamedResource(cli, pid1, pid2, "groups", fetchGroupNames)
		case "labels":
			return compareNamedResource(cli, pid1, pid2, "labels", fetchLabelNames)
		case "templates":
			return compareNamedResource(cli, pid1, pid2, "templates", fetchTemplateNames)
		case "configurations":
			return compareConfigurations(cli, pid1, pid2)
		default:
			return fmt.Errorf("неизвестный ресурс: %s", resource)
		}
	},
}

type resourceDiff struct {
	resource    string
	onlyFirst   []string
	onlySecond  []string
	common      []string
	totalFirst  int
	totalSecond int
}

type summaryItem struct {
	resource        string
	diff            resourceDiff
	diffByField     int
	err             error
	diffByFieldErr  error
}

type resourceFetcher func(cli client.ClientInterface, projectID int64) ([]string, error)

func compareCases(cli client.ClientInterface, pid1 int64, pid2 int64, field string) error {
	diff, err := cli.DiffCasesData(pid1, pid2, field)
	if err != nil {
		return fmt.Errorf("ошибка сравнения кейсов: %w", err)
	}

	// Вывод
	fmt.Printf("Сравнение проектов %d и %d по полю '%s':\n\n", pid1, pid2, field)

	fmt.Printf("Только в проекте %d:\n", pid1)
	if len(diff.OnlyInFirst) == 0 {
		fmt.Println("  - Нет уникальных кейсов")
	} else {
		for _, c := range diff.OnlyInFirst {
			fmt.Printf("  - %d: %s\n", c.ID, c.Title)
		}
	}

	fmt.Printf("\nТолько в проекте %d:\n", pid2)
	if len(diff.OnlyInSecond) == 0 {
		fmt.Println("  - Нет уникальных кейсов")
	} else {
		for _, c := range diff.OnlyInSecond {
			fmt.Printf("  - %d: %s\n", c.ID, c.Title)
		}
	}

	fmt.Printf("\nОтличаются по полю '%s':\n", field)
	if len(diff.DiffByField) == 0 {
		fmt.Println("  - Нет отличий")
	} else {
		for _, d := range diff.DiffByField {
			firstValue := strings.TrimSpace(utils.GetFieldValue(d.First, field))
			secondValue := strings.TrimSpace(utils.GetFieldValue(d.Second, field))
			if firstValue == "" {
				firstValue = "<пусто>"
			}
			if secondValue == "" {
				secondValue = "<пусто>"
			}
			fmt.Printf("  - %s:\n", d.First.Title)
			fmt.Printf("    Проект %d (case %d): %s\n", pid1, d.First.ID, firstValue)
			fmt.Printf("    Проект %d (case %d): %s\n", pid2, d.Second.ID, secondValue)
		}
	}

	return nil
}

func compareNamedResource(cli client.ClientInterface, pid1, pid2 int64, resource string, fetcher resourceFetcher) error {
	first, err := fetcher(cli, pid1)
	if err != nil {
		return fmt.Errorf("ошибка получения %s для проекта %d: %w", resource, pid1, err)
	}

	second, err := fetcher(cli, pid2)
	if err != nil {
		return fmt.Errorf("ошибка получения %s для проекта %d: %w", resource, pid2, err)
	}

	diff := buildResourceDiff(resource, first, second)
	printResourceDiff(diff, pid1, pid2)
	return nil
}

func compareConfigurations(cli client.ClientInterface, pid1, pid2 int64) error {
	groupDiff, err := compareNamedResourceInline(cli, pid1, pid2, "config-groups", fetchConfigGroupNames)
	if err != nil {
		return err
	}
	printResourceDiff(groupDiff, pid1, pid2)

	configDiff, err := compareNamedResourceInline(cli, pid1, pid2, "configs", fetchConfigNames)
	if err != nil {
		return err
	}
	printResourceDiff(configDiff, pid1, pid2)
	return nil
}

func compareNamedResourceInline(cli client.ClientInterface, pid1, pid2 int64, resource string, fetcher resourceFetcher) (resourceDiff, error) {
	first, err := fetcher(cli, pid1)
	if err != nil {
		return resourceDiff{}, fmt.Errorf("ошибка получения %s для проекта %d: %w", resource, pid1, err)
	}

	second, err := fetcher(cli, pid2)
	if err != nil {
		return resourceDiff{}, fmt.Errorf("ошибка получения %s для проекта %d: %w", resource, pid2, err)
	}

	return buildResourceDiff(resource, first, second), nil
}

func compareAllResources(cli client.ClientInterface, pid1, pid2 int64, field string) error {
	resources := []struct {
		name    string
		fetcher resourceFetcher
	}{
		{"suites", fetchSuiteNames},
		{"sections", fetchSectionNames},
		{"sharedsteps", fetchSharedStepNames},
		{"runs", fetchRunNames},
		{"plans", fetchPlanNames},
		{"milestones", fetchMilestoneNames},
		{"datasets", fetchDatasetNames},
		{"groups", fetchGroupNames},
		{"labels", fetchLabelNames},
		{"templates", fetchTemplateNames},
		{"config-groups", fetchConfigGroupNames},
		{"configs", fetchConfigNames},
	}

	summaries := make([]summaryItem, 0, len(resources)+1)

	caseDiffByField := 0
	caseDiff, caseDiffErr := cli.DiffCasesData(pid1, pid2, field)
	if caseDiffErr == nil {
		caseDiffByField = len(caseDiff.DiffByField)
	}

	caseDiffSummary, caseErr := compareNamedResourceInline(cli, pid1, pid2, "cases", fetchCaseNames)
	if caseErr != nil {
		summaries = append(summaries, summaryItem{resource: "cases", err: caseErr})
	} else {
		summaries = append(summaries, summaryItem{resource: "cases", diff: caseDiffSummary, diffByField: caseDiffByField, diffByFieldErr: caseDiffErr})
	}

	for _, res := range resources {
		diff, err := compareNamedResourceInline(cli, pid1, pid2, res.name, res.fetcher)
		summaries = append(summaries, summaryItem{resource: res.name, diff: diff, err: err})
	}

	printSummary(summaries, pid1, pid2, field)
	return nil
}

func buildResourceDiff(resource string, first, second []string) resourceDiff {
	firstSet := toNameSet(first)
	secondSet := toNameSet(second)

	commonKeys := make([]string, 0)
	onlyFirstKeys := make([]string, 0)
	onlySecondKeys := make([]string, 0)

	for key := range firstSet {
		if _, ok := secondSet[key]; ok {
			commonKeys = append(commonKeys, key)
		} else {
			onlyFirstKeys = append(onlyFirstKeys, key)
		}
	}

	for key := range secondSet {
		if _, ok := firstSet[key]; !ok {
			onlySecondKeys = append(onlySecondKeys, key)
		}
	}

	return resourceDiff{
		resource:    resource,
		onlyFirst:   namesFromKeys(onlyFirstKeys, firstSet),
		onlySecond:  namesFromKeys(onlySecondKeys, secondSet),
		common:      namesFromKeys(commonKeys, firstSet),
		totalFirst:  len(firstSet),
		totalSecond: len(secondSet),
	}
}

func toNameSet(values []string) map[string]string {
	set := make(map[string]string, len(values))
	for _, value := range values {
		normalized := normalizeName(value)
		if normalized == "" {
			continue
		}
		if _, ok := set[normalized]; !ok {
			set[normalized] = strings.TrimSpace(value)
		}
	}
	return set
}

func normalizeName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func namesFromKeys(keys []string, lookup map[string]string) []string {
	if len(keys) == 0 {
		return nil
	}
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		values = append(values, lookup[key])
	}
	sort.Strings(values)
	return values
}

func printResourceDiff(diff resourceDiff, pid1, pid2 int64) {
	fmt.Printf("\nРесурс '%s':\n", diff.resource)
	fmt.Printf("  Всего в проекте %d: %d\n", pid1, diff.totalFirst)
	fmt.Printf("  Всего в проекте %d: %d\n", pid2, diff.totalSecond)
	fmt.Printf("  Общие: %d\n", len(diff.common))

	fmt.Printf("\n  Только в проекте %d:\n", pid1)
	if len(diff.onlyFirst) == 0 {
		fmt.Println("    - Нет")
	} else {
		for _, name := range diff.onlyFirst {
			fmt.Printf("    - %s\n", name)
		}
	}

	fmt.Printf("\n  Только в проекте %d:\n", pid2)
	if len(diff.onlySecond) == 0 {
		fmt.Println("    - Нет")
	} else {
		for _, name := range diff.onlySecond {
			fmt.Printf("    - %s\n", name)
		}
	}
}

func printSummary(summaries []summaryItem, pid1, pid2 int64, field string) {
	fmt.Printf("Сводный отчет по проектам %d и %d:\n\n", pid1, pid2)

	for _, item := range summaries {
		if item.err != nil {
			fmt.Printf("- %s: ошибка: %v\n", item.resource, item.err)
			continue
		}
		line := fmt.Sprintf("- %s: всего %d/%d, общие %d, только %d/%d", item.resource, item.diff.totalFirst, item.diff.totalSecond, len(item.diff.common), len(item.diff.onlyFirst), len(item.diff.onlySecond))
		if item.resource == "cases" {
			if item.diffByFieldErr != nil {
				line = fmt.Sprintf("%s, отличий по полю '%s': ошибка (%v)", line, field, item.diffByFieldErr)
			} else {
				line = fmt.Sprintf("%s, отличий по полю '%s': %d", line, field, item.diffByField)
			}
		}
		fmt.Println(line)
	}

	for _, item := range summaries {
		if item.err != nil {
			continue
		}
		if len(item.diff.onlyFirst) == 0 && len(item.diff.onlySecond) == 0 {
			continue
		}
		fmt.Printf("\nРасхождения по ресурсу '%s':\n", item.resource)
		fmt.Printf("  Только в проекте %d:\n", pid1)
		if len(item.diff.onlyFirst) == 0 {
			fmt.Println("    - Нет")
		} else {
			for _, name := range item.diff.onlyFirst {
				fmt.Printf("    - %s\n", name)
			}
		}

		fmt.Printf("\n  Только в проекте %d:\n", pid2)
		if len(item.diff.onlySecond) == 0 {
			fmt.Println("    - Нет")
		} else {
			for _, name := range item.diff.onlySecond {
				fmt.Printf("    - %s\n", name)
			}
		}
	}
}

func fetchCaseNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	cases, err := cli.GetCases(projectID, 0, 0)
	if err != nil {
		return nil, err
	}
	return collectNames(len(cases), func(i int) string { return cases[i].Title }), nil
}

func fetchSuiteNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	suites, err := cli.GetSuites(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(suites), func(i int) string { return suites[i].Name }), nil
}

func fetchSectionNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	suites, err := cli.GetSuites(projectID)
	if err != nil {
		return nil, err
	}

	suiteNames := make(map[int64]string, len(suites))
	for _, suite := range suites {
		suiteNames[suite.ID] = suite.Name
	}

	sections := make([]string, 0)
	if len(suites) == 0 {
		list, err := cli.GetSections(projectID, 0)
		if err != nil {
			return nil, err
		}
		for _, section := range list {
			sections = append(sections, formatSectionName(section.SuiteID, section.Name, suiteNames))
		}
		return sections, nil
	}

	for _, suite := range suites {
		list, err := cli.GetSections(projectID, suite.ID)
		if err != nil {
			return nil, err
		}
		for _, section := range list {
			sections = append(sections, formatSectionName(section.SuiteID, section.Name, suiteNames))
		}
	}

	return sections, nil
}

func fetchSharedStepNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	steps, err := cli.GetSharedSteps(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(steps), func(i int) string { return steps[i].Title }), nil
}

func fetchRunNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	runs, err := cli.GetRuns(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(runs), func(i int) string { return runs[i].Name }), nil
}

func fetchPlanNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	plans, err := cli.GetPlans(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(plans), func(i int) string { return plans[i].Name }), nil
}

func fetchMilestoneNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	milestones, err := cli.GetMilestones(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(milestones), func(i int) string { return milestones[i].Name }), nil
}

func fetchDatasetNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	datasets, err := cli.GetDatasets(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(datasets), func(i int) string { return datasets[i].Name }), nil
}

func fetchGroupNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	groups, err := cli.GetGroups(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(groups), func(i int) string { return groups[i].Name }), nil
}

func fetchLabelNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	labels, err := cli.GetLabels(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(labels), func(i int) string { return labels[i].Name }), nil
}

func fetchTemplateNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	templates, err := cli.GetTemplates(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(templates), func(i int) string { return templates[i].Name }), nil
}

func fetchConfigGroupNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	groups, err := cli.GetConfigs(projectID)
	if err != nil {
		return nil, err
	}
	return collectNames(len(groups), func(i int) string { return groups[i].Name }), nil
}

func fetchConfigNames(cli client.ClientInterface, projectID int64) ([]string, error) {
	groups, err := cli.GetConfigs(projectID)
	if err != nil {
		return nil, err
	}
	configs := make([]string, 0)
	for _, group := range groups {
		for _, config := range group.Configs {
			configs = append(configs, fmt.Sprintf("%s / %s", group.Name, config.Name))
		}
	}
	return configs, nil
}

func collectNames(size int, getter func(int) string) []string {
	if size == 0 {
		return nil
	}
	names := make([]string, 0, size)
	for i := 0; i < size; i++ {
		name := strings.TrimSpace(getter(i))
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func formatSectionName(suiteID int64, sectionName string, suiteNames map[int64]string) string {
	suiteName := strings.TrimSpace(suiteNames[suiteID])
	if suiteName == "" {
		if suiteID == 0 {
			suiteName = "suite:default"
		} else {
			suiteName = fmt.Sprintf("suite:%d", suiteID)
		}
	}
	sectionName = strings.TrimSpace(sectionName)
	if sectionName == "" {
		return suiteName
	}
	return fmt.Sprintf("%s / %s", suiteName, sectionName)
}
