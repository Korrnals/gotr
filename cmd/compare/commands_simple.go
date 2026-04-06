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
	"Compare groups between projects",
	`Compares user groups between two projects.

Examples:
  # Compare groups
  gotr compare groups --pid1 30 --pid2 31

  # Save result to the default file
  gotr compare groups --pid1 30 --pid2 31 --save

  # Save result to a specific file
  gotr compare groups --pid1 30 --pid2 31 --save-to groups_diff.json
`,
	fetchGroupItems,
)

var labelsCmd = newSimpleCompareCmd(
	"labels", "labels",
	"Compare labels between projects",
	`Compares labels between two projects.

Examples:
  # Compare labels
  gotr compare labels --pid1 30 --pid2 31

  # Save result to the default file
  gotr compare labels --pid1 30 --pid2 31 --save

  # Save result to a specific file
  gotr compare labels --pid1 30 --pid2 31 --save-to labels_diff.json
`,
	fetchLabelItems,
)

var milestonesCmd = newSimpleCompareCmd(
	"milestones", "milestones",
	"Compare milestones between projects",
	`Compares milestones between two projects.

Examples:
  # Compare milestones
  gotr compare milestones --pid1 30 --pid2 31

  # Save result to the default file
  gotr compare milestones --pid1 30 --pid2 31 --save

  # Save result to a specific file
  gotr compare milestones --pid1 30 --pid2 31 --save-to milestones_diff.json
`,
	fetchMilestoneItems,
)

var plansCmd = newSimpleCompareCmd(
	"plans", "plans",
	"Compare test plans between projects",
	`Compares test plans between two projects.

Examples:
  # Compare test plans
  gotr compare plans --pid1 30 --pid2 31

  # Save result to the default file
  gotr compare plans --pid1 30 --pid2 31 --save

  # Save result to a specific file
  gotr compare plans --pid1 30 --pid2 31 --save-to plans_diff.json
`,
	fetchPlanItems,
)

var runsCmd = newSimpleCompareCmd(
	"runs", "runs",
	"Compare test runs between projects",
	`Compares test runs between two projects.

Examples:
  # Compare test runs
  gotr compare runs --pid1 30 --pid2 31

  # Save result to the default file
  gotr compare runs --pid1 30 --pid2 31 --save

  # Save result to a specific file
  gotr compare runs --pid1 30 --pid2 31 --save-to runs_diff.json
`,
	fetchRunItems,
)

var sharedStepsCmd = newSimpleCompareCmd(
	"sharedsteps", "sharedsteps",
	"Compare shared steps between projects",
	`Compares shared steps between two projects.

Examples:
  # Compare shared steps
  gotr compare sharedsteps --pid1 30 --pid2 31

  # Save result to the default file
  gotr compare sharedsteps --pid1 30 --pid2 31 --save

  # Save result to a specific file
  gotr compare sharedsteps --pid1 30 --pid2 31 --save-to sharedsteps_diff.json
`,
	fetchSharedStepItems,
)

var templatesCmd = newSimpleCompareCmd(
	"templates", "templates",
	"Compare templates between projects",
	`Compares case templates between two projects.

Examples:
  # Compare templates
  gotr compare templates --pid1 30 --pid2 31

  # Save result to the default file
  gotr compare templates --pid1 30 --pid2 31 --save

  # Save result to a specific file
  gotr compare templates --pid1 30 --pid2 31 --save-to templates_diff.json
`,
	fetchTemplateItems,
)

var configurationsCmd = newSimpleCompareCmd(
	"configurations", "configurations",
	"Compare configurations between projects",
	`Compares configurations (config groups and configs) between two projects.

Examples:
  # Compare configurations
  gotr compare configurations --pid1 30 --pid2 31

  # Save result to the default file
  gotr compare configurations --pid1 30 --pid2 31 --save

  # Save result to a specific file
  gotr compare configurations --pid1 30 --pid2 31 --save-to configs_diff.json
`,
	fetchConfigurationItems,
)

var datasetsCmd = newSimpleCompareCmd(
	"datasets", "datasets",
	"Compare datasets between projects",
	`Compares datasets between two projects.

Examples:
  # Compare datasets
  gotr compare datasets --pid1 30 --pid2 31

  # Save result to the default file
  gotr compare datasets --pid1 30 --pid2 31 --save

  # Save result to a specific file
  gotr compare datasets --pid1 30 --pid2 31 --save-to datasets_diff.json
`,
	fetchDatasetItems,
)
