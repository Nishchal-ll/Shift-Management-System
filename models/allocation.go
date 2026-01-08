package models

import (
	"fmt"
	"time"
)

// --- DATA STRUCTURES ---

type Allocation struct {
	ID                int
	EmployeeName      string
	ShiftName         string
	StartDate         time.Time
	EndDate           time.Time
	Status            string
	NewRequestedShift string
}

type ShiftDetail struct {
	ID                int
	ShiftName         string
	Status            string
	NewRequestedShift string
	StartDate         time.Time
	EndDate           time.Time
}

type GroupedAllocation struct {
	EmployeeName string
	StartDate    string
	EndDate      string
	Timeline     string
	Details      []ShiftDetail
}

// --- BASIC CRUD ---

func CreateAllocation(emp, shift string, start, end time.Time) error {
	_, err := DB.Exec("INSERT INTO allocations (employee_name, shift_name, start_date, end_date) VALUES ($1, $2, $3, $4)", emp, shift, start, end)
	return err
}

func GetAllocations() []Allocation {
	query := `
        SELECT 
            id, 
            employee_name, 
            shift_name, 
            start_date, 
            end_date, 
            status, 
            COALESCE(new_requested_shift, '') 
        FROM allocations 
        ORDER BY start_date DESC
    `
	rows, err := DB.Query(query)
	if err != nil {
		fmt.Println("Error fetching allocations:", err)
		return []Allocation{}
	}
	defer rows.Close()

	var allocs []Allocation
	for rows.Next() {
		var a Allocation
		err := rows.Scan(&a.ID, &a.EmployeeName, &a.ShiftName, &a.StartDate, &a.EndDate, &a.Status, &a.NewRequestedShift)
		if err != nil {
			continue
		}
		allocs = append(allocs, a)
	}
	return allocs
}

func GetUserAllocations(username string) []Allocation {
	rows, err := DB.Query(`SELECT id, employee_name, shift_name, start_date, end_date, status, new_requested_shift 
                           FROM allocations 
                           WHERE employee_name = $1 
                           ORDER BY start_date DESC`, username)
	if err != nil {
		return []Allocation{}
	}
	defer rows.Close()

	var allocs []Allocation
	for rows.Next() {
		var a Allocation
		rows.Scan(&a.ID, &a.EmployeeName, &a.ShiftName, &a.StartDate, &a.EndDate, &a.Status, &a.NewRequestedShift)
		allocs = append(allocs, a)
	}
	return allocs
}

func DeleteAllocation(id int) error {
	_, err := DB.Exec("DELETE FROM allocations WHERE id = $1", id)
	return err
}

func UpdateAllocation(id int, employee, shift string, start, end string) error {
	// Calls validation functions from validation.go
	if err := CheckQuota(shift, start, end); err != nil {
		return err
	}
	if err := CheckAvailability(employee, start, end, id); err != nil {
		return err
	}
	query := `UPDATE allocations 
              SET employee_name=$1, shift_name=$2, start_date=$3, end_date=$4 
              WHERE id=$5`
	_, err := DB.Exec(query, employee, shift, start, end, id)
	return err
}
