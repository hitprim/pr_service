package models

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

const (
	ErrTeamExists  = "TEAM_EXISTS"
	ErrPRExists    = "PR_EXISTS"
	ErrPRMerged    = "PR_MERGED"
	ErrNotAssigned = "NOT_ASSIGNED"
	ErrNoCandidate = "NO_CANDIDATE"
	ErrNotFound    = "NOT_FOUND"
)
