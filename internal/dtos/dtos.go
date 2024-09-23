package dtos

type CodeSubmissionRequest struct {
	SourceCode     string `json:"source_code"`
	LanguageID     int    `json:"language_id"`
	Stdin          string `json:"stdin,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
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

type CreateGameRoomDto struct {
	RoomName string `json:"room_name"`
}

type CreateGameRoomResponseDto struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
	RoomID  string `json:"room_id,omitempty"`
}

type JoinGameRoomDto struct {
	Id string `json:"id"`
}

type JoinGameRoomResponseDto struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}
