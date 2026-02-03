// internal/migration/log.go
package migration

// LogInfo — удобный метод для info-лога с полями
func (m *Migration) LogInfo(action string, fields ...interface{}) {
	m.logger.Infow(action, fields...)
}

// LogWarn — warning
func (m *Migration) LogWarn(action string, fields ...interface{}) {
	m.logger.Warnw(action, fields...)
}

// LogError — error
func (m *Migration) LogError(action string, err error, fields ...interface{}) {
	m.logger.Errorw(action, append(fields, "error", err)...)
}
