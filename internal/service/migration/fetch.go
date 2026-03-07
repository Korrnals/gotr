// internal/migration/fetch.go
package migration

import (
	"context"
	"github.com/Korrnals/gotr/internal/models/data"
)

// FetchSharedStepsData — получает shared steps из source и target проектов
func (m *Migration) FetchSharedStepsData(ctx context.Context) (source data.GetSharedStepsResponse, target data.GetSharedStepsResponse, err error) {
	m.logger.Info("Начало получения shared steps из source проекта")

	source, err = m.Client.GetSharedSteps(ctx, m.srcProject)
	if err != nil {
		m.logger.Errorw("Ошибка получения source shared steps", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Получено shared steps из source", "count", len(source))

	m.logger.Info("Начало получения shared steps из target проекта")
	target, err = m.Client.GetSharedSteps(ctx, m.dstProject)
	if err != nil {
		m.logger.Errorw("Ошибка получения target shared steps", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Получено shared steps из target", "count", len(target))

	return source, target, nil
}

// FetchCasesData — получает кейсы из source и target сюит
func (m *Migration) FetchCasesData(ctx context.Context) (source data.GetCasesResponse, target data.GetCasesResponse, err error) {
	m.logger.Info("Начало получения кейсов из source suite")

	source, err = m.Client.GetCases(ctx, m.srcProject, m.srcSuite, 0)
	if err != nil {
		m.logger.Errorw("Ошибка получения source cases", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Получено кейсов из source", "count", len(source))

	m.logger.Info("Начало получения кейсов из target suite")
	target, err = m.Client.GetCases(ctx, m.dstProject, m.dstSuite, 0)
	if err != nil {
		m.logger.Errorw("Ошибка получения target cases", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Получено кейсов из target", "count", len(target))

	return source, target, nil
}

// FetchSuitesData — получает suites из source и target проектов
func (m *Migration) FetchSuitesData(ctx context.Context) (source data.GetSuitesResponse, target data.GetSuitesResponse, err error) {
	m.logger.Info("Начало получения suites из source проекта")
	source, err = m.Client.GetSuites(ctx, m.srcProject)
	if err != nil {
		m.logger.Errorw("Ошибка получения source suites", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Получено suites из source", "count", len(source))

	m.logger.Info("Начало получения suites из target проекта")
	target, err = m.Client.GetSuites(ctx, m.dstProject)
	if err != nil {
		m.logger.Errorw("Ошибка получения target suites", "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Получено suites из target", "count", len(target))

	return source, target, nil
}

// FetchSectionsData — получает sections из source и target suite
func (m *Migration) FetchSectionsData(ctx context.Context) (source data.GetSectionsResponse, target data.GetSectionsResponse, err error) {
	m.logger.Info("Начало получения sections из source suite")

	source, err = m.Client.GetSections(ctx, m.srcProject, m.srcSuite)
	if err != nil {
		m.logger.Errorw("Ошибка получения source sections", "suite_id", m.srcSuite, "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Получено sections из source", "count", len(source))

	m.logger.Info("Начало получения sections из target suite")
	target, err = m.Client.GetSections(ctx, m.dstProject, m.dstSuite)
	if err != nil {
		m.logger.Errorw("Ошибка получения target sections", "suite_id", m.dstSuite, "error", err)
		return nil, nil, err
	}
	m.logger.Infow("Получено sections из target", "count", len(target))

	return source, target, nil
}
