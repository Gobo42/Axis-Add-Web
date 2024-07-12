package main

import (
    "fmt"
    "bufio"
    "io"
    "os"
    "crypto/tls"
    "net/http"
    "strings"
    "strconv"
    "regexp"
    "bytes"
)


func main() {
    fmt.Println("Axis Web TLD Creator")
    fmt.Println("v1.0.0 by matt.hum@hpe.com")

    apikey := ""
    dat, err := os.ReadFile("apikey")
    if err != nil {
        fmt.Println("Missing API Key")
        fmt.Print("Enter Key here: ")
        reader := bufio.NewReader(os.Stdin)
        text,_ := reader.ReadString('\n')
        text = strings.Replace(text, "\n","",-1)
        
        f, err := os.Create("apikey")
        if err!=nil {
            fmt.Println("Couldn't open file for opening")
        }
        defer f.Close()

        w:=bufio.NewWriter(f)
        _, err = w.WriteString(text)
        if err!=nil {
            fmt.Println("Couldn't write API key to file")
        }
        apikey = text
        w.Flush()
    } else {
        apikey = string(dat)
    }
    bearer:= "Bearer " + strings.TrimSpace(apikey)

    tr := &http.Transport {
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}
    url := "https://admin-api.axissecurity.com/api/v1/connectorzones?pageSize=100&pageNumber=1"
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        fmt.Println("Error formatting request")
        panic(1)
    }
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Authorization", bearer)

    res, err := client.Do(req)
    if err != nil {
        fmt.Println("status code: ", res.StatusCode)
        fmt.Println("Error sending req: ", err)
        panic(1)
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
        fmt.Println("Error, got status code: ", res.StatusCode)
        fmt.Println(res)
        panic(1)
    }
    msg, _ := io.ReadAll(res.Body)
    data := strings.SplitAfter(string(msg),"data\":[")[1]
    var re = regexp.MustCompile(`"connectors[^\]]+\],`)
    s:=re.ReplaceAllString(data, ``)
    connectors := strings.Split(s,"}")
    count:=len(connectors)
    fmt.Println("I found", count, "connector zones")
    type entry struct {
        id string
        name string
    }
    var arr []entry
    for i:=0; i<count; i++ {
        b:= strings.Split(connectors[i],",")
        var t entry
        for j:=0; j<len(b); j++ {
            if strings.Contains(b[j],"id") {
                c:=strings.Split(b[j],"\"")
                t.id = strings.TrimSpace(c[3])
            }
            if strings.Contains(b[j],"\"name\"") {
                c:=strings.Split(b[j],"\"")
                t.name = strings.TrimSpace(c[3])
                fmt.Printf("%v: %v\n",i,c[3])
            }
        }
        arr = append(arr, t)
    }
    text:=""
    connectorzone:=""
    fmt.Print("Enter number of connector zone to egress (or press enter for the Public Connector): ")
    reader := bufio.NewReader(os.Stdin)
    text,_ = reader.ReadString('\n')
    text = strings.Replace(text, "\n","",-1)
    num, err :=strconv.Atoi(text)
    if err != nil {
        fmt.Println("Using Public Connector Zone")
    } else {
        fmt.Println("Using " + arr[num].name)
        connectorzone=",\"connectorZoneID\": \""+arr[num].id+"\""
    }



    tr = &http.Transport {
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client = &http.Client{Transport: tr}
    url = "https://data.iana.org/TLD/tlds-alpha-by-domain.txt"
    req, err = http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        fmt.Println("Error formatting request")
        panic(1)
    }
    
    res, err = client.Do(req)
    if err != nil {
        fmt.Println("status code: ", res.StatusCode)
        fmt.Println("Error sending req: ", err)
        panic(1)
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
        fmt.Println("Error, got status code: ", res.StatusCode)
        fmt.Println(res)
        panic(1)
    }
    msg, _ = io.ReadAll(res.Body)  
    tlds:= strings.Split(string(msg),"\n")
    jsontlds:="["
    for _, tld := range tlds {
        if (!strings.Contains(tld,"#") && (len(tld)!=0)) {
            jsontlds=jsontlds + "\"*." + tld + "\","
        }
    }
    jsontlds = jsontlds[:len(jsontlds)-1]+"]"
    jsreq:="{\"name\": \"All Web\",\"description\": \"All Web TLDs\",\"IncludedDomainsOrUrls\": "+jsontlds+connectorzone+"}"

    //fmt.Println(jsreq)
    fmt.Println("Sending TLDs")
    url="https://admin-api.axissecurity.com/api/v1/webcategories"
    buf:=[]byte(jsreq)

    req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(buf))
    if err != nil {
        fmt.Println("Error formatting request")
        panic(1)
    }
    req.Header.Add("Accept", "application/json")
    req.Header.Add("Authorization", bearer)
    req.Header.Add("Content-Type","application/json")

    client = &http.Client{}
    res, err = client.Do(req)
    if err != nil {
        fmt.Println("status code: ", res.StatusCode)
        fmt.Println("Error sending req: ", err)
        panic(1)
    }
    defer res.Body.Close()
    msg, err = io.ReadAll(res.Body)
    if err != nil {
        fmt.Println(err)
    }
    if res.StatusCode==201 {
        fmt.Println("Success!")
    } else {
        fmt.Print("Got response: ")
        fmt.Println(res.StatusCode)
        fmt.Println(string(msg))
    }
    
}
