package serve

import (
    "fmt"
    "net/http"
    "path/filepath"
    "strings"
)

func (ser *ProxyServe) Broadcast_Req(req *http.Request, id int64, docid uint64, user string) bool {
    data := make(map[string]interface{})
    data["docid"] = fmt.Sprintf("%d", docid)
    data["sid"] = id % 1000
    data["host"] = req.Host
    data["client_ip"] = req.RemoteAddr
    data["path"] = req.URL.Path
    if(req.Method=="CONNECT"){
       data["path"]="https req,unknow path"
    }
    data["method"] = req.Method
    ser.mu.RLock()
    defer ser.mu.RUnlock()
    hasSend := false
    for _, client := range ser.wsClients {
        if(ser.conf.SessionView==SessionView_IP_FILTER && client.filter_client_ip==""){
           continue
        }
        if (client.user == user||user=="guest") && checkFilter(req, client) {
            send_req(client, data)
            hasSend = true
        }
    }
    return hasSend
}


var extTypes map[string][]string = map[string][]string{
    "js":    []string{"js"},
    "css":   []string{"css"},
    "image": []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff", "jpe", "tif", "webp", "ico"},
}

func checkFilter(req *http.Request, client *wsClient) bool {
    addr_info:=strings.Split(req.RemoteAddr,":")
    if client.filter_client_ip != "" && addr_info[0]!=client.filter_client_ip {
        return false
    }
    if len(client.filter_url) > 0 {
        url := req.URL.String()
        has := false
        for _, subUrl := range client.filter_url {
            if strings.Contains(url, subUrl) {
                has = true
                break
            }
        }
        if !has {
            return false
        }
    }
    if len(client.filter_hide) > 0 {
        ext := strings.ToLower(strings.Trim(filepath.Ext(req.URL.Path), "."))
        for _, hide_type := range client.filter_hide {
            for _, hide_ext := range extTypes[hide_type] {
                if ext == hide_ext {
                    return false
                }
            }
        }
    }
    if(len(client.filter_url_hide)>0){
        for _, hide_kw := range client.filter_url_hide {
           if hide_kw!="" && strings.Contains(req.URL.String(),hide_kw) {
                 return false
           }
        }
    }
    return true
}
