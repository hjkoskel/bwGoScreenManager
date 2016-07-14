package bwGoScreenManager


import (
	"fmt"
    "net"
    "bufio"
    "github.com/hjkoskel/gomonochromebitmap"
    "strings"
    "image"
    "encoding/base64"
    "strconv"
)


//App connection. Scrollin etc.. are handled on application software
type AppConnection struct{
    Display chan gomonochromebitmap.MonoBitmap
    Notifications chan string
    KeyPresses chan string
    Quit chan bool
    
    Keylist []string
}

func InitAppConnection() AppConnection{
    return AppConnection{
        Notifications:make(chan string,1),
        KeyPresses:make(chan string,10),
        Quit:make(chan bool,1),
        Display:make(chan gomonochromebitmap.MonoBitmap,1),
    }
}

//Runs server side
func(p* AppConnection)RunServerSide(name string,connection *net.Conn){
    fmt.Printf("Run server side")
    //Key scanner
    go func(){
        for{
            ke:=<-p.KeyPresses
            fmt.Printf("TODO send %v key",ke);
            msg:=fmt.Sprintf("KEY:%v\n",ke)
            (*connection).Write([]byte(msg))        
        }
    }()
        
    scn:=bufio.NewScanner(*connection)    
    for(scn.Scan()){
        textline:=scn.Text()    
        //fmt.Printf("Yhteys %v sanoo %v\n",name,textline);
        arr:=strings.Split(textline,":")
        data:=strings.Join(arr[1:],":")
        
        //Lets see
        switch(arr[0]){
            case "NOTIFY":
                //TODO IMPLEMENT
            case "LCD128x64": //Dummy [0,0,1,1,.... array
            case "BASE64IMAGE":
                darr:=strings.Split(data," ")
                pw,_:=strconv.ParseInt(darr[0],10,32)
                ph,_:=strconv.ParseInt(darr[1],10,32)    
                    
                newImg:=gomonochromebitmap.NewMonoBitmap(int(pw),int(ph),false)
                //Parsing BASE64...TODO bytearr to uint32 arr
                bytedata, _ := base64.StdEncoding.DecodeString(darr[2])
                for i:=0;i<len(newImg.Pix);i++{
                    newImg.Pix[i]=(uint32(bytedata[i*4+3])<<24)|(uint32(bytedata[i*4+2])<<16)|(uint32(bytedata[i*4+1])<<8)|uint32(bytedata[i*4]);
                }
                p.Display<-newImg            
            case "NEEDKEY": //List of keys used by client All on on message at startup..prefered
                p.Keylist=append(p.Keylist,data)
            case "TESTIMAGE":
                fmt.Printf("Lahetetaan testi-image\n");
                testikuva:=gomonochromebitmap.NewMonoBitmap(100,100,false)
                testikuva.Fill(image.Rect(40,20,60,40),true)
                testikuva.Fill(image.Rect(50,30,80,60),false)
                p.Display<-testikuva
                fmt.Printf("Testikuva lähetetty!!\n")
        }
        if(len(p.Quit)>0){
            break
        }
        
    }
    (*connection).Close()
}


//Runs client side. Client software pushes images. This recieves keypresses
func(p* AppConnection)RunClientSide(name string,usedKeys []string,connection *net.Conn){
    //Send name
    (*connection).Write([]byte(fmt.Sprintf("%v\n",name)))
    //Send list of keys
    for _,needkey:=range usedKeys{
        (*connection).Write([]byte(fmt.Sprintf("NEEDKEY:%v\n",needkey)))
    }
        
        
    //Recieve keypresses and quits
    go func(){
        scn:=bufio.NewScanner(*connection);
        for(scn.Scan()){
            textline:=scn.Text()
            arr:=strings.Split(textline,":")
            data:=strings.Join(arr[1:],":")
            switch(arr[0]){ //TODO Other like lost visibility etc..
                case "KEY":
                    p.KeyPresses<-data
            }
        }
        p.Quit<-true
    }()
    
    //Send image
    for(len(p.Quit)==0){
        fmt.Printf("App connection display buf len is %v\n",len(p.Display))
        var img gomonochromebitmap.MonoBitmap
        //for(len(p.Display)>0){
            img=<-p.Display //Skips frames, latest is important
        //}
        //Print dimensions and base64 data
        //fmt.Printf("APP CONNECTION RECIEVED IMAGE %vx%v\n",img.W,img.H)    
        //Convert uint32 bytedata to bytearr
        bytedata:=make([]byte,len(img.Pix)*4)
        for i:=0;i<len(img.Pix);i++{
            bytedata[i*4+3]=byte((img.Pix[i]>>24)&0xFF)
            bytedata[i*4+2]=byte((img.Pix[i]>>16)&0xFF)
            bytedata[i*4+1]=byte((img.Pix[i]>>8)&0xFF)
            bytedata[i*4]=byte((img.Pix[i])&0xFF)
        }
        msg:=fmt.Sprintf("BASE64IMAGE:%v %v %v\n",img.W,img.H,base64.StdEncoding.EncodeToString(bytedata))
        (*connection).Write([]byte(msg))
    }
    fmt.Printf("APP CONNECTION GOT QUIT\n");
}
