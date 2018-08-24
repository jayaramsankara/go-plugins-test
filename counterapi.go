// counterapi
package main

import (
	"fmt"
	"plugin"
	"os"
	"github.com/gorilla/mux"
	"github.com/fsnotify/fsnotify"
	"strconv"
	"net/http"
	"log"
)
var counterFunc  func() []int;

func count(rw http.ResponseWriter, req *http.Request) {
	cnts := counterFunc()
	fmt.Printf("%v", cnts)
	fmt.Println("DONE")
	rw.Write([]byte("OK"))
}

func loadCounterPlugin() {
	
	plug, err := plugin.Open("counter.so")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Plugin loaded is %v",*plug)
	
	counterSym, err := plug.Lookup("Counter")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var ok bool
	counterFunc, ok = counterSym.(func() []int)
    if !ok {
        panic("Plugin has no 'Counter() []int' function")
    } else {
		fmt.Printf("new plugin synmbol loaded %v",counterSym)
		fmt.Println("")
	}
	
}
func main() {
	fmt.Println("Hello World!")
	
	loadCounterPlugin()
	
	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()
	watcher.Add("counter.so")
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("*** event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("*** modified file:", event.Name)
				}  else if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("*** created file:", event.Name)
					loadCounterPlugin()
				}
				
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/count", count).Methods("GET")
	http.ListenAndServe("127.0.0.1"+":"+strconv.Itoa(8091), r)
	
	
   	
	<-done
}
