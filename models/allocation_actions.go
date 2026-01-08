package models

import (
	"time"
)

// --- COMPLEX GROUPING LOGIC ---

func GetGroupedAllocations() []GroupedAllocation {
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
        ORDER BY employee_name ASC, start_date ASC
    `
	rows, err := DB.Query(query)
	if err != nil {
		return []GroupedAllocation{}
	}
	defer rows.Close()

	groupedMap := make(map[string]*GroupedAllocation)
	var order []string

	for rows.Next() {
		var id int
		var emp, shift, start, end, status, newReq string
		rows.Scan(&id, &emp, &shift, &start, &end, &status, &newReq)

		if len(start) > 10 {
			start = start[:10]
		}
		if len(end) > 10 {
			end = end[:10]
		}

		key := emp + "|" + start + "|" + end

		if _, exists := groupedMap[key]; !exists {
			t_start, _ := time.Parse("2006-01-02", start)
			t_end, _ := time.Parse("2006-01-02", end)
			timeline := t_start.Format("Jan 02") + " - " + t_end.Format("Jan 02")

			groupedMap[key] = &GroupedAllocation{
				EmployeeName: emp,
				StartDate:    start,
				EndDate:      end,
				Timeline:     timeline,
				Details:      []ShiftDetail{},
			}
			order = append(order, key)
		}

		groupedMap[key].Details = append(groupedMap[key].Details, ShiftDetail{
			ID:                id,
			ShiftName:         shift,
			Status:            status,
			NewRequestedShift: newReq,
		})
	}

	var result []GroupedAllocation
	for _, k := range order {
		result = append(result, *groupedMap[k])
	}
	return result
}

// --- REQUESTS & SWAPS ---

func RequestSwap(id int, newShift string) error {
	query := `UPDATE allocations SET new_requested_shift = $1, status = 'Pending' WHERE id = $2`
	_, err := DB.Exec(query, newShift, id)
	return err
}

func ApproveSwap(id int) error {
	query := `
        UPDATE allocations 
        SET shift_name = new_requested_shift, 
            status = 'Confirmed', 
            new_requested_shift = NULL 
        WHERE id = $1
    `
	_, err := DB.Exec(query, id)
	return err
}

func AddDaysToShift(id, days int) {
	DB.Exec("UPDATE allocations SET end_date = end_date + make_interval(days => $1) WHERE id=$2", days, id)
}

func ReplaceEmployee(targetShiftID int, newEmployee string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentOwner, startDate, endDate string
	err = tx.QueryRow(`SELECT employee_name, start_date, end_date FROM allocations WHERE id = $1`, targetShiftID).
		Scan(&currentOwner, &startDate, &endDate)
	if err != nil {
		return err
	}

	var conflictingShiftID int
	err = tx.QueryRow(`SELECT id FROM allocations WHERE employee_name = $1 AND start_date <= $3 AND end_date >= $2`,
		newEmployee, startDate, endDate).Scan(&conflictingShiftID)

	updateQuery := `
        UPDATE allocations 
        SET employee_name = $1, 
            status = 'Confirmed', 
            new_requested_shift = NULL 
        WHERE id = $2
    `

	if err == nil {
		// SWAP Logic
		_, err = tx.Exec(updateQuery, newEmployee, targetShiftID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(updateQuery, currentOwner, conflictingShiftID)
		if err != nil {
			return err
		}
	} else {
		// REPLACE Logic
		_, err = tx.Exec(updateQuery, newEmployee, targetShiftID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
