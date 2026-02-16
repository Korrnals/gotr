// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"fmt"
	"strconv"
	"text/tabwriter"

	"github.com/Korrnals/gotr/cmd/internal/output"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/spf13/cobra"
)

// newListCmd создаёт команду 'attachments list'
// Поддерживает списки для case, plan, run, test, plan-entry
func newListCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Получить список вложений",
		Long: `Получает список вложений для различных типов ресурсов.

Поддерживаемые типы ресурсов:
  • case       — вложения тест-кейса
  • plan       — вложения тест-плана
  • plan-entry — вложения записи плана
  • run        — вложения тест-рана
  • test       — вложения теста`,
		Example: `  # Список вложений кейса
  gotr attachments list case 123

  # Список вложений плана
  gotr attachments list plan 456

  # Список вложений теста
  gotr attachments list test 789`,
	}

	// Добавляем подкоманды для каждого типа
	cmd.AddCommand(newListCaseCmd(getClient))
	cmd.AddCommand(newListPlanCmd(getClient))
	cmd.AddCommand(newListPlanEntryCmd(getClient))
	cmd.AddCommand(newListRunCmd(getClient))
	cmd.AddCommand(newListTestCmd(getClient))

	return cmd
}

func newListCaseCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "case <case_id>",
		Short: "Список вложений тест-кейса",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			caseID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || caseID <= 0 {
				return fmt.Errorf("invalid case_id: %s", args[0])
			}

			client := getClient(cmd)
			attachments, err := client.GetAttachmentsForCase(caseID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	cmd.Flags().StringP("output", "o", "", "Формат вывода: json")
	return cmd
}

func newListPlanCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan <plan_id>",
		Short: "Список вложений тест-плана",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			client := getClient(cmd)
			attachments, err := client.GetAttachmentsForPlan(planID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	cmd.Flags().StringP("output", "o", "", "Формат вывода: json")
	return cmd
}

func newListPlanEntryCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan-entry <plan_id> <entry_id>",
		Short: "Список вложений записи плана",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || planID <= 0 {
				return fmt.Errorf("invalid plan_id: %s", args[0])
			}

			client := getClient(cmd)
			attachments, err := client.GetAttachmentsForPlanEntry(planID, args[1])
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	cmd.Flags().StringP("output", "o", "", "Формат вывода: json")
	return cmd
}

func newListRunCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <run_id>",
		Short: "Список вложений тест-рана",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || runID <= 0 {
				return fmt.Errorf("invalid run_id: %s", args[0])
			}

			client := getClient(cmd)
			attachments, err := client.GetAttachmentsForRun(runID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	cmd.Flags().StringP("output", "o", "", "Формат вывода: json")
	return cmd
}

func newListTestCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <test_id>",
		Short: "Список вложений теста",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			testID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || testID <= 0 {
				return fmt.Errorf("invalid test_id: %s", args[0])
			}

			client := getClient(cmd)
			attachments, err := client.GetAttachmentsForTest(testID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	cmd.Flags().StringP("output", "o", "", "Формат вывода: json")
	return cmd
}

func outputAttachmentsList(cmd *cobra.Command, attachments data.GetAttachmentsResponse) error {
	outputFlag, _ := cmd.Flags().GetString("output")
	if outputFlag == "json" {
		return output.JSON(cmd, attachments)
	}

	if len(attachments) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No attachments found")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSIZE\tCREATED_ON")
	for _, a := range attachments {
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\n", a.ID, a.Name, a.Size, a.CreatedOn)
	}
	return w.Flush()
}
