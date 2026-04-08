// internal/migration/fetch.go
package migration

import (
	"context"

	"github.com/Korrnals/gotr/internal/models/data"
)

// FetchSharedStepsData retrieves shared steps from both source and target projects.
func (m *Migration) FetchSharedStepsData(ctx context.Context) (source, target data.GetSharedStepsResponse, err error) {
	m.logger.Info("Starting to fetch shared steps from source project")

	source, err = m.Client.GetSharedSteps(ctx, m.srcProject)
	if err != nil {
		m.logger.Errorw("Error fetching source shared steps", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Fetched shared steps from source", "count", len(source))

	m.logger.Info("Starting to fetch shared steps from target project")
	target, err = m.Client.GetSharedSteps(ctx, m.dstProject)
	if err != nil {
		m.logger.Errorw("Error fetching target shared steps", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Fetched shared steps from target", "count", len(target))

	return source, target, nil
}

// FetchCasesData retrieves cases from both source and target suites.
func (m *Migration) FetchCasesData(ctx context.Context) (source, target data.GetCasesResponse, err error) {
	m.logger.Info("Starting to fetch cases from source suite")

	source, err = m.Client.GetCases(ctx, m.srcProject, m.srcSuite, 0)
	if err != nil {
		m.logger.Errorw("Error fetching source cases", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Fetched cases from source", "count", len(source))

	m.logger.Info("Starting to fetch cases from target suite")
	target, err = m.Client.GetCases(ctx, m.dstProject, m.dstSuite, 0)
	if err != nil {
		m.logger.Errorw("Error fetching target cases", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Fetched cases from target", "count", len(target))

	return source, target, nil
}

// FetchSuitesData retrieves suites from both source and target projects.
func (m *Migration) FetchSuitesData(ctx context.Context) (source, target data.GetSuitesResponse, err error) {
	m.logger.Info("Starting to fetch suites from source project")
	source, err = m.Client.GetSuites(ctx, m.srcProject)
	if err != nil {
		m.logger.Errorw("Error fetching source suites", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Fetched suites from source", "count", len(source))

	m.logger.Info("Starting to fetch suites from target project")
	target, err = m.Client.GetSuites(ctx, m.dstProject)
	if err != nil {
		m.logger.Errorw("Error fetching target suites", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Fetched suites from target", "count", len(target))

	return source, target, nil
}

// FetchSectionsData retrieves sections from both source and target suites.
func (m *Migration) FetchSectionsData(ctx context.Context) (source, target data.GetSectionsResponse, err error) {
	m.logger.Info("Starting to fetch sections from source suite")

	source, err = m.Client.GetSections(ctx, m.srcProject, m.srcSuite)
	if err != nil {
		m.logger.Errorw("Error fetching source sections", "suite_id", m.srcSuite, "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Fetched sections from source", "count", len(source))

	m.logger.Info("Starting to fetch sections from target suite")
	target, err = m.Client.GetSections(ctx, m.dstProject, m.dstSuite)
	if err != nil {
		m.logger.Errorw("Error fetching target sections", "suite_id", m.dstSuite, "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Fetched sections from target", "count", len(target))

	return source, target, nil
}
