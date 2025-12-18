package data

type GetCasesResponse struct {
    Cases []TestCase `json:"cases"`
    Offset int       `json:"offset"`
    Limit  int       `json:"limit"`
    Size   int       `json:"size"`
    // _links можно игнорировать, если не нужен
}

type TestCase struct {
    ID         int    `json:"id"`
    Title      string `json:"title"`
    SectionID  int    `json:"section_id"`
    TypeID     int    `json:"type_id"`
    PriorityID int    `json:"priority_id"`
    Estimate   string `json:"estimate,omitempty"`   // может отсутствовать
    Custom     struct {
        Preconds   string `json:"preconds,omitempty"`
        Steps      string `json:"steps,omitempty"`
        Expected   string `json:"expected,omitempty"`
    } `json:"custom,omitempty"`
}
