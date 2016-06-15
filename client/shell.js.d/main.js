(function() {
  "use strict";

  var consoleElement;
  var inputElement;

  var socket;

  function initialize() {
    consoleElement = document.getElementById("console");
    inputElement = document.getElementById("input");

    var proto = (window.location.protocol === "http:") ? "ws:" : "wss:";
    var path = proto + "//" + window.location.hostname + ":" + window.location.port + "/engine";

    print("= Trying to connect to " + path + "...\n");

    socket = new WebSocket(path);

    socket.addEventListener("open", function() {
      print("= Successfully connected to " + path + ".\n");
    });

    socket.addEventListener("error", function() {
      print("= Network error.\n");
    });

    socket.addEventListener("close", function() {
      print("= Connection closed.\n");
    });

    socket.addEventListener("message", function(message) {
      print("< " + message.data + "\n");
    });

    inputElement.addEventListener("keydown", function(event) {
      var string = inputElement.value;
      if(event.keyCode === 13 && string !== "") {
        event.preventDefault();
        inputElement.value = "";
        handleInput(string);
      }
    });
  }

  function handleInput(string) {
    if(socket.readyState === WebSocket.OPEN) {
      try {
        socket.send(string);
        print("> " + string + "\n");
      } catch(err) {
        print("! " + string + "\n");
      }
    } else {
      print("= Not connected.\n");
    }
  }

  function print(string) {
    consoleElement.appendChild(document.createTextNode(string));
  }

  function parseCommand(string) {
    var tokens = [];

    function skipWhitespace(index) {
      while(index < string.length && /\s/.test(string[index])) {
        index ++;
      }
      return index;
    }

    function parseToken(index) {
      if(index >= string.length) {
        return;
      }

      var right = index;

      if(string[index] != "\"" && string[index] != "'") {
        while(right < string.length && !(/\s/.test(string[right]))) {
          right ++;
        }
        tokens.push(string.substring(index, right));
      } else {
        var escapeState = false;
        var token = "";

        while(++ right < string.length) {
          if(escapeState) {
            escapeState = false;
            token += string[right];
          } else if(string[right] === "\\") {
            escapeState = true;
          } else if(string[right] === string[index]) {
            right ++;
            break;
          } else {
            token += string[right];
          }
        }

        tokens.push(token);
      }

      return right;
    }

    for(var index = 0; index < string.length;) {
      index = skipWhitespace(index);
      index = parseToken(index);
    }

    return tokens;
  }

  window.addEventListener("load", initialize);
})();
