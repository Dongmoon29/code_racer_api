package judge0

import "github.com/Dongmoon29/code_racer_api/internal/dtos"

func mapJudge0Request(dto dtos.CodeSubmissionRequest) map[string]interface{} {
	mappedRequest := map[string]interface{}{
		"source_code": dto.SourceCode,
		"language_id": dto.LanguageID,
		"stdin":       dto.Stdin,
	}
	return mappedRequest
}
