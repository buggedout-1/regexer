package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
    "sync"
    "time"
)

var (
    baseURL    string
    domainList string
    searchWord string
    counterMutex sync.Mutex
)

func main() {
    flag.StringVar(&baseURL, "u", "", "Single URL to test, e.g., https://example.com")
    flag.StringVar(&domainList, "l", "", "Path to the file containing the list of URLs.")
    flag.StringVar(&searchWord, "w", "", "Word to search for in response bodies.")
    flag.Parse()

    // Check if either -u or -l is specified
    if searchWord == "" {
        fmt.Println("Usage: ./regexer -l <file_path> -w <word>")
        os.Exit(1)
    }

    if baseURL != "" {
        // Single URL provided via -u flag
        processSingleURL(baseURL)
    } else if domainList != "" {
        // URL file provided via -l flag
        urls, err := readURLsFromFile(domainList)
        if err != nil {
            fmt.Println("Error reading URLs from the file:", err)
            os.Exit(1)
        }

        client := &http.Client{Timeout: 5 * time.Second}
        urlsChannel := make(chan string, len(urls))
        resultsChannel := make(chan string, len(urls))
        done := make(chan bool)

        go startWorkerPool(urlsChannel, resultsChannel, 50, client)
        go processResults(resultsChannel, done)

        for _, url := range urls {
            incrementCounter()
            urlsChannel <- url
        }
        close(urlsChannel)

        <-done
    } else {
        fmt.Println("Usage: ./regexer -l <file_path> or -u <url> -w <word>")
        os.Exit(1)
    }
}

func processSingleURL(url string) {
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        fmt.Println("Error fetching URL:", err)
        os.Exit(1)
    }
    defer resp.Body.Close()

    body, readErr := readResponseBodyWithTimeout(resp.Body, 2*time.Second)
    if readErr != nil {
        fmt.Println("Error reading response:", readErr)
        os.Exit(1)
    }

    if strings.Contains(string(body), searchWord) {
        fmt.Println(url)
    }
}

func startWorkerPool(urls <-chan string, results chan<- string, numWorkers int, client *http.Client) {
    var wg sync.WaitGroup

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            processURLs(urls, results, client)
        }()
    }

    go func() {
        wg.Wait()
        close(results)
    }()
}

func processURLs(urls <-chan string, results chan<- string, client *http.Client) {
    for url := range urls {
        resp, err := client.Get(url)
        if err != nil {
            continue
        }
        defer resp.Body.Close()

        body, readErr := readResponseBodyWithTimeout(resp.Body, 2*time.Second)
        if readErr != nil {
            continue
        }

        if strings.Contains(string(body), searchWord) {
            results <- fmt.Sprintf("%s\n", url)
        }
    }
}

func processResults(results <-chan string, done chan<- bool) {
    for result := range results {
        fmt.Print(result)
    }
    close(done)
}

func readURLsFromFile(filePath string) ([]string, error) {
    var urls []string

    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        urls = append(urls, line)
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return urls, nil
}

func incrementCounter() {
    counterMutex.Lock()
    defer counterMutex.Unlock()
}

func readResponseBodyWithTimeout(body io.Reader, timeout time.Duration) ([]byte, error) {
    done := make(chan struct{})
    var result []byte
    var err error

    go func() {
        defer close(done)
        result, err = io.ReadAll(body)
    }()

    select {
    case <-done:
        return result, err
    case <-time.After(timeout):
        return nil, fmt.Errorf("timeout while reading body")
    }
}
