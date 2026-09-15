package models

import (
	"log"
	"time"
)

type Notification struct {
	ID           int       `json:"id"`
	EmployeeName string    `json:"employee_name"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	ShiftName    string    `json:"shift_name"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
	TimeAgo      string    `json:"time_ago"`
}

func CreateNotification(employee, notifType, title, message, shiftName string) error {
	query := `INSERT INTO notifications (employee_name, type, title, message, shift_name) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err := DB.Exec(query, employee, notifType, title, message, shiftName)
	if err != nil {
		log.Println("[DB] Error creating notification:", err)
	}
	return err
}

func GetNotifications(employee string) []Notification {
	query := `SELECT id, employee_name, type, title, message, COALESCE(shift_name, ''), is_read, created_at 
              FROM notifications 
              WHERE employee_name = $1 OR employee_name = 'all' 
              ORDER BY created_at DESC LIMIT 20`
	rows, err := DB.Query(query, employee)
	if err != nil {
		log.Println("[DB] Error querying notifications:", err)
		return []Notification{}
	}
	defer rows.Close()

	var notifs []Notification
	now := time.Now()
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.EmployeeName, &n.Type, &n.Title, &n.Message, &n.ShiftName, &n.IsRead, &n.CreatedAt); err != nil {
			continue
		}

		diff := now.Sub(n.CreatedAt)
		if diff < time.Minute {
			n.TimeAgo = "Just now"
		} else if diff < time.Hour {
			n.TimeAgo = diff.Truncate(time.Minute).String() + " ago"
		} else if diff < 24*time.Hour {
			n.TimeAgo = diff.Truncate(time.Hour).String() + " ago"
		} else {
			n.TimeAgo = n.CreatedAt.Format("Jan 02 15:04")
		}

		notifs = append(notifs, n)
	}
	return notifs
}

func ClearNotifications(employee string) error {
	_, err := DB.Exec("DELETE FROM notifications WHERE employee_name = $1", employee)
	return err
}
