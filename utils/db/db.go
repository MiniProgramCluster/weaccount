package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
	"weaccount/internal/conf"
	"weaccount/utils/log"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db   *sql.DB
	once sync.Once
)

// Initialize 初始化数据库连接
func Initialize() error {
	var err error
	once.Do(func() {
		dbConf := conf.Database()
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbConf.User,
			dbConf.Password,
			dbConf.Host,
			dbConf.Port,
			dbConf.Database,
		)
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Error opening database")
			return
		}
		// 设置连接池参数
		db.SetMaxOpenConns(dbConf.PoolMaxOpen)
		db.SetMaxIdleConns(dbConf.PoolMaxIdle)
		db.SetConnMaxLifetime(time.Hour)
		// 测试连接
		err = db.Ping()
		if err != nil {
			log.Logger().Error().Err(err).Msg("Error connecting to the database")
			return
		}
		log.Logger().Info().Str("host", dbConf.Host).
			Int("port", dbConf.Port).
			Str("db_name", dbConf.Database).
			Msg("Database connection established successfully")
	})
	return err
}

func Instance() *sql.DB {
	return db
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// Transaction 执行事务
func Transaction(txFunc func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		}
	}()
	if err := txFunc(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
