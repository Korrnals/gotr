package compare

// commands_simple.go — declares the 9 simple compare subcommands using the
// generic newSimpleCompareCmd factory. Each command fetches a resource from
// two projects in parallel and produces a diff.
//
// Replaces the formerly duplicated files: groups.go, labels.go, milestones.go,
// plans.go, runs.go, sharedsteps.go, templates.go, configurations.go,
// datasets.go (~1080 LOC → ~90 LOC).

var groupsCmd = newSimpleCompareCmd(
	"groups", "groups",
	"Сравнить группы между проектами",
	`Выполняет сравнение групп пользователей между двумя проектами.

Примеры:
  # Сравнить группы
  gotr compare groups --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare groups --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare groups --pid1 30 --pid2 31 --save-to groups_diff.json
`,
	fetchGroupItems,
)

var labelsCmd = newSimpleCompareCmd(
	"labels", "labels",
	"Сравнить метки между проектами",
	`Выполняет сравнение меток (labels) между двумя проектами.

Примеры:
  # Сравнить метки
  gotr compare labels --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare labels --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare labels --pid1 30 --pid2 31 --save-to labels_diff.json
`,
	fetchLabelItems,
)

var milestonesCmd = newSimpleCompareCmd(
	"milestones", "milestones",
	"Сравнить milestones между проектами",
	`Выполняет сравнение milestones между двумя проектами.

Примеры:
  # Сравнить milestones
  gotr compare milestones --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare milestones --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare milestones --pid1 30 --pid2 31 --save-to milestones_diff.json
`,
	fetchMilestoneItems,
)

var plansCmd = newSimpleCompareCmd(
	"plans", "plans",
	"Сравнить test plans между проектами",
	`Выполняет сравнение test plans между двумя проектами.

Примеры:
  # Сравнить test plans
  gotr compare plans --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare plans --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare plans --pid1 30 --pid2 31 --save-to plans_diff.json
`,
	fetchPlanItems,
)

var runsCmd = newSimpleCompareCmd(
	"runs", "runs",
	"Сравнить test runs между проектами",
	`Выполняет сравнение test runs между двумя проектами.

Примеры:
  # Сравнить test runs
  gotr compare runs --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare runs --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare runs --pid1 30 --pid2 31 --save-to runs_diff.json
`,
	fetchRunItems,
)

var sharedStepsCmd = newSimpleCompareCmd(
	"sharedsteps", "sharedsteps",
	"Сравнить shared steps между проектами",
	`Выполняет сравнение shared steps (общих шагов) между двумя проектами.

Примеры:
  # Сравнить shared steps
  gotr compare sharedsteps --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare sharedsteps --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare sharedsteps --pid1 30 --pid2 31 --save-to sharedsteps_diff.json
`,
	fetchSharedStepItems,
)

var templatesCmd = newSimpleCompareCmd(
	"templates", "templates",
	"Сравнить шаблоны между проектами",
	`Выполняет сравнение шаблонов кейсов между двумя проектами.

Примеры:
  # Сравнить шаблоны
  gotr compare templates --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare templates --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare templates --pid1 30 --pid2 31 --save-to templates_diff.json
`,
	fetchTemplateItems,
)

var configurationsCmd = newSimpleCompareCmd(
	"configurations", "configurations",
	"Сравнить конфигурации между проектами",
	`Выполняет сравнение конфигураций (config groups и configs) между двумя проектами.

Примеры:
  # Сравнить конфигурации
  gotr compare configurations --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare configurations --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare configurations --pid1 30 --pid2 31 --save-to configs_diff.json
`,
	fetchConfigurationItems,
)

var datasetsCmd = newSimpleCompareCmd(
	"datasets", "datasets",
	"Сравнить datasets между проектами",
	`Выполняет сравнение datasets между двумя проектами.

Примеры:
  # Сравнить datasets
  gotr compare datasets --pid1 30 --pid2 31

  # Сохранить результат в файл по умолчанию
  gotr compare datasets --pid1 30 --pid2 31 --save

  # Сохранить результат в указанный файл
  gotr compare datasets --pid1 30 --pid2 31 --save-to datasets_diff.json
`,
	fetchDatasetItems,
)
