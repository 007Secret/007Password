package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	// 使用mutecomm/go-sqlcipher进行SQLite加密
	"github.com/007Secret/007Password/models"
	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// DB 全局数据库连接
var DB *sql.DB

// 数据库路径
const (
	dbFolder = "data"
	dbFile   = "passwordManager.db"
)

// 全局保存当前的主密码
var currentMasterPassword string

// InitDB 初始化数据库连接（无加密）
func InitDB() error {
	var err error

	// 确保数据目录存在
	err = createDataDirIfNotExist()
	if err != nil {
		log.Printf("Failed to create data directory: %v", err)
		return err
	}

	// 数据库连接字符串
	dbPath := filepath.Join(dbFolder, dbFile)
	connectionString := fmt.Sprintf("%s?_foreign_keys=on", dbPath)
	log.Printf("使用数据库路径: %s", dbPath)

	// 打开数据库连接
	DB, err = sql.Open("sqlite3", connectionString)
	if err != nil {
		log.Printf("Failed to open database: %v", err)
		return err
	}

	// 测试连接
	err = DB.Ping()
	if err != nil {
		log.Printf("Failed to ping database: %v", err)
		return err
	}

	// 初始化数据库表
	err = initTables()
	if err != nil {
		log.Printf("Failed to initialize tables: %v", err)
		return err
	}

	log.Println("数据库初始化成功")
	return nil
}

// InitDBWithKey 使用加密密钥初始化数据库
func InitDBWithKey(key string) error {
	// 保存主密码
	currentMasterPassword = key
	log.Printf("设置数据库加密密钥：%s", maskString(key))

	var err error

	// 确保数据目录存在
	err = createDataDirIfNotExist()
	if err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// 关闭现有连接
	if DB != nil {
		log.Printf("关闭现有数据库连接")
		DB.Close()
		DB = nil
	}

	// 数据库路径
	dbPath := filepath.Join(dbFolder, dbFile)
	log.Printf("使用数据库路径：%s", dbPath)

	// 检查文件是否存在
	fileExists := true
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		fileExists = false
		log.Printf("数据库文件不存在，将创建新文件")
	} else {
		log.Printf("数据库文件已存在")
	}

	// 如果文件存在，检查是否已经加密，若未加密则需要备份并删除
	if fileExists {
		// 先尝试以非加密模式打开
		log.Printf("尝试以非加密模式打开检查是否为未加密数据库")
		tempDB, tempErr := sql.Open("sqlite3", dbPath)
		if tempErr == nil {
			// 测试能否成功打开
			pingErr := tempDB.Ping()
			if pingErr == nil {
				log.Printf("检测到未加密数据库，需要进行备份并创建新的加密数据库")
				// 关闭非加密连接
				tempDB.Close()

				// 备份文件
				backupPath := dbPath + ".bak." + time.Now().Format("20060102150405")
				log.Printf("备份原数据库到: %s", backupPath)
				if err := copyFile(dbPath, backupPath); err != nil {
					return fmt.Errorf("备份数据库失败: %w", err)
				}

				// 删除原始文件，以便创建新的加密数据库
				log.Printf("删除原始非加密数据库文件")
				if err := os.Remove(dbPath); err != nil {
					return fmt.Errorf("删除原数据库失败: %w", err)
				}

				fileExists = false
			} else {
				tempDB.Close()
				log.Printf("数据库打开失败，可能已加密或损坏: %v", pingErr)
			}
		} else {
			log.Printf("尝试以非加密模式打开失败: %v", tempErr)
		}
	}

	// 使用SQLCipher文档推荐的DSN格式
	dsn := fmt.Sprintf("%s?_pragma_key=%s&_pragma_cipher_page_size=4096&_foreign_keys=on&_journal_mode=WAL",
		dbPath, url.QueryEscape(key))
	log.Printf("正在连接数据库，使用DSN参数设置密钥")

	// 打开连接
	DB, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("无法打开数据库连接: %w", err)
	}

	// 设置数据库超时和连接池
	DB.SetConnMaxLifetime(time.Hour)
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	// 验证连接
	var version string
	err = DB.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		DB.Close()
		DB = nil
		return fmt.Errorf("验证数据库连接失败(密钥可能不正确): %w", err)
	}

	log.Printf("成功连接到SQLite (版本 %s)", version)

	// 如果是新数据库，初始化表结构
	if !fileExists {
		log.Printf("初始化新数据库表结构")
		err = initTables()
		if err != nil {
			DB.Close()
			DB = nil
			return fmt.Errorf("创建数据库表失败: %w", err)
		}
		log.Printf("成功创建数据库表")
	}

	// 进行最终的ping测试
	err = DB.Ping()
	if err != nil {
		DB.Close()
		DB = nil
		return fmt.Errorf("数据库ping测试失败: %w", err)
	}

	log.Printf("数据库连接和初始化完成，加密有效")
	return nil
}

// 辅助函数：掩盖字符串，用于安全日志记录
func maskString(s string) string {
	if len(s) <= 2 {
		return "***"
	}
	return s[:1] + "***" + s[len(s)-1:]
}

// 创建数据目录
func createDataDirIfNotExist() error {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 在Docker环境中，确保使用绝对路径
	dataPath := dbFolder
	if !filepath.IsAbs(dataPath) {
		dataPath = filepath.Join(currentDir, dbFolder)
	}

	log.Printf("创建数据目录: %s", dataPath)
	return os.MkdirAll(dataPath, 0755)
}

