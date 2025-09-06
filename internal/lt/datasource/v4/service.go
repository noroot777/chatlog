package repository

import (
	"database/sql"
	"path/filepath"

	"github.com/sjzar/chatlog/internal/lt/model"
	"github.com/sjzar/chatlog/pkg/util"
)

type Service struct {
	// sc *conf.ServerConfig
	db *sql.DB
}

func (s *Service) CheckDB() {
	dbPath := filepath.Join(util.DefaultWorkDir(""), "lt.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic("failed to open database: " + err.Error())
	}
	// defer db.Close()
	s.db = db

	schema1 := `
	CREATE TABLE IF NOT EXISTS tzs (
		ltid TEXT PRIMARY KEY,
		tz TEXT NOT NULL,
		admins TEXT NOT NULL,
		cron TEXT
	);
	`
	schema2 := `
	CREATE TABLE IF NOT EXISTS groups (
		ltid TEXT NOT NULL,
		tz TEXT NOT NULL,
		chatroom TEXT NOT NULL,
		cursor INTEGER DEFAULT 0,
		PRIMARY KEY(ltid, chatroom)
	);
	`
	_, err = s.db.Exec(schema1)
	if err != nil {
		panic("建表失败 tzs: " + err.Error())
	}
	_, err = s.db.Exec(schema2)
	if err != nil {
		panic("建表失败 groups: " + err.Error())
	}
}

func (s *Service) Exec(query string, args ...any) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *Service) GetLtInfo(ltid string) model.LtConfig {
	var tzConf model.LtConfig
	var tzItem model.ConfigItem

	row := s.db.QueryRow("SELECT ltid, tz, admins, cron FROM tzs WHERE ltid=?", ltid)
	row.Scan(
		&tzItem.Ltid,
		&tzItem.Tz,
		&tzItem.Admins,
		&tzItem.Cron,
	)
	tzConf.Tzs = append(tzConf.Tzs, &tzItem)

	rows, err := s.db.Query("SELECT chatroom, cursor FROM groups WHERE ltid=?", ltid)
	if err != nil {
		panic("查询groups失败: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var gr model.ConfigGroupIn
		err := rows.Scan(
			&gr.Chatroom,
			&gr.Cursor,
		)
		if err != nil {
			panic("扫描groups失败: " + err.Error())
		}
		tzItem.Groups = append(tzItem.Groups, &gr)
	}
	return tzConf
}
