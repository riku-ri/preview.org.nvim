package main
import (
	//"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"os"
	"os/signal"
	"context"
	//"strconv"
	"syscall"

	"golang.org/x/net/websocket"
)

type File struct {name string; echo []byte; mode os.FileMode;}
type FileInterface interface { write() }
func (f File) write() { os.WriteFile(f.name, f.echo, f.mode) }

func dataSocket(stringchan chan string, file File) {

	var err error
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("[HTTP Server][Data Socket][Listen()]", err)
	}
	addr := listener.Addr().String()
	file.echo = []byte(addr[strings.LastIndex(addr, ":"):])
	file.write()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("[HTTP Server][Data Socket][Accept()]", err)
		}
		//defer listener.Close() /* doing nothing */
		const readBytes = 0x10000
		recieved := ""
		for func (recieved *string) error {
			var read [readBytes]byte
			readed, err := conn.Read(read[:])
			if err != nil {
				//log.Fatal("[Server Data Read]", err)
			}
			*recieved += string(read[:readed])
			return err
		}(&recieved) == nil {}
		stringchan <- recieved
	}

}

func getVimInfo() (vimPreviewRoot string, pid string, bufnr string) {
	b := make([]byte, 1)
	for func () int {
		readed, err := os.Stdin.Read(b)
		if b[0] == 0xff || err != nil {return -1}
		return readed
	}() > 0 {
		vimPreviewRoot += string(b)
	}
	for func () int {
		readed, err := os.Stdin.Read(b)
		if b[0] == 0xff || err != nil {return -1}
		return readed
	}() > 0 {
		pid += string(b)
	}
	for func () int {
		readed, err := os.Stdin.Read(b)
		if b[0] == 0xff || err != nil {return -1}
		return readed
	}() > 0 {
		bufnr += string(b)
	}
	return
}

func main() {

	//vimPreviewRoot, vimPid, vimBufnr := getVimInfo()
	path , fileNameID := os.Args[1] , os.Args[2]
	dataPort := File{
		//name : vimPreviewRoot + "/" + vimPid + "." + vimBufnr + "." + "data.port",
		name : path + "/" + fileNameID + "." + "data.port",
		echo : []byte("0"),
		mode : 0644,
	}
	stringchan := make(chan string)
	go dataSocket(stringchan, dataPort)
	defer func () {
		err := os.Remove(dataPort.name)
		if err != nil {
			log.Printf("[HTTP Server][defer][remove file]: %s", dataPort.name)
		}
	}()

	http.HandleFunc("/", func (w http.ResponseWriter, req *http.Request) {
		wsServer := websocket.Server{Handler: websocket.Handler(
			func (wsConn *websocket.Conn) {
				beforeunload := make(chan byte)
				go func () {
					for {
						var received string
						if
						err := websocket.Message.Receive(wsConn, &received);
						err != nil {break}
						log.Printf("%v\n", received)
					}
					beforeunload <- 0
				}()
				go func () {
					for {
						fromDataSocket := <- stringchan
						if fromDataSocket == "" {continue}
						if
						err := websocket.Message.Send(wsConn, fromDataSocket)
						err != nil {
							stringchan <- fromDataSocket
							/* keep the failed send to after refresh browser */
							break
						}
					}
				}()
				<- beforeunload
			})}
		wsServer.ServeHTTP(w, req)
	})
	/* this will not allow any origin client
		http.Handle("/", websocket.Handler(func (wsConn_p *websocket.Conn) {
			var err error
			if func (err_p *error) error {
				*err_p = websocket.Message.Send(wsConn_p, "<h1>WEBSOCKET</h1>")
				return *err_p
			}(&err) != nil {
				log.Fatal("[Server]:[Websocket]:[Send]:", err)
			}
		}))
	*/

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("[Server]:[TCP]:[Listrrn]:", err)
	}
	defer listener.Close() /* doing nothing */
	addr := listener.Addr().String()

	files := []File{
		//{
		//	name : vimPreviewRoot + "/" + "vim" + "." + "pid",
		//	//echo : []byte(strconv.Itoa(os.Getpid())),
		//	echo : []byte(vimPid),
		//	mode : 0644,
		//},
		{
			//name : vimPreviewRoot + "/" + vimPid + "." + vimBufnr + "." + "websocket.port",
			name : path + "/" + fileNameID + "." + "websocket.port",
			echo : []byte(addr[strings.LastIndex(addr, ":"):]),
			mode : 0644,
		},
		{
			//name : vimPreviewRoot + "/" + vimPid + "." + vimBufnr + "." + "websocket.html",
			name : path + "/" + fileNameID + "." + "websocket.html",
			echo : []byte(`
				<html><head>
				<script type="text/javascript">
					var websocket = null;
					var wsuri = "ws://127.0.0.1` + addr[strings.LastIndex(addr, ":"):] + `";
					window.onload = function () {
						console.log("onload");
						websocket = new WebSocket(wsuri);
						websocket.onopen = function () {};
						websocket.onmessage = function (e) {
							const body = document.getElementsByTagName('body');
							var json = JSON.parse(e.data)
							if("body" in json)
								body[0].innerHTML = json["body"];
							if("move" in json) {
								const anchor = document.getElementById("al".concat(JSON.parse(json["move"])[0]))
								//location.href = "#al".concat(JSON.parse(json["move"])[0])
								//anchor.scrollIntoView({behavior: 'smooth'})
								window.scrollBy({
									top: (anchor.getBoundingClientRect().top - (0.1*window.innerHeight)),
									// here 0.1 is just a random number.
									// to make it center in the page it should be 0.5
									// but center is not always good to see espcially when it is a huge element
									left: 0,
									behavior: "smooth",
								})
							}
						};
					};
					window.onbeforeunload = function () {
						websocket.onclose = function () {};
						websocket.close();
					};
				</script></head>
				<body/>
				</html>
			`),
			mode : 0644,
		},
	}
	for _, f := range files {
		f.write()
	}
	defer func () {
		for _, f := range files {
			err := os.Remove(f.name)
			if err != nil {
				log.Printf("[HTTP Server][defer][remove file]: %s", f.name)
			}
		}
	}()

	var server http.Server
	idleConnsClosed := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	err = server.Serve(listener)
	if err != http.ErrServerClosed {
		log.Fatalf("[HTTP Server]:[Serve()]: %v", err)
	}

}
