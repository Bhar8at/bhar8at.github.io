package database

import (
	"fmt"
	"log"

	"github.com/Bhar8at/bhar8at.github.io/models"
	"github.com/lib/pq"
)

func CreateUser(user *models.User) bool {
	// inserting user data into t_users table
	if _, err := db.Exec(
		`INSERT INTO t_users(email, username, password, id, verified, avatar, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.Email,
		user.Username,
		user.Password,
		user.Id,
		user.Verified,
		user.Avatar,
		user.CreatedAt,
	); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func CreateOAuthUser(id string) bool {
	if _, err := db.Exec(`INSERT INTO o_users(id) VALUES ($1)`, id); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func ReadUserByName(username string) *models.User {
	var user models.User
	// stores data retrieved from database into the user struct
	if err := db.QueryRow(`SELECT * FROM t_users WHERE username = $1`, username).Scan(
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Id,
		&user.Verified,
		&user.Avatar,
		&user.CreatedAt,
	); err != nil {
		log.Println(err)
		return nil
	}
	return &user
}

func ReadUserByEmail(email string) *models.User {
	var user models.User
	if err := db.QueryRow(`SELECT * FROM t_users WHERE email = $1`, email).Scan(
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Id,
		&user.Verified,
		&user.Avatar,
		&user.CreatedAt,
	); err != nil {
		log.Println(err)
		return nil
	}
	return &user
}

func ReadUserById(id string) *models.User {
	var user models.User
	if err := db.QueryRow(`SELECT * FROM t_users WHERE id = $1`, id).Scan(
		&user.Email,
		&user.Username,
		&user.Password,
		&user.Id,
		&user.Verified,
		&user.Avatar,
		&user.CreatedAt,
	); err != nil {
		log.Println(err)
		return nil
	}
	return &user
}

func IsOAuthUser(id string) bool {
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM o_users WHERE id = $1`, id).Scan(&count)
	switch count {
	case 0:
		return false
	default:
		return true
	}
}

func ReadUsers(username string, limit int, offset int) []models.User {
	var users []models.User
	rows, err := db.Query(
		`SELECT * FROM t_users WHERE username LIKE $1 ORDER BY username
		LIMIT $2 OFFSET $3`,
		"%"+username+"%", limit, offset)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var user models.User
		rows.Scan(
			&user.Email,
			&user.Username,
			&user.Password,
			&user.Id,
			&user.Verified,
			&user.Avatar,
			&user.CreatedAt,
		)
		users = append(users, user)
	}
	return users
}

func UpdateUser(id string, updates map[string]any) bool {
	for column := range updates {
		if _, err := db.Exec(
			fmt.Sprintf(`UPDATE t_users SET %s = $1 WHERE id = $2`, pq.QuoteIdentifier(column)),
			updates[column], id,
		); err != nil {
			log.Println(err)
			return false
		}
	}
	return true
}

func DeleteUser(id string) bool {
	if _, err := db.Exec(`DELETE FROM t_users WHERE id = $1`, id); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func Followed(userId string, followId string) bool {
	var count int
	db.QueryRow(
		`SELECT COUNT(*) FROM follows WHERE user_id = $1 AND follow_id = $2`,
		userId, followId,
	).Scan(&count)

	switch count {
	case 0:
		return false
	default:
		return true
	}
}

func ToggleFollow(userId string, followId string) {
	var query string
	voted := Followed(userId, followId)

	switch voted {
	case false:
		query = `INSERT INTO follows(user_id, follow_id) VALUES ($1, $2)`
	default:
		query = `DELETE FROM follows WHERE user_id = $1 AND follow_id = $2`
	}
	if _, err := db.Exec(query, userId, followId); err != nil {
		log.Println(err)
	}
}

func ReadFollowers(userId string) []string {
	var followers []string
	rows, err := db.Query(
		`SELECT username FROM t_users WHERE id in
		(SELECT user_id FROM follows WHERE follow_id = $1)`,
		userId,
	)
	if err != nil {
		log.Println(err)
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		var username string
		rows.Scan(&username)
		followers = append(followers, username)
	}
	return followers
}

func ReadFollowersCount(userId string) int {
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM t_users WHERE id in
		(SELECT user_id FROM follows WHERE follow_id = $1)`,
		userId,
	).Scan(&count); err != nil {
		return 0
	}
	return count
}

func ReadFollowing(userId string) []string {
	var followers []string
	rows, err := db.Query(
		`SELECT username FROM t_users WHERE id in
		(SELECT follow_id FROM follows WHERE user_id = $1)`,
		userId,
	)
	if err != nil {
		log.Println(err)
		return nil
	}

	defer rows.Close()
	for rows.Next() {
		var username string
		rows.Scan(&username)
		followers = append(followers, username)
	}
	return followers
}

func ReadFollowingCount(userId string) int {
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM t_users WHERE id in
		(SELECT follow_id FROM follows WHERE user_id = $1)`,
		userId,
	).Scan(&count); err != nil {
		return 0
	}
	return count
}
