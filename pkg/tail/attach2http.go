package tail

import (
	"html/template"
	"net/http"
	"qpush/pkg/logger"

	"github.com/gorilla/websocket"
)

// Attach2Http will attach endpoints to ServeMux
func Attach2Http(mux *http.ServeMux, httpAddr string, wsAddr string, file string) {
	mux.HandleFunc(httpAddr, func(w http.ResponseWriter, r *http.Request) {
		httpHandler(w, r, wsAddr)
	})
	mux.HandleFunc(wsAddr, func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, file)
	})
}

var upgrader = websocket.Upgrader{} // use default options

func wsHandler(w http.ResponseWriter, r *http.Request, file string) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("upgrade:", err)
		return
	}

	// Push2WS will Close c when return
	Push2WS(c, file, 5)
}

func httpHandler(w http.ResponseWriter, r *http.Request, wsAddr string) {
	homeTemplate.Execute(w, "ws://"+r.Host+wsAddr)
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
