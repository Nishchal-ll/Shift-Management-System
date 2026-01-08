package models

// GetUserRole checks credentials and returns the role (admin/user)
func GetUserRole(username, password string) (string, bool) {
	var role string
	err := DB.QueryRow("SELECT role FROM users WHERE username=$1 AND password=$2", username, password).Scan(&role)
	if err != nil {
		return "", false
	}
	return role, true
}

func CreateUser(username string) error {
	_, err := DB.Exec("INSERT INTO users (username, password, role) VALUES ($1, $1, 'user') ON CONFLICT (username) DO NOTHING", username)
	return err
}

func GetAllEmployees() []string {
	rows, err := DB.Query("SELECT username FROM users WHERE role = 'user'")
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var u string
		rows.Scan(&u)
		users = append(users, u)
	}
	return users
}

func DeleteUserComplete(username string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM allocations WHERE employee_name = $1", username)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM users WHERE username = $1", username)
	if err != nil {
		return err
	}

	return tx.Commit()
}
