package dtos

type CodeSubmissionRequest struct {
	SourceCode string `json:"source_code"`
	LanguageID int    `json:"language_id"`
}

type CodeSubmissionResponse struct {
	Token string `json:"token"`
}

type CodeSubmissionResult struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Status struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
	} `json:"status"`
	Time   string `json:"time"`
	Memory string `json:"memory"`
}
