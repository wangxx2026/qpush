package cmd

import (
	"html/template"
	"net/http"
	"qpush/pkg/logger"
	"qpush/pkg/tail"

	"github.com/gorilla/websocket"
)

func logs(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/wslogs")
}

var upgrader = websocket.Upgrader{} // use default options

func wslogs(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("upgrade:", err)
		return
	}
	defer c.Close()

	tail.Push2WS(c, conf.QPTailFile, 5)

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
		alert("ws closed!")
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
