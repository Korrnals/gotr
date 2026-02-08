// cmd/common/client.go
// Общие функции для работы с HTTP клиентом в CLI командах
package common

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) *client.HTTPClient

// ClientAccessor предоставляет доступ к HTTP клиенту для команд
type ClientAccessor struct {
	getClient GetClientFunc
}

// NewClientAccessor создаёт новый accessor с заданной функцией получения клиента
func NewClientAccessor(fn GetClientFunc) *ClientAccessor {
	return &ClientAccessor{getClient: fn}
}

// GetClientSafe безопасно получает клиент с проверкой на nil
func (ca *ClientAccessor) GetClientSafe(cmd *cobra.Command) *client.HTTPClient {
	if ca.getClient == nil {
		return nil
	}
	return ca.getClient(cmd)
}

// SetClientForTests устанавливает функцию получения клиента для тестов
func (ca *ClientAccessor) SetClientForTests(fn GetClientFunc) {
	ca.getClient = fn
}

// GetClientSafe безопасно вызывает getClient с проверкой на nil (глобальная функция)
func GetClientSafe(cmd *cobra.Command, getClient GetClientFunc) *client.HTTPClient {
	if getClient == nil {
		return nil
	}
	return getClient(cmd)
}
