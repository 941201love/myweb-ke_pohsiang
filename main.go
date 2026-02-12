package main

import (
	"fmt" // 新增：印出更詳細的啟動訊息
	"html/template"
	"net/http"
	"os" // 新增：引入處理環境變數的工具
	// 新增：防止多個人同時造訪造成計算錯誤
)

/* ----------------------------------------------------------- */

func home(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "找不到首頁檔案", http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
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

	fmt.Printf("伺服器準備就緒，正在監聽 Port: %s\n", port)

	// 這裡必須使用變數 port，不要寫死 :8080
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("伺服器啟動失敗: %v\n", err)
	}
}
