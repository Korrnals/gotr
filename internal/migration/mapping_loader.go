package migration

import (
	"fmt"
	"github.com/Korrnals/gotr/internal/utils"
)

// LoadMappingFromFile загружает mapping из файла и заполняет внутреннюю структуру m.mapping.
// Поддерживает два формата: полный SharedStepMapping (LoadSharedStepMapping) и простой map JSON (utils.LoadMapping).
func (m *Migration) LoadMappingFromFile(file string) error {
	// Попробуем сначала полный формат mapping (pairs)
	sm, err := LoadSharedStepMapping(file)
	if err == nil && sm != nil && len(sm.Pairs) > 0 {
		m.mapping = sm
		return nil
	}

	// Иначе попробуем простой формат
	simple, err := utils.LoadMapping(file)
	if err == nil && len(simple) > 0 {
		for s, t := range simple {
			m.mapping.AddPair(s, t, "existing")
		}
		return nil
	}

	return fmt.Errorf("не удалось загрузить mapping из файла %s", file)
}
