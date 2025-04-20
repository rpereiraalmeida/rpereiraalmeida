// Add these functions to your existing database.go file

// User roles
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// UpdateUserRole updates a user's role
func (db *Database) UpdateUserRole(userID string, role string) error {
	_, err := db.db.Exec("UPDATE users SET role = ? WHERE id = ?", role, userID)
	return err
}

// GetAllUsers retrieves all users with pagination
func (db *Database) GetAllUsers(page, perPage int) ([]User, error) {
	offset := (page - 1) * perPage
	rows, err := db.db.Query("SELECT id, username, email, created_at, role FROM users LIMIT ? OFFSET ?", 
		perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.Role); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetTotalUserCount returns the total number of users
func (db *Database) GetTotalUserCount() (int, error) {
	var count int
	err := db.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// GetUserStats retrieves user statistics
func (db *Database) GetUserStats() (map[string]int, error) {
	stats := make(map[string]int)
	
	// Total users
	totalUsers, err := db.GetTotalUserCount()
	if err != nil {
		return nil, err
	}
	stats["totalUsers"] = totalUsers
	
	// Users registered in the last 7 days
	err = db.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE created_at > datetime('now', '-7 days')").Scan(&stats["newUsers"])
	if err != nil {
		return nil, err
	}
	
	// Users with OAuth connections
	err = db.db.QueryRow(
		"SELECT COUNT(DISTINCT user_id) FROM oauth_connections").Scan(&stats["oauthUsers"])
	if err != nil {
		return nil, err
	}
	
	return stats, nil
}

// GetTripStats retrieves trip statistics
func (db *Database) GetTripStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total trips
	var totalTrips int
	err := db.db.QueryRow("SELECT COUNT(*) FROM trips").Scan(&totalTrips)
	if err != nil {
		return nil, err
	}
	stats["totalTrips"] = totalTrips
	
	// Trips in the last 30 days
	var recentTrips int
	err = db.db.QueryRow(
		"SELECT COUNT(*) FROM trips WHERE date > datetime('now', '-30 days')").Scan(&recentTrips)
	if err != nil {
		return nil, err
	}
	stats["recentTrips"] = recentTrips
	
	// Average trip cost
	var avgCost float64
	err = db.db.QueryRow("SELECT AVG(cost) FROM trips").Scan(&avgCost)
	if err != nil {
		return nil, err
	}
	stats["avgCost"] = avgCost
	
	// Monthly trip counts for the past 6 months
	rows, err := db.db.Query(`
		SELECT strftime('%m-%Y', date) as month, COUNT(*) as count 
		FROM trips 
		WHERE date > datetime('now', '-6 months') 
		GROUP BY month 
		ORDER BY date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	monthlyData := make(map[string]int)
	for rows.Next() {
		var month string
		var count int
		if err := rows.Scan(&month, &count); err != nil {
			return nil, err
		}
		monthlyData[month] = count
	}
	stats["monthlyTrips"] = monthlyData
	
	return stats, nil
}

// DeleteUser removes a user and all associated data
func (db *Database) DeleteUser(userID string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	
	// Delete all trips
	_, err = tx.Exec("DELETE FROM trips WHERE user_id = ?", userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// Delete OAuth connections
	_, err = tx.Exec("DELETE FROM oauth_connections WHERE user_id = ?", userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// Delete password reset tokens
	_, err = tx.Exec("DELETE FROM password_reset_tokens WHERE user_id = ?", userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	// Delete user
	_, err = tx.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	return tx.Commit()
}

// GetAuditLogs retrieves system audit logs with pagination
func (db *Database) GetAuditLogs(page, perPage int) ([]AuditLog, error) {
	offset := (page - 1) * perPage
	rows, err := db.db.Query(`
		SELECT id, user_id, action, details, ip_address, created_at 
		FROM audit_logs 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.Details, 
			&log.IPAddress, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	
	return logs, nil
}

// AddAuditLog adds a new audit log entry
func (db *Database) AddAuditLog(userID, action, details, ipAddress string) error {
	_, err := db.db.Exec(`
		INSERT INTO audit_logs (user_id, action, details, ip_address, created_at)
		VALUES (?, ?, ?, ?, datetime('now'))
	`, userID, action, details, ipAddress)
	return err
}

// AuditLog represents a system audit log entry
type AuditLog struct {
	ID         string
	UserID     string
	Action     string
	Details    string
	IPAddress  string
	CreatedAt  string
}
