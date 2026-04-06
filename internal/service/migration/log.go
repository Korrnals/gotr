// internal/migration/log.go
package migration

// LogInfo is a convenience method for structured info-level logging.
func (m *Migration) LogInfo(action string, fields ...interface{}) {
	m.logger.Infow(action, fields...)
}

// LogWarn is a convenience method for structured warning-level logging.
func (m *Migration) LogWarn(action string, fields ...interface{}) {
	m.logger.Warnw(action, fields...)
}

// LogError is a convenience method for structured error-level logging.
func (m *Migration) LogError(action string, err error, fields ...interface{}) {
	m.logger.Errorw(action, append(fields, "error", err)...)
}
