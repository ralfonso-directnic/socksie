package main

// socksie is a SOCKS4/5 compatible proxy that forwards connections via
// SSH to a remote host

import (
	"flag"
	"fmt"
	"log"
	"net"
	"github.com/spf13/viper"
    "github.com/getlantern/systray"
    "github.com/rob121/vhelp"
	"golang.org/x/crypto/ssh"
    "github.com/ralfonso-directnic/socksie/icon"
)
//windows ubild env GO111MODULE=on go build -ldflags "-H=windowsgui"

 var config string
 var v *viper.Viper
 var addrs string
 var addr string
 var done chan bool

func init() { 
    
    flag.Parse() 
    done = make(chan bool)
    
}

type Dialer interface {
	DialTCP(net string, laddr, raddr *net.TCPAddr) (net.Conn, error)
}

func main() {
    
    
    flag.StringVar(&config,"config","config","Config File Name")
    
    vhelp.Load(config)
    
    var verr error
    
    v,verr = vhelp.Get(config)
    
    if(verr!=nil){}
    
  
    
    host := v.GetString("host")
    
    port := v.GetString("port")
    
    sshport := v.GetString("sshport")
    
    addrs = fmt.Sprintf("%s:%s", host, sshport)
    
    addr = fmt.Sprintf("%s:%s", "127.0.0.1",port)
    
    systray.Run(onReady, onExit)
    
}

func connectUp() {
    

    pass := v.GetString("password") 
    
	var auths []ssh.AuthMethod
	if pass != "" {
		auths = append(auths, ssh.Password(pass))
	}

	sconfig := &ssh.ClientConfig{
		User: v.GetString("user"),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: auths,
	}
	
	conn, err := ssh.Dial("tcp", addrs, sconfig)
	
	if err != nil {
		log.Fatalf("unable to connect to [%s]: %v", addrs, err)
	}
	
	defer conn.Close()
	
	l, err := net.Listen("tcp", addr)
	
	if err != nil {
		log.Fatalf("unable to listen on SOCKS port [%s]: %v", addr, err)
	}
	
	defer l.Close()
	log.Printf("listening for incoming SOCKS connections on [%s]\n", addr)


    var shutdown bool
    
    shutdown = false

    go func () {
        
        <-done
        log.Println("Got Done!")
        l.Close()
        shutdown = true
        conn.Close()
        
    }()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("failed to accept incoming SOCKS connection: %v", err)
			
			if(shutdown==true){
    			
    			return
			}
			
			continue
		}
		go handleConn(c.(*net.TCPConn), conn)
	}
	
	log.Println("waiting for all existing connections to finish")
	connections.Wait()
	log.Println("shutting down")
}

func onReady() {
    
	systray.SetIcon(icon.Data)
	systray.SetTitle(fmt.Sprintf("%s Proxy",v.GetString("instance")))
	systray.SetTooltip(fmt.Sprintf("Connected to %s connect on %s",addrs,addr))
	
	mChecked := systray.AddMenuItemCheckbox(fmt.Sprintf("Connect to %s",addrs), "Click to activate", false)
	
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	// Sets the icon of a menu item. Only available on Mac and Windows.
	mQuit.SetIcon(icon.Data)
	
	go func (){
    	
      
    	
		for {
			select {
			case <-mChecked.ClickedCh:
				if mChecked.Checked() {
					mChecked.Uncheck()
					
					done <- true
					
					mChecked.SetTitle(fmt.Sprintf("Connect to %s",addrs))
				} else {
					mChecked.Check()
					
					go connectUp()
					
					mChecked.SetTitle(fmt.Sprintf("Listening on %s",addr))
				}
		    }
		  }		
    	
    	
    	
	}()
	
	go func() {
		<-mQuit.ClickedCh
		fmt.Println("Requesting quit")
		systray.Quit()
		fmt.Println("Finished quitting")
	}()
	
}

func onExit() {
	// clean up here
}
