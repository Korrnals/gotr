// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package attachments

import (
	"fmt"

	"github.com/Korrnals/gotr/internal/flags"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/Korrnals/gotr/internal/output"
	"github.com/Korrnals/gotr/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
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
			caseID, err := flags.ValidateRequiredID(args, 0, "case_id")
			if err != nil {
				return err
			}

			client := getClient(cmd)
			ctx := cmd.Context()
			attachments, err := client.GetAttachmentsForCase(ctx, caseID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	output.AddFlag(cmd)
	return cmd
}

func newListPlanCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan <plan_id>",
		Short: "Список вложений тест-плана",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := flags.ValidateRequiredID(args, 0, "plan_id")
			if err != nil {
				return err
			}

			client := getClient(cmd)
			ctx := cmd.Context()
			attachments, err := client.GetAttachmentsForPlan(ctx, planID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	output.AddFlag(cmd)
	return cmd
}

func newListPlanEntryCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan-entry <plan_id> <entry_id>",
		Short: "Список вложений записи плана",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID, err := flags.ValidateRequiredID(args, 0, "plan_id")
			if err != nil {
				return err
			}

			client := getClient(cmd)
			ctx := cmd.Context()
			attachments, err := client.GetAttachmentsForPlanEntry(ctx, planID, args[1])
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	output.AddFlag(cmd)
	return cmd
}

func newListRunCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <run_id>",
		Short: "Список вложений тест-рана",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID, err := flags.ValidateRequiredID(args, 0, "run_id")
			if err != nil {
				return err
			}

			client := getClient(cmd)
			ctx := cmd.Context()
			attachments, err := client.GetAttachmentsForRun(ctx, runID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	output.AddFlag(cmd)
	return cmd
}

func newListTestCmd(getClient GetClientFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test <test_id>",
		Short: "Список вложений теста",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			testID, err := flags.ValidateRequiredID(args, 0, "test_id")
			if err != nil {
				return err
			}

			client := getClient(cmd)
			ctx := cmd.Context()
			attachments, err := client.GetAttachmentsForTest(ctx, testID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			return outputAttachmentsList(cmd, attachments)
		},
	}
	output.AddFlag(cmd)
	return cmd
}

func outputAttachmentsList(cmd *cobra.Command, attachments data.GetAttachmentsResponse) error {
	saveFlag, _ := cmd.Flags().GetBool("save")
	if saveFlag {
		_, err := output.Output(cmd, attachments, "attachments", "json")
		return err
	}

	if len(attachments) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No attachments found")
		return nil
	}

	t := ui.NewTable(cmd)
	t.AppendHeader(table.Row{"ID", "NAME", "SIZE", "CREATED_ON"})
	for _, a := range attachments {
		t.AppendRow(table.Row{a.ID, a.Name, a.Size, a.CreatedOn})
	}
	ui.Table(cmd, t)
	return nil
}
