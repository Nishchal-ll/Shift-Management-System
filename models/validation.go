package models

import (
	"fmt"
)

// --- VALIDATION & CHECKS ---

func CheckQuota(shiftName, startDate, endDate string) error {
	var limit int
	err := DB.QueryRow("SELECT quota FROM shift_types WHERE TRIM(name) = TRIM($1)", shiftName).Scan(&limit)
	if err != nil {
		return fmt.Errorf("shift type not found: %s", shiftName)
	}

	var currentCount int
	query := `
        SELECT COUNT(*) FROM allocations 
        WHERE shift_name = $1 
        AND start_date <= $3 AND end_date >= $2`

	err = DB.QueryRow(query, shiftName, startDate, endDate).Scan(&currentCount)
	if err != nil {
		return err
	}

	if currentCount >= limit {
		return fmt.Errorf("Quota Exceeded! %s shift limit is %d.", shiftName, limit)
	}
	return nil
}

func CheckQuotaAvailability(shiftName, startStr, endStr string) error {
	var limit int
	err := DB.QueryRow("SELECT quota FROM shift_types WHERE name = $1", shiftName).Scan(&limit)
	if err != nil {
		return fmt.Errorf("Shift type not found")
	}

	var count int
	query := `
        SELECT COUNT(DISTINCT employee_name) 
        FROM allocations 
        WHERE shift_name = $1 
        AND (start_date <= $3 AND end_date >= $2)`

	err = DB.QueryRow(query, shiftName, startStr, endStr).Scan(&count)
	if err != nil {
		return err
	}

	if count >= limit {
		return fmt.Errorf("Quota Exceeded! Limit is %d, currently assigned: %d", limit, count)
	}

	return nil
}

func CheckAvailability(employee, start, end string, excludeID int) error {
	query := `
        SELECT count(*) 
        FROM allocations 
        WHERE employee_name = $1 
        AND (start_date < $3 AND end_date > $2) 
        AND id != $4
    `
	var count int
	err := DB.QueryRow(query, employee, start, end, excludeID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("Time overlap detected")
	}

	return nil
}
