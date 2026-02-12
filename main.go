package main

import (
	"fmt" // 新增：印出更詳細的啟動訊息
	"html/template"
	"net/http"
	"os"      // 新增：引入處理環境變數的工具
	"strconv" // 新增：用來把文字轉成數字
	"sync"    // 新增：防止多個人同時造訪造成計算錯誤
)

/* ----------------------------------------------------------- */

// 宣告一個全域變數來存次數
var visitorCount int
var mu sync.Mutex // 這是「互斥鎖」，確保加法時不會出錯

// --- 新增：讀取檔案的函式 ---
func loadCount() int {
	data, err := os.ReadFile("counter.txt")
	if err != nil {
		fmt.Println("⚠️ 找不到 counter.txt，將從 0 開始")
		return 0
	}
	count, err := strconv.Atoi(string(data))
	if err != nil {
		fmt.Println("⚠️ 檔案內容格式錯誤")
		return 0
	}
	fmt.Printf("✅ 成功讀取舊數據：%d\n", count)
	return count
}

// --- 新增：寫入檔案的函式 ---
func saveCount(count int) {
	// 把數字轉成字串，再轉成 byte 寫入檔案
	// 0644 是 Linux 的權限設定，代表「我能讀寫，其他人只能讀」
	os.WriteFile("counter.txt", []byte(strconv.Itoa(count)), 0644)
}

func home(w http.ResponseWriter, r *http.Request) {

	// 每次有人進首頁，數字就加 1
	mu.Lock()
	visitorCount++
	saveCount(visitorCount)
	fmt.Printf("檢測到新造訪！目前總人數：%d | 來源 IP: %s\n", visitorCount, r.RemoteAddr)
	mu.Unlock()

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

	visitorCount = loadCount()

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

	// 這裡微調一下，讓你啟動時就能看到目前讀到了多少人
	fmt.Println("------------------------------------")
	fmt.Printf("🚀 伺服器啟動成功！\n")
	fmt.Printf("📊 目前累積訪客數：%d\n", visitorCount)
	fmt.Printf("🌐 監聽埠號 (Port): %s\n", port)
	fmt.Println("------------------------------------")

	// 這裡必須使用變數 port，不要寫死 :8080
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("伺服器啟動失敗: %v\n", err)
	}
}
