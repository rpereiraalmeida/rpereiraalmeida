// Add this to your database initialization function

// Create audit_logs table if it doesn't exist
_, err = db.db.Exec(`
CREATE TABLE IF NOT EXISTS audit_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    action TEXT NOT NULL,
    details TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
)
`)
if err != nil {
    return nil, fmt.Errorf("error creating audit_logs table: %v", err)
}

// Add role column to users table if it doesn't exist
_, err = db.db.Exec(`
PRAGMA table_info(users)
`)
if err != nil {
    return nil, fmt.Errorf("error checking users table schema: %v", err)
}

// Check if role column exists in users table
var hasRoleColumn bool
rows, err := db.db.Query("PRAGMA table_info(users)")
if err != nil {
    return nil, fmt.Errorf("error checking users table schema: %v", err)
}
defer rows.Close()

for rows.Next() {
    var cid, notnull, pk int
    var name, type_ string
    var dflt_value interface{}
    
    if err := rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk); err != nil {
        return nil, fmt.Errorf("error scanning table info: %v", err)
    }
    
    if name == "role" {
        hasRoleColumn = true
        break
    }
}

// Add role column if it doesn't exist
if !hasRoleColumn {
    _, err = db.db.Exec(`
    ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'
    `)
    if err != nil {
        return nil, fmt.Errorf("error adding role column to users table: %v", err)
    }
    
    // Set first user as admin
    _, err = db.db.Exec(`
    UPDATE users SET role = 'admin' WHERE id IN (SELECT id FROM users LIMIT 1)
    `)
    if err != nil {
        return nil, fmt.Errorf("error setting first user as admin: %v", err)
    }
}
