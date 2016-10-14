(function() {
  "use strict";

  var socket;

  function sendJSON(object) {
    var string = JSON.stringify(object);
    if(socket.readyState === WebSocket.OPEN) {
      try {
        socket.send(string);
      } catch(err) {
        return;
      }
    } else {
      return;
    }
  }

  window.joinLobby = function(lobbyName) {
    sendJSON({"command": "join", "lobbyName": lobbyName});
  };

  window.partLobby = function() {
    sendJSON({"command": "part"});
  };

  window.ready = function() {
    sendJSON({"command": "ready"});
  };

  window.word = function(word) {
    sendJSON({"command": "word", "word": word});
  };

  window.addEventListener("load", function() {
    var proto = (window.location.protocol === "http:") ? "ws:" : "wss:";
    var path = proto + "//" + window.location.hostname + ":" + window.location.port + "/engine";

    socket = new WebSocket(path);

    socket.addEventListener("close", function() {
      updateInterface({type: "error", message: "Something went wrong. Please refresh the page."});
    });

    socket.addEventListener("error", function() {
      updateInterface({type: "error", message: "Something went wrong. Please refresh the page."});
    });

    socket.addEventListener("message", function(payload) {
      var data;
      try {
        data = JSON.parse(payload.data);
      } catch(e) {
        return;
      }
      console.log(data);
      updateInterface(data);
    });
  });
})();
