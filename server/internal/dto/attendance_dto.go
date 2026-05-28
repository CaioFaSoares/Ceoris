package dto

import "chantry/server/internal/pocketbase"

// AttendanceListResponse defines the response payload structure for attendance records.
type AttendanceListResponse struct {
	AttendanceID    string `json:"attendance_id"`
	StudentName     string `json:"student_name"`
	StudentNickname string `json:"student_nickname"`
	Date            string `json:"date"`
	ClockIn         string `json:"clock_in"`
	ClockOut        string `json:"clock_out"`
	Status          string `json:"status"`
	Source          string `json:"source"`
}

// ToAttendanceListResponse maps a raw PocketBase AttendanceRecord into a flattened DTO.
func ToAttendanceListResponse(model pocketbase.AttendanceRecord) AttendanceListResponse {
	studentName := model.Expand.Student.Username
	studentNickname := model.Expand.Student.Nickname
	if studentNickname == "" {
		studentNickname = studentName
	}

	return AttendanceListResponse{
		AttendanceID:    model.ID,
		StudentName:     studentName,
		StudentNickname: studentNickname,
		Date:            model.Date,
		ClockIn:         model.ClockIn,
		ClockOut:        model.ClockOut,
		Status:          model.Status,
		Source:          model.Source,
	}
}
