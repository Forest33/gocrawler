## Small web crawler library on Go

Example using the callback channnel

    c := crawler.New("https://gobyexample.com/", "", "")
    ch := make(chan *crawler.HttpResponse)
    c.SetCallbackChan(ch)
    c.SetTimeout(10)
    c.SetMaxDepth(1)
    c.SetMaxWorkers(5)
    c.Start()
    
    for {
        select {
        case r := <-ch:
            fmt.Println(r.Uri, r.Header.Get("Content-Type"))
        case <-time.After(time.Second):
        }
        if !c.IsProcessing() {
            break
        }
    }

Example using the callback function

    uri := "https://gobyexample.com/"
    c := crawler.New(uri, "", "")
    c.SetCallbackFunc(cb)
    c.SetTimeout(10)
    c.SetMaxDepth(1)
    c.SetMaxWorkers(5)
    c.Start()
    
    func cb(response *crawler.HttpResponse) {
        fmt.Println(response.Uri, response.Header.Get("Content-Type"))
    }
