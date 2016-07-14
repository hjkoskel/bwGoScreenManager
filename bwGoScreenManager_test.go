package bwGoScreenManager_test

import (
	"fmt"
    "testing"
    "github.com/hjkoskel/gomonochromebitmap"
    "github.com/hjkoskel/bwGoScreenManager"
    "time"
    "net"
    "image"
)

    
func exampleClient(){
    testfont1:=gomonochromebitmap.GetFont_8x8()
    ac:=bwGoScreenManager.InitAppConnection()
    
    conn,_:=net.Dial("tcp", "127.0.0.1:8888")
    go ac.RunClientSide("testAPP",[]string{},&conn)
    //Demo graphics, spinning line
    gfx:=gomonochromebitmap.NewMonoBitmap(128,64,false)
    for{
        gfx.Fill(image.Rect(0,0,320,200),false)
        for r:=1;r<32;r++{
            gfx.CircleFill(image.Point{X:64,Y:32}, r,true)
            fmt.Printf("Printing on screen %vx%v pic  r=%v\n",gfx.W,gfx.H,r)
            ac.Display<-gfx
            time.Sleep(100 * time.Millisecond)
            if(len(ac.KeyPresses)>0){
                //OK, render as text on screen
                newKey:=<-ac.KeyPresses
                textOnScreen:=fmt.Sprintf("KEY=%v",newKey)
                gfx.Print(textOnScreen,testfont1,8,2,gfx.Bounds(),true,true,false,true)
                ac.Display<-gfx
                time.Sleep(100 * time.Millisecond)
            }    
        }
    }
}

//Example client2 just prints counter
func exampleClient2(){
    testfont1:=gomonochromebitmap.GetFont_8x8()
    ac:=bwGoScreenManager.InitAppConnection()
    
    conn,_:=net.Dial("tcp", "127.0.0.1:8888")
    go ac.RunClientSide("counterAPP",[]string{},&conn)
    gfx:=gomonochromebitmap.NewMonoBitmap(128,64,false)
    counter:=0
    for{
       gfx.Fill(image.Rect(0,0,320,200),false)
       gfx.Print(fmt.Sprintf("COUNT:%v",counter),testfont1,8,2,gfx.Bounds(),true,true,false,true)
       counter++
       ac.Display<-gfx
       time.Sleep(200*time.Millisecond)
    }    
}

    
func Test1(t *testing.T){
    fmt.Printf("----Manager test----\n")
    toDisplay:=make(chan gomonochromebitmap.MonoBitmap,2)
    fromKeyboard:=make(chan string)
    
    manager:=bwGoScreenManager.InitScreenManager(128,64,8888,toDisplay,fromKeyboard)
    go manager.Run()
    
    go bwGoScreenManager.HostWebUI(80,fromKeyboard,toDisplay)
    
    go exampleClient()
    exampleClient2()
}
