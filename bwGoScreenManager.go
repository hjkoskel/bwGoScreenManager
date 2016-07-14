/*
Black and white screen manager library

-Runs socket SERVER for incoming connections,
-Any application can take socket CLIENT connection to screen (client applications are running in background)
-Can host optional web ui for testing screen manager and client software without specific hardware
-Use AppConnection module when connecting to screen manager
-Manager catches keys "\n",LEFT,RIGHT,HELP from KeyPresses

This manager is resolution and hardware independent. Keys are just strings (mapped by actual software) and display is only monochrome bitmap. 

Screen manager have AppSelected attribute
false)
Manager shifts view on LEFT and RIGHT keypresses, newline key "\n" makes app selected
true) Manager redirects commands to software. Key "EXIT" will make appSelected false

13.7.2016
At the moment notifications are not yet supported. My focus is now on applications. And make large bugfix commit after that


*/

package bwGoScreenManager

import (
	"fmt"
    "github.com/hjkoskel/gomonochromebitmap"
    "net"
    "bufio"
    "time"
)

type ScreenManager struct{
    ActiveApp string //Application name
    AppSelected bool //If not selected. blink box around display + name?
    ScreenW int //Screen dimensions, this keep target size visible
    ScreenH int
    AppChanged chan string //Notify, application change
    KeyPresses chan string  //Presses. Empty string "" means keys released
    ImageOutput chan gomonochromebitmap.MonoBitmap //Assigned by who is using manager
    RunningApps map[string]*AppConnection //Connecte
    listenConn net.Listener //Private, listens connection
}

func InitScreenManager(w int,h int,tcpPort int,imageOut chan gomonochromebitmap.MonoBitmap,keys chan string) *ScreenManager{
    fmt.Printf("Initializing socket server\n");
    lis, err := net.Listen("tcp", fmt.Sprintf(":%v",tcpPort))
    if(err!=nil){
        fmt.Printf("Listening error %v\n",err)
        return nil
    }
    return &ScreenManager{
        ScreenW:w,ScreenH:h,
        KeyPresses:keys,
        ImageOutput:imageOut,
        RunningApps: make(map[string]*AppConnection),
        listenConn:lis,
    }
}

/*
Lists All running programs. Returns index of currently selected program
*/
func (p* ScreenManager)ProgNameList() ([]string,int){
    progNameList:=[]string{}
    selectedIndex:=-1
    for k := range p.RunningApps{
        if(k==p.ActiveApp){
            selectedIndex=len(progNameList)
        }
        progNameList = append(progNameList, k)
    }
    return progNameList,selectedIndex
}


/*
Switch to next program
*/
func (p* ScreenManager)ChangeProg(forward bool){
    if(len(p.RunningApps)<2){//Impossibru!!! -_-
        fmt.Printf("ERROR: Only one or zero progs running\n")
        return
    }
    names,index:=p.ProgNameList()
    //Deattach current prog
    if(index>-1){
        if(forward){
            p.ActiveApp=names[(index+1)%len(p.RunningApps)]
        }else{
            p.RunningApps[p.ActiveApp].Display=make(chan gomonochromebitmap.MonoBitmap,1) //Make active app writing to another chan TODO copy here when chancing
            if(index>0){
                p.ActiveApp=names[index-1]
            }else{
                p.ActiveApp=names[len(p.RunningApps)-1]
            }
        }
    }else{
        p.ActiveApp=names[0]
    }
    fmt.Printf("\n\n!!!!! ACTIVE APP IS NOW %v !!!!!\n",p.ActiveApp)
}


//Runs screen manager
func (p* ScreenManager)Run(){
    //Handle input. Low level has handled display on/off
    go func(){
        for{
            key:=<-p.KeyPresses
            switch(key){
                case "EXIT": //Exits from current program
                    p.AppSelected=false
                case "\n": //Selects program
                    p.AppSelected=true
                case "LEFT": //TODO map other keys
                    if(!p.AppSelected){//Ok to scroll next app
                        p.ChangeProg(true)
                    }
                case "RIGHT":
                    if(!p.AppSelected){//Ok to scroll prev app
                        p.ChangeProg(false)
                    }
                case "HELP": //TODO display something usefull
            }
            fmt.Printf("Manager got key %v, AppSelected=%v active=%v\n",key,p.AppSelected,p.ActiveApp)
            if(p.AppSelected)&&(len(p.ActiveApp)>0){
                app,ok:=p.RunningApps[p.ActiveApp]
                fmt.Printf("Running app found ok=%v\n",ok)
                if(ok){
                    fmt.Printf("Giving keypresses to app key=%v, number of keys waiting=%v\n",key,len(app.KeyPresses))
                    app.KeyPresses<-key
                }
            }
        }
    }()
    
    //Handle display. Needs to load images in cache when chancing in between apps
    go func(){
        for{
            noUpdates:=true //If no updates, sleep and give time to other procecces
            for k,app := range p.RunningApps{
                fmt.Printf("Check display %v\n",k)
                if(len(app.Display)>0){
                    newPic:=<-app.Display
                    if(k==p.ActiveApp){
                        noUpdates=false
                        p.ImageOutput<-newPic
                    }
                }
            }
            if(noUpdates){
                time.Sleep(300*time.Millisecond)
            }
        }
    }()
    
    
    //Listen new connections in background
    for{
        conn, err := p.listenConn.Accept()
        if err != nil {
            fmt.Printf("Socket connection problem %v\n",err)
        }
        fmt.Print("socket connect taken..prepare asking name") //TODO bug, this can jam...goroutine or timeout
        scn:=bufio.NewScanner(conn)
        if(scn.Scan()){
            name:=scn.Text()    
            fmt.Printf("the name of program is %v\n",name);
            newApp:=InitAppConnection()
            if(p.ActiveApp==""){
                p.ActiveApp=name
            }
            go newApp.RunServerSide(name,&conn)
            p.RunningApps[name]=&newApp
        }
    }
    
}
