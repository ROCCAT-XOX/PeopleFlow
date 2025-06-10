package model

// ContractType definiert die Art des Arbeitsvertrags
type ContractType string

const (
	ContractTypeFullTime  ContractType = "full_time"
	ContractTypePartTime  ContractType = "part_time"
	ContractTypeMiniJob   ContractType = "mini_job"
	ContractTypeIntern    ContractType = "intern"
	ContractTypeFreelance ContractType = "freelance"
)

// IsValid prüft, ob der ContractType gültig ist
func (ct ContractType) IsValid() bool {
	switch ct {
	case ContractTypeFullTime, ContractTypePartTime, ContractTypeMiniJob, ContractTypeIntern, ContractTypeFreelance:
		return true
	default:
		return false
	}
}
