package get

import (
	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
)

// selectProjectInteractively показывает список проектов и просит выбрать
// DEPRECATED: используйте interactive.SelectProjectInteractively
func selectProjectInteractively(client *client.HTTPClient) (int64, error) {
	return interactive.SelectProjectInteractively(client)
}

// selectSuiteInteractively показывает список сьютов и просит выбрать
// DEPRECATED: используйте interactive.SelectSuiteInteractively
func selectSuiteInteractively(suites data.GetSuitesResponse) (int64, error) {
	return interactive.SelectSuiteInteractively(suites)
}
