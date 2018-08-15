package cmd

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"qpush/pkg/logger"

	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
)

func logs(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/wslogs")
}

var upgrader = websocket.Upgrader{} // use default options

func wslogs(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	t, err := tail.TailFile(conf.QPTailFile, tail.Config{Follow: true, MustExist: true, Location: &tail.SeekInfo{Whence: os.SEEK_CUR}})
	if t != nil {
		defer t.Kill(nil)
	}
	if err != nil {
		err = c.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		if err != nil {
			logger.Error("WriteMessage:", err)
		}
		return
	}

	for {
		for line := range t.Lines {
			if line == nil {
				return
			}
			err = c.WriteMessage(websocket.TextMessage, []byte(line.Text))
			if err != nil {
				logger.Error("WriteMessage:", err)
				return
			}
		}
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
	var $log = document.getElementById("log")
	ws = new WebSocket("{{.}}");
	
	var print = function(message) {
		console.log(message)
	}
	ws.onopen = function(evt) {
		print("OPEN");
	}
	ws.onclose = function(evt) {
		print("CLOSE");
		ws = null;
	}
	ws.onmessage = function(evt) {
		print("RESPONSE: " + evt.data);
		var d = document.createElement("div");
		d.innerHTML = evt.data;
		$log.appendChild(d)
	}
	ws.onerror = function(evt) {
		print("ERROR: " + evt.data);
	}
});
</script>
</head>
<body>
<div id="log"></div>
</body>
</html>
`))
