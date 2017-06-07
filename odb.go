package main


import (
    "fmt"
    "os"
    "encoding/json"
    "io/ioutil"
  //  "strconv"
    "github.com/rjeczalik/notify"
    "log"
  	"time"
  	"io"
   "strings"
)

type Server struct {
    Host,Alias,Username,Password string
    Port int
} 

type Session struct {
    Directories []Directory
    Servers []Server
    Source string

}



type Directory struct {
    Alias,Server,Path,ServerPath,Original string
    Files []string
}

func DeleteInvisble (server Server,dir Directory) ( ret []string ) {
	for _,v := range dir.Files {
	_, e := os.Stat(v)
    if e != nil {
    	//ftp remove
    	log.Println("FTP REMOVE ", v);
    } else {
    	ret = append(ret, v )
    }
	} 
	return
}



func LoadSession() Session {
    file, e := ioutil.ReadFile("./config.json")
    if e != nil {
        return Session{Directories: []Directory{}, Servers:[]Server{}}
    }
    fmt.Printf("Loading Config %s\n", string(file))

    //m := new(Dispatch)
    //var m interface{}
    var jsontype Session
    json.Unmarshal(file, &jsontype)
    fmt.Printf("Results: %v\n", jsontype)
    return jsontype
}

func SaveSession(slcD Session){
     slcB, _ := json.Marshal(slcD)
     _, e := ioutil.ReadFile("./config.json")
    if e == nil {
      os.Remove("./config.json")
    }

    if  err := ioutil.WriteFile("./config.json", slcB, 0777); err != nil {
    fmt.Printf("%s" ,err)
    }


}

// none supplied launch a daemon

// add connect

// watch

func isValueInServers(value string, list []Server) bool {
    for _, v := range list {
        if v.Alias == value {
            return true
        }
    }
    return false
}

func isValueInDirs(value string, list []Directory) bool {
    for _, v := range list {
        if v.Alias == value {
            return true
        }
    }
    return false
}

func  GetDir(value string, list []Directory) Directory {
    for _, v := range list {
        if v.Alias == value {
            return v
        }
    }
    return Directory{Alias:"__&&"}
}

func  GetSer(value string, list []Server) Server {
    for _, v := range list {
        if v.Alias == value {
            return v
        }
    }
    return Server{Alias:"__&&"}
}

func removeFromServers(value string, list []Server) (servers []Server) {
//	servers = []Server{}

	 for _, v := range list {
        if v.Alias != value {
            servers = append(servers, v)
        }
    }

    return
}

func removeFromDirs(value string, list []Directory) (dirs []Directory) {
//	servers = []Server{}

	 for _, v := range list {
        if v.Alias != value {
            dirs = append(dirs, v)
        }
    }

    return
}

func GetFiles(path string) (ret []string) {
	     files, _ := ioutil.ReadDir(path)

			    for _, f := range files {
			    	if(!f.IsDir()){

			    		ret = append(ret, path + "/" + f.Name())
			    			
			    	}
			          
			    }
	return
}

func RecuWatch(dir string, out chan notify.EventInfo){
	// files, _ := ioutil.ReadDir( dir )
	    go Watch(dir, out);
			    /*  for _, f := range files {
			    	//fmt.Println(f.Name())
			  	if f.IsDir() {
			    		if f.Name() != "backups-appe" {
				    		dirx := Directory{}
				    		dirx.Path = dir + "/" + f.Name()
				    		//dirx.Files = GetFiles(dirx.Path);
				    	 	go RecuWatch(dirx.Path, out)
			    	 	}
			    	} 
			} */
	
}

func Watch(dir string,  out chan notify.EventInfo ) {

	_, err := os.Stat(dir + "/")
		    if err != nil {
		        fmt.Println(err)
		        if(unwatch == "SKIP"){
		        	unwatch = "";
		        } else {
		        unwatch = dir;
		      
		       	}
		        return
		    }

		//    fmt.Println(dir)
	fmt.Println("\x1b[31;1m-> Restore monitor daemon [PATH ] : \x1b[0m", dir);
	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	c := make(chan notify.EventInfo, 1)

	// Set up a watchpoint listening for events within a directory tree rooted
	// at current working directory. Dispatch remove events to c.
	if err := notify.Watch(dir, c, notify.All); err != nil {
	    log.Fatal(err)
	}
	defer notify.Stop(c)
	//
	//<-c
	// Block until an event is received.
	ei := <-c
	go Backup( dir, ei, out)
	
	log.Println("\x1b[31;1m-> ", ei ," \x1b[0m")
	
	Watch(dir, out);
}

var unwatch = "";
var latestwatch = "";

func SyncJ (server Server, dir Directory ){
	if unwatch == "SKIP" {
		unwatch = ""
	}
	if unwatch != "" && latestwatch != "" {
		//perform ftp move
		fmt.Println("\x1b[31;1m->", "Moving directory online ",unwatch , " -> ", latestwatch , "\x1b[0m");
		unwatch = "";
		latestwatch = "";
	} else if unwatch == "" && latestwatch != "" {
		//perform ftp move
		fmt.Println("\x1b[31;1m->", "Moving directory online ",unwatch , " -> ", latestwatch , "\x1b[0m");
		unwatch = "";
		latestwatch = "";
	} 
}

func Copy(dst, src string) error {
    in, err := os.Open(src)
    if err != nil { return err }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil { return err }
    defer out.Close()
    _, err = io.Copy(out, in)
    cerr := out.Close()
    if err != nil { return err }
    return cerr
}

func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

func Backup( dir string,ei notify.EventInfo , out chan notify.EventInfo){

	if ei.Event() == notify.Write {
		//direct upload
		  log.Println("\x1b[34;1m", "Backing up file",  ei.Path(), "\x1b[0m");
		 	slc := strings.Split(ei.Path(), "/")
		 	
		  exis , _ := exists("backups-appe/" +   ( slc[len(slc) - 1] ))
		  if !exis {
		  	os.MkdirAll("backups-appe/"  + ( slc[len(slc) - 1] ),0777);
		  }
		  t := time.Now()
		  Copy("backups-appe/" + ( slc[len(slc) - 1] ) + "/" + t.String(), ei.Path() )
		  //copy file
	}
}

func main() {
	
	//session := LoadSession()
	//fmt.Println(os.Args);
	if len(os.Args) > 1 {
	
			
				
					//connect and watch //create channel
					ch := make(chan notify.EventInfo)
					
					go RecuWatch(os.Args[1],ch)
				
					for _= range ch {
				       // log.Println("\x1b[33;1m",i, "\x1b[0m")
				    }

			
			
	
	


	} else {
		fmt.Println("Usage: odb run wb <folder_alias> ")
	}

}