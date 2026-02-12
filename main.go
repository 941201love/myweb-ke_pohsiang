package main

import (
	"database/sql"
	"fmt" // 新增：印出更詳細的啟動訊息
	"html/template"
	"net/http"
	"os"      // 新增：引入處理環境變數的工具
	"strings" // 新增：用來轉換資料庫網址格式
	"sync"    // 新增：防止多個人同時造訪造成計算錯誤

	_ "github.com/go-sql-driver/mysql" // 重要：請記得執行 go get github.com/go-sql-driver/mysql
)

/* ----------------------------------------------------------- */

var db *sql.DB
var mu sync.Mutex // 這是「互斥鎖」，確保加法時不會出錯

// 初始化資料庫連線：讀取 Railway 的 MYSQL_URL 並轉換格式
func initDB() {
	// 從環境變數讀取
	rawURL := os.Getenv("MYSQL_URL")
	if rawURL == "" {
		fmt.Println("⚠️ 警告：找不到 MYSQL_URL，將無法儲存訪客數據")
		return
	}

	// 格式轉換魔術：把 mysql://user:pass@host:port/db
	// 轉成 Go 驅動要求的 user:pass@tcp(host:port)/db
	dsn := strings.Replace(rawURL, "mysql://", "", 1)
	dsn = strings.Replace(dsn, "@", "@tcp(", 1)
	parts := strings.Split(dsn, "/")
	if len(parts) > 0 {
		parts[0] = parts[0] + ")"
	}
	dsn = strings.Join(parts, "/")

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("❌ 資料庫連線失敗: %v\n", err)
		return
	}

	// 測試連線是否真的通了
	err = db.Ping()
	if err != nil {
		fmt.Printf("❌ 無法與資料庫建立通訊: %v\n", err)
	} else {
		fmt.Println("✅ 資料庫連線成功！")
	}
}

// 從資料庫更新並抓取最新的訪客數
func getCountFromDB() int {
	if db == nil {
		return 0
	}

	mu.Lock()
	defer mu.Unlock()

	// 1. 先把資料庫裡的數字 +1
	_, err := db.Exec("UPDATE stats SET counter = counter + 1 WHERE id = 1")
	if err != nil {
		fmt.Println("更新失敗:", err)
	}

	// 2. 抓出目前的數字
	var count int
	err = db.QueryRow("SELECT counter FROM stats WHERE id = 1").Scan(&count)
	if err != nil {
		fmt.Println("讀取失敗:", err)
		return 0
	}
	return count
}

func home(w http.ResponseWriter, r *http.Request) {

	visitorCount := getCountFromDB()
	fmt.Printf("檢測到新造訪！目前總人數：%d | 來源 IP: %s\n", visitorCount, r.RemoteAddr)

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "找不到首頁檔案", http.StatusInternalServerError)
		return
	}

	// 重點：把 visitorCount 傳進 Execute 的第二個參數
	t.Execute(w, visitorCount)
}

func about(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/about.html")
	t.Execute(w, nil)
}

func projects(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/projects.html")
	t.Execute(w, nil)
}

func awards(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/awards.html")
	t.Execute(w, nil)
}

/* ----------------------------------------------------------- */

func main() {

	// 啟動時先連線資料庫
	initDB()

	// 當 Google 來找這個檔案時，直接把檔案內容讀給它看
	http.HandleFunc("/google2d7020435e6908ed.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "google2d7020435e6908ed.html")
	})

	http.Handle("/favicon.png", http.FileServer(http.Dir(".")))

	// 1. 靜態檔案設定
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 2. 路由設定
	http.HandleFunc("/", home)
	http.HandleFunc("/about", about)
	http.HandleFunc("/projects", projects)
	http.HandleFunc("/awards", awards)

	// 3. 重要修改：自動偵測 Render 分配的 Port
	// Render 會透過環境變數傳入 PORT，如果沒有則預設 8080 (本地測試用)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("------------------------------------")
	fmt.Printf("🚀 伺服器啟動成功！Port: %s\n", port)
	fmt.Println("------------------------------------")

	// 這裡必須使用變數 port，不要寫死 :8080
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("伺服器啟動失敗: %v\n", err)
	}
}
