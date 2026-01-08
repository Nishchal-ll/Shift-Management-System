package models

import (
	"log"
	"time"
)

type ShiftType struct {
	Name      string
	Hours     string
	Quota     int
	Rate      float64
	ID        string
	StartTime string
	EndTime   string
}

func GetShiftTypes() []ShiftType {
	rows, err := DB.Query("SELECT id, name, start_time, end_time, quota FROM shift_types ORDER BY id ASC")
	if err != nil {
		log.Println("Error fetching shifts:", err)
		return []ShiftType{}
	}
	defer rows.Close()

	var shifts []ShiftType
	for rows.Next() {
		var s ShiftType
		if err := rows.Scan(&s.ID, &s.Name, &s.StartTime, &s.EndTime, &s.Quota); err != nil {
			continue
		}
		shifts = append(shifts, s)
	}
	return shifts
}

func AddShift(name, start, end string, quota int) error {
	layout := "15:04"
	t1, err := time.Parse(layout, start)
	if err != nil {
		return err
	}
	t2, err := time.Parse(layout, end)
	if err != nil {
		return err
	}

	if t2.Before(t1) {
		t2 = t2.Add(24 * time.Hour)
	}

	durationHours := t2.Sub(t1).Hours()

	query := `INSERT INTO shift_types 
              (name, start_time, end_time, quota, hours) 
              VALUES ($1, $2, $3, $4, $5)`

	_, err = DB.Exec(query, name, start, end, quota, durationHours)
	return err
}

func UpdateShiftQuota(id string, newQuota int) error {
	_, err := DB.Exec("UPDATE shift_types SET quota = $1 WHERE id = $2", newQuota, id)
	return err
}

func DeleteShiftType(id string) error {
	_, err := DB.Exec("DELETE FROM shift_types WHERE id = $1", id)
	return err
}
