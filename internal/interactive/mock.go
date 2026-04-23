package interactive

import "fmt"

// SelectResponse is a queued response for Select calls in MockPrompter.
type SelectResponse struct {
	Index int
	Value string
	// Raw disables auto-adjustment for Browse navigation options.
	// When false (default), the mock automatically skips "← Back" / "✕ Exit"
	// options prepended by Browse, so Index 0 always means the first real item.
	// Set to true when testing navigation option selection explicitly.
	Raw bool
}

// MockPrompter is deterministic prompter for unit tests.
type MockPrompter struct {
	inputs   []string
	confirms []bool
	selects  []SelectResponse

	inputPos   int
	confirmPos int
	selectPos  int
}

// NewMockPrompter creates a mock prompter.
func NewMockPrompter() *MockPrompter {
	return &MockPrompter{}
}

// WithInputResponses appends queued input responses.
func (m *MockPrompter) WithInputResponses(values ...string) *MockPrompter {
	m.inputs = append(m.inputs, values...)
	return m
}

// WithConfirmResponses appends queued confirm responses.
func (m *MockPrompter) WithConfirmResponses(values ...bool) *MockPrompter {
	m.confirms = append(m.confirms, values...)
	return m
}

// WithSelectResponses appends queued select responses.
func (m *MockPrompter) WithSelectResponses(values ...SelectResponse) *MockPrompter {
	m.selects = append(m.selects, values...)
	return m
}

// Input returns next queued input response.
func (m *MockPrompter) Input(message, defaultVal string) (string, error) {
	if m.inputPos >= len(m.inputs) {
		return "", fmt.Errorf("mock input queue exhausted")
	}

	answer := m.inputs[m.inputPos]
	m.inputPos++

	return answer, nil
}

// Confirm returns next queued confirm response.
func (m *MockPrompter) Confirm(message string, def bool) (bool, error) {
	if m.confirmPos >= len(m.confirms) {
		return false, fmt.Errorf("mock confirm queue exhausted")
	}

	answer := m.confirms[m.confirmPos]
	m.confirmPos++

	return answer, nil
}

// Select returns next queued select response.
// When the response has Raw=false (default), the index is auto-adjusted
// to skip Browse navigation options (← Back, ✕ Exit) at the top of the list.
func (m *MockPrompter) Select(message string, options []string) (idx int, value string, err error) {
	if len(options) == 0 {
		return 0, "", fmt.Errorf("select options list is empty")
	}

	if m.selectPos >= len(m.selects) {
		return 0, "", fmt.Errorf("mock select queue exhausted")
	}

	response := m.selects[m.selectPos]
	m.selectPos++

	actualIdx := response.Index
	if !response.Raw {
		actualIdx += countNavPrefix(options)
	}

	if actualIdx < 0 || actualIdx >= len(options) {
		return 0, "", fmt.Errorf("mock select index out of range: %d (raw=%d, nav=%d, opts=%d)",
			actualIdx, response.Index, countNavPrefix(options), len(options))
	}

	value = response.Value
	if value == "" {
		value = options[actualIdx]
	}

	return actualIdx, value, nil
}

// countNavPrefix counts Browse navigation options (← Back, ✕ Exit) at the
// start of the option list.
func countNavPrefix(options []string) int {
	n := 0
	for _, opt := range options {
		if opt == OptBack || opt == OptExit {
			n++
		} else {
			break
		}
	}
	return n
}

// MultilineInput behaves the same as Input for test purposes.
func (m *MockPrompter) MultilineInput(message, defaultVal string) (string, error) {
	return m.Input(message, defaultVal)
}
