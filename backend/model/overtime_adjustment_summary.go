package model

type OvertimeAdjustmentSummary struct {
	EmployeeID    string  `json:"employeeId"`
	TotalPending  int     `json:"totalPending"`
	TotalApproved int     `json:"totalApproved"`
	TotalRejected int     `json:"totalRejected"`
	HoursPending  float64 `json:"hoursPending"`
	HoursApproved float64 `json:"hoursApproved"`
	HoursRejected float64 `json:"hoursRejected"`
}
