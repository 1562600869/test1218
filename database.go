package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

var ValidStyles = []string{"传统", "现代", "创意融合", "儿童启蒙"}
var ValidSubjects = []string{"花鸟", "人物", "山水", "动物", "文字", "其他"}
var ValidLevels = []string{"入门", "初级", "中级", "高级"}

var LevelOrder = map[string]int{
	"入门": 1,
	"初级": 2,
	"中级": 3,
	"高级": 4,
}

func isValidStyle(style string) bool {
	for _, s := range ValidStyles {
		if s == style {
			return true
		}
	}
	return false
}

func isValidSubject(subject string) bool {
	for _, s := range ValidSubjects {
		if s == subject {
			return true
		}
	}
	return false
}

func isValidLevel(level string) bool {
	_, ok := LevelOrder[level]
	return ok
}

type Student struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Style string `json:"style"`
	Level string `json:"level"`
}

type Work struct {
	ID           int64  `json:"id"`
	StudentID    int64  `json:"student_id"`
	StudentName  string `json:"student_name"`
	WorkName     string `json:"work_name"`
	Subject      string `json:"subject"`
	CompleteDate string `json:"complete_date"`
	SizeDesc     string `json:"size_desc"`
	Exhibited    int    `json:"exhibited"`
}

type Exhibition struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Location  string `json:"location"`
}

type ExhibitionWork struct {
	ID            int64  `json:"id"`
	ExhibitionID  int64  `json:"exhibition_id"`
	WorkID        int64  `json:"work_id"`
	WorkName      string `json:"work_name"`
	StudentName   string `json:"student_name"`
}

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	DB.SetMaxOpenConns(1)
	return createTables()
}

func createTables() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			phone TEXT NOT NULL,
			style TEXT NOT NULL,
			level TEXT NOT NULL DEFAULT '入门'
		)`,
		`CREATE TABLE IF NOT EXISTS works (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			work_name TEXT NOT NULL,
			subject TEXT NOT NULL,
			complete_date TEXT NOT NULL,
			size_desc TEXT,
			exhibited INTEGER DEFAULT 0,
			FOREIGN KEY (student_id) REFERENCES students(id)
		)`,
		`CREATE TABLE IF NOT EXISTS promotions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			from_level TEXT NOT NULL,
			to_level TEXT NOT NULL,
			promote_date TEXT NOT NULL,
			FOREIGN KEY (student_id) REFERENCES students(id)
		)`,
		`CREATE TABLE IF NOT EXISTS exhibitions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			location TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS exhibition_works (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			exhibition_id INTEGER NOT NULL,
			work_id INTEGER NOT NULL,
			FOREIGN KEY (exhibition_id) REFERENCES exhibitions(id),
			FOREIGN KEY (work_id) REFERENCES works(id),
			UNIQUE(exhibition_id, work_id)
		)`,
	}

	for _, s := range stmts {
		if _, err := DB.Exec(s); err != nil {
			return fmt.Errorf("create table failed: %w", err)
		}
	}
	return nil
}

func CreateStudent(s *Student) (int64, error) {
	if !isValidStyle(s.Style) {
		return 0, errors.New("风格方向必须是：传统、现代、创意融合、儿童启蒙")
	}
	if !isValidLevel(s.Level) {
		return 0, errors.New("等级必须是：入门、初级、中级、高级")
	}
	result, err := DB.Exec(
		"INSERT INTO students (name, phone, style, level) VALUES (?, ?, ?, ?)",
		s.Name, s.Phone, s.Style, s.Level,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func ListStudents() ([]Student, error) {
	rows, err := DB.Query("SELECT id, name, phone, style, level FROM students ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Student
	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.ID, &s.Name, &s.Phone, &s.Style, &s.Level); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func GetStudent(id int64) (*Student, error) {
	var s Student
	err := DB.QueryRow(
		"SELECT id, name, phone, style, level FROM students WHERE id = ?", id,
	).Scan(&s.ID, &s.Name, &s.Phone, &s.Style, &s.Level)
	if err == sql.ErrNoRows {
		return nil, errors.New("学员不存在")
	}
	return &s, err
}

func PromoteStudent(id int64, toLevel string) error {
	if !isValidLevel(toLevel) {
		return errors.New("等级必须是：入门、初级、中级、高级")
	}
	s, err := GetStudent(id)
	if err != nil {
		return err
	}

	if LevelOrder[toLevel] <= LevelOrder[s.Level] {
		return errors.New("只能向更高等级晋升")
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE students SET level = ? WHERE id = ?", toLevel, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO promotions (student_id, from_level, to_level, promote_date) VALUES (?, ?, ?, ?)",
		id, s.Level, toLevel, time.Now().Format("2006-01-02"),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func CreateWork(w *Work) (int64, error) {
	if !isValidSubject(w.Subject) {
		return 0, errors.New("题材必须是：花鸟、人物、山水、动物、文字、其他")
	}
	result, err := DB.Exec(
		"INSERT INTO works (student_id, work_name, subject, complete_date, size_desc, exhibited) VALUES (?, ?, ?, ?, ?, ?)",
		w.StudentID, w.WorkName, w.Subject, w.CompleteDate, w.SizeDesc, w.Exhibited,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func ListWorks() ([]Work, error) {
	rows, err := DB.Query(`
		SELECT w.id, w.student_id, COALESCE(s.name, ''), w.work_name, w.subject, 
		       w.complete_date, w.size_desc, w.exhibited
		FROM works w LEFT JOIN students s ON w.student_id = s.id
		ORDER BY w.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Work
	for rows.Next() {
		var w Work
		if err := rows.Scan(&w.ID, &w.StudentID, &w.StudentName, &w.WorkName,
			&w.Subject, &w.CompleteDate, &w.SizeDesc, &w.Exhibited); err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	return list, rows.Err()
}

