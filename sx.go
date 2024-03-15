package main

import (
    "bufio"
    "crypto/tls"
    "flag"
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
    "net/url"
    "os"
    "strings"
    "sync"
    "time"
)

var (
    httpClient = &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
            DialContext: (&net.Dialer{
                Timeout:   30 * time.Second,
                KeepAlive: 30 * time.Second,
                DualStack: true,
            }).DialContext,
        },
    }
    maxRetries  int
    concurrency int
)

func init() {
    flag.IntVar(&maxRetries, "retries", 3, "Maximum number of retries for each request")
    flag.IntVar(&concurrency, "concurrency", 10, "Number of concurrent requests to process")
}

type paramCheck struct {
    url   string
    param string
}

func main() {
    startTime := time.Now() // Record the start time

    flag.Parse()

    httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    }

    sc := bufio.NewScanner(os.Stdin)

    initialChecks := make(chan paramCheck, concurrency)
    var wg sync.WaitGroup
    var processWg sync.WaitGroup

    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for check := range initialChecks {
                processCheck(check, &processWg)
            }
        }()
    }

    go func() {
        for sc.Scan() {
            processWg.Add(1)
            initialChecks <- paramCheck{url: sc.Text()}
        }
        processWg.Wait()
        close(initialChecks)
    }()

    wg.Wait()

    duration := time.Since(startTime) // Calculate the duration
    fmt.Printf("Total execution time: %s\n", duration)
}
func processCheck(c paramCheck, wg *sync.WaitGroup) {
    defer wg.Done()
    reflected, err := checkReflectedWithRetry(c.url)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error from checkReflected: %s\n", err)
        return
    }
    if len(reflected) == 0 {
        return
    }
    for _, param := range reflected {
        checkParam(c.url, param)
    }
}

func checkParam(url, param string) {
    for _, char := range []string{"\"", "'", "<", ">"} {
        wasReflected, err := checkAppendWithRetry(url, param, "aprefix"+char+"asuffix")
        if err != nil {
            fmt.Fprintf(os.Stderr, "error from checkAppend for url %s with param %s with %s: %s\n", url, param, char, err)
            continue
        }
        if wasReflected {
            fmt.Printf("param %s is reflected and allows %s on %s\n", param, char, url)
        }
    }
}

func checkReflectedWithRetry(targetURL string) ([]string, error) {
    var err error
    var reflected []string
    for i := 0; i < maxRetries; i++ {
        reflected, err = checkReflected(targetURL)
        if err == nil {
            return reflected, nil
        }
        time.Sleep(2 * time.Second) // Simple backoff, consider exponential backoff for production
    }
    return nil, err
}

func checkAppendWithRetry(targetURL, param, suffix string) (bool, error) {
    var err error
    var wasReflected bool
    for i := 0; i < maxRetries; i++ {
        wasReflected, err = checkAppend(targetURL, param, suffix)
        if err == nil {
            return wasReflected, nil
        }
        time.Sleep(2 * time.Second) // Simple backoff, consider exponential backoff for production
    }
    return false, err
}

// Your existing checkReflected function here (unchanged).


func checkReflected(targetURL string) ([]string, error) {

	out := make([]string, 0)

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return out, err
	}

	// temporary. Needs to be an option
	req.Header.Add("User-Agent", "User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.100 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return out, err
	}
	if resp.Body == nil {
		return out, err
	}
	defer resp.Body.Close()

	// always read the full body so we can re-use the tcp connection
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}

	// nope (:
	if strings.HasPrefix(resp.Status, "3") {
		return out, nil
	}

	// also nope
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(ct, "html") {
		return out, nil
	}

	body := string(b)

	u, err := url.Parse(targetURL)
	if err != nil {
		return out, err
	}

	for key, vv := range u.Query() {
		for _, v := range vv {
			if !strings.Contains(body, v) {
				continue
			}

			out = append(out, key)
		}
	}

	return out, nil
}

func checkAppend(targetURL, param, suffix string) (bool, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return false, err
	}

	qs := u.Query()
	val := qs.Get(param)
	//if val == "" {
	//return false, nil
	//return false, fmt.Errorf("can't append to non-existant param %s", param)
	//}

	qs.Set(param, val+suffix)
	u.RawQuery = qs.Encode()

	reflected, err := checkReflected(u.String())
	if err != nil {
		return false, err
	}

	for _, r := range reflected {
		if r == param {
			return true, nil
		}
	}

	return false, nil
}
