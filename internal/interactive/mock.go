package interactive

import "fmt"

// SelectResponse is a queued response for Select calls in MockPrompter.
type SelectResponse struct {
	Index int
	Value string
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
func (m *MockPrompter) Select(message string, options []string) (idx int, value string, err error) {
	if len(options) == 0 {
		return 0, "", fmt.Errorf("select options list is empty")
	}

	if m.selectPos >= len(m.selects) {
		return 0, "", fmt.Errorf("mock select queue exhausted")
	}

	response := m.selects[m.selectPos]
	m.selectPos++

	if response.Index < 0 || response.Index >= len(options) {
		return 0, "", fmt.Errorf("mock select index out of range: %d", response.Index)
	}

	value = response.Value
	if value == "" {
		value = options[response.Index]
	}

	return response.Index, value, nil
}

// MultilineInput behaves the same as Input for test purposes.
func (m *MockPrompter) MultilineInput(message, defaultVal string) (string, error) {
	return m.Input(message, defaultVal)
}
