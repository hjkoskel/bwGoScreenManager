/**
HTML5 UI
Hosted by this module.

Used when testing on PC without actual oled/led/lcd.. display and keyboard
**/

package bwGoScreenManager


import (
	"fmt"
    "io"
    "net/http"
	"golang.org/x/net/websocket"
    "github.com/hjkoskel/gomonochromebitmap"
    "image"
    "bytes"
    "image/png"
    "image/color"
    "strconv"
)

//TODO WS implementation later
func wsHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}


func writePNG(w http.ResponseWriter, img *image.Image) {
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, *img); err != nil {
		fmt.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		fmt.Println("unable to write image.")
	}
}

var imgNow image.Image  //TODO Mutexlock?    
    
func serveScreen(w http.ResponseWriter, r *http.Request) {
    writePNG(w,&imgNow)
}

var keyPressChan *chan string
func serveKey(w http.ResponseWriter, r *http.Request) {
    key:=r.URL.Query().Get("key")
    
    fmt.Printf("UI key is %#v\n",key)
    switch(key){
        case "BKSP":
            (*keyPressChan)<-"\b"
        case "ENTER":
            (*keyPressChan)<-"\n"
        case "TAB":
            (*keyPressChan)<-"\t"
        default:
            (*keyPressChan)<-key
    }
    (*keyPressChan)<-""
}


func HostWebUI(port int, web_KeyPresses chan string,web_ImageOutput chan gomonochromebitmap.MonoBitmap){
    go func(){
        colTrue:=color.RGBA{R:200,G:200,B:200,A:255}
        colFalse:=color.RGBA{R:40,G:40,B:40,A:255}
        for{
        bw:=<-web_ImageOutput
        imgNow=bw.GetImage(colTrue,colFalse)
        }
    }()
    
    keyPressChan=&web_KeyPresses
    http.Handle("/", http.FileServer(http.Dir("./staticFiles/")))
    http.Handle("/ws", websocket.Handler(wsHandler))
    
    http.HandleFunc("/screen.png",serveScreen)
    http.HandleFunc("/key",serveKey)
        
    fmt.Printf("Serving web ui on port %v\n",port)
    err := http.ListenAndServe(fmt.Sprintf(":%v",port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}