// 初始化数据库表
func initTables() error {
	// 创建settings表
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// 创建passwords表
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS passwords (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			username TEXT,
			phone TEXT,
			password TEXT NOT NULL,
			website TEXT,
			auth_logins TEXT,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// InitTables 公开初始化表结构的函数
func InitTables() error {
	return initTables()
}

// GetAllPasswords 获取所有密码
func GetAllPasswords() ([]models.Password, error) {
	rows, err := DB.Query("SELECT id, name, username, phone, password, website, auth_logins, notes, created_at, updated_at FROM passwords")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passwords []models.Password
	for rows.Next() {
		var p models.Password
		var authLoginsJSON string
		err := rows.Scan(&p.ID, &p.Name, &p.Username, &p.Phone, &p.Password, &p.Website, &authLoginsJSON, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// 解析JSON格式的AuthLogins
		if authLoginsJSON != "" {
			if err := json.Unmarshal([]byte(authLoginsJSON), &p.AuthLogins); err != nil {
				log.Printf("Error unmarshaling auth_logins: %v", err)
			}
		}

		passwords = append(passwords, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return passwords, nil
}

// GetPasswordByID 通过ID获取密码
func GetPasswordByID(id int) (models.Password, error) {
	var p models.Password
	var authLoginsJSON string

	err := DB.QueryRow("SELECT id, name, username, phone, password, website, auth_logins, notes, created_at, updated_at FROM passwords WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.Username, &p.Phone, &p.Password, &p.Website, &authLoginsJSON, &p.Notes, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return p, err
	}

	// 解析JSON格式的AuthLogins
	if authLoginsJSON != "" {
		if err := json.Unmarshal([]byte(authLoginsJSON), &p.AuthLogins); err != nil {
			log.Printf("Error unmarshaling auth_logins: %v", err)
		}
	}

	return p, nil
}

// CreatePassword 创建新密码
func CreatePassword(p models.Password) (int64, error) {
	// 设置时间戳为当前时间
	currentTime := time.Now()
	p.CreatedAt = currentTime
	p.UpdatedAt = currentTime

	// 将AuthLogins转换为JSON
	authLoginsJSON, err := json.Marshal(p.AuthLogins)
	if err != nil {
		return 0, err
	}

	result, err := DB.Exec(
		"INSERT INTO passwords (name, username, phone, password, website, auth_logins, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		p.Name, p.Username, p.Phone, p.Password, p.Website, string(authLoginsJSON), p.Notes, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdatePassword 更新密码
func UpdatePassword(p models.Password) error {
	// 设置更新时间为当前时间
	p.UpdatedAt = time.Now()

	// 将AuthLogins转换为JSON
	authLoginsJSON, err := json.Marshal(p.AuthLogins)
	if err != nil {
		return err
	}

	_, err = DB.Exec(
		"UPDATE passwords SET name = ?, username = ?, phone = ?, password = ?, website = ?, auth_logins = ?, notes = ?, updated_at = ? WHERE id = ?",
		p.Name, p.Username, p.Phone, p.Password, p.Website, string(authLoginsJSON), p.Notes, p.UpdatedAt, p.ID,
	)
	return err
}

// DeletePassword 删除密码
func DeletePassword(id int) error {
	_, err := DB.Exec("DELETE FROM passwords WHERE id = ?", id)
	return err
}

// GetSetting 获取配置项
func GetSetting(key string) (string, error) {
	if DB == nil {
		return "", fmt.Errorf("数据库连接不存在")
	}

	var value string
	err := DB.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	return value, err
}

// SetSetting 设置配置项
func SetSetting(key, value string) error {
	if DB == nil {
		return fmt.Errorf("数据库连接不存在")
	}

	_, err := DB.Exec(`
		INSERT INTO settings (key, value) 
		VALUES (?, ?) 
		ON CONFLICT(key) DO UPDATE SET value = ?
	`, key, value, value)
	return err
}

// SearchPasswordsByName 通过名称搜索密码
func SearchPasswordsByName(query string) ([]models.Password, error) {
	rows, err := DB.Query("SELECT id, name, username, phone, password, website, auth_logins, notes, created_at, updated_at FROM passwords WHERE name LIKE ?", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passwords []models.Password
	for rows.Next() {
		var p models.Password
		var authLoginsJSON string
		err := rows.Scan(&p.ID, &p.Name, &p.Username, &p.Phone, &p.Password, &p.Website, &authLoginsJSON, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// 解析JSON格式的AuthLogins
		if authLoginsJSON != "" {
			if err := json.Unmarshal([]byte(authLoginsJSON), &p.AuthLogins); err != nil {
				log.Printf("Error unmarshaling auth_logins: %v", err)
			}
		}

		passwords = append(passwords, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return passwords, nil
}

// GetDBFolder 获取数据库文件夹路径
func GetDBFolder() string {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current working directory: %v", err)
		return dbFolder // 使用相对路径作为后备
	}

	// 在Docker环境中，确保使用绝对路径
	dataPath := dbFolder
	if !filepath.IsAbs(dataPath) {
		dataPath = filepath.Join(currentDir, dbFolder)
	}

	return dataPath
}

// 辅助函数：复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
