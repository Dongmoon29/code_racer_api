package judge0

import (
	"encoding/json"
	"fmt"

	"github.com/Dongmoon29/code_racer_api/internal/dtos"
	"github.com/Dongmoon29/code_racer_api/internal/util/client"
)

type Judge0Service struct {
}

func NewJudge0Service() Judge0Service {
	return Judge0Service{}
}

func (js *Judge0Service) CreateCodeSubmission(dto dtos.CodeSubmissionRequest) (map[string]interface{}, error) {
	submissionData := mapJudge0Request(dto)

	response, err := client.J0Client.POST("/submissions?base64_encoded=false&wait=true", submissionData)
	if err != nil {
		return nil, fmt.Errorf("failed to submit code: %w", err)
	}
	defer response.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