func GetWork(id int64) (*Work, error) {
	var w Work
	err := DB.QueryRow(`
		SELECT w.id, w.student_id, COALESCE(s.name, ''), w.work_name, w.subject,
		       w.complete_date, w.size_desc, w.exhibited
		FROM works w LEFT JOIN students s ON w.student_id = s.id WHERE w.id = ?
	`, id).Scan(&w.ID, &w.StudentID, &w.StudentName, &w.WorkName,
		&w.Subject, &w.CompleteDate, &w.SizeDesc, &w.Exhibited)
	if err == sql.ErrNoRows {
		return nil, errors.New("作品不存在")
	}
	return &w, err
}

func CreateExhibition(e *Exhibition) (int64, error) {
	result, err := DB.Exec(
		"INSERT INTO exhibitions (name, start_date, end_date, location) VALUES (?, ?, ?, ?)",
		e.Name, e.StartDate, e.EndDate, e.Location,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func ListExhibitions() ([]Exhibition, error) {
	rows, err := DB.Query("SELECT id, name, start_date, end_date, location FROM exhibitions ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Exhibition
	for rows.Next() {
		var e Exhibition
		if err := rows.Scan(&e.ID, &e.Name, &e.StartDate, &e.EndDate, &e.Location); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func AddExhibitionWork(exhibitionID, workID int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM exhibition_works WHERE exhibition_id = ? AND work_id = ?",
		exhibitionID, workID,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该作品已在此展览中参展")
	}

	_, err = tx.Exec(
		"INSERT INTO exhibition_works (exhibition_id, work_id) VALUES (?, ?)",
		exhibitionID, workID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE works SET exhibited = 1 WHERE id = ?", workID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func ListExhibitionWorks(exhibitionID int64) ([]ExhibitionWork, error) {
	rows, err := DB.Query(`
		SELECT ew.id, ew.exhibition_id, ew.work_id, w.work_name, COALESCE(s.name, '')
		FROM exhibition_works ew
		LEFT JOIN works w ON ew.work_id = w.id
		LEFT JOIN students s ON w.student_id = s.id
		WHERE ew.exhibition_id = ?
		ORDER BY ew.id DESC
	`, exhibitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ExhibitionWork
	for rows.Next() {
		var ew ExhibitionWork
		if err := rows.Scan(&ew.ID, &ew.ExhibitionID, &ew.WorkID, &ew.WorkName, &ew.StudentName); err != nil {
			return nil, err
		}
		list = append(list, ew)
	}
	return list, rows.Err()
}

type SubjectStat struct {
	Subject string `json:"subject"`
	Count   int    `json:"count"`
}

func MonthlySubjectStats() ([]SubjectStat, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	end := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	rows, err := DB.Query(`
		SELECT subject, COUNT(*) as cnt 
		FROM works 
		WHERE complete_date >= ? AND complete_date < ?
		GROUP BY subject
		ORDER BY cnt DESC
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SubjectStat
	for rows.Next() {
		var s SubjectStat
		if err := rows.Scan(&s.Subject, &s.Count); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}
