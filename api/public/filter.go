// Package public
package public

const (
	RoleValidator = "validator"
	RoleCandidate = "candidate"
)

type ValidatorFilter struct {
	Role  string
	Skip  int64
	Limit int64
}
