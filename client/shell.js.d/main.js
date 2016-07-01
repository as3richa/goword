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

    print("= Trying to connect to " + path + "...");

    socket = new WebSocket(path);

    socket.addEventListener("open", function() {
      print("= Successfully connected to " + path + ".");
    });

    socket.addEventListener("error", function() {
      print("= Network error.");
    });

    socket.addEventListener("close", function() {
      print("= Connection closed.");
    });

    socket.addEventListener("message", function(message) {
      print("< " + message.data + "");
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

  var commands = {
    "join": function() {
      if(arguments.length !== 3) {
        print("< /join takes exactly 3 arguments");
        return;
      }

      sendJSON({
        "command": "join",
        "name": arguments[0],
        "password": arguments[1],
        "nickname": arguments[2]
      });
    },
    "part": function() {
      if(arguments.length !== 0) {
        print("< /part takes no arguments");
        return;
      }

      sendJSON({"command": "part"});
    },
    "help": function() {
      print(
        "< Commands:\n" +
        "<  /join <lobby name> <password> <nickname> - attempts to join a lobby\n" +
        "<  /part                                    - attempts to leave the lobby\n" +
        "<  /help                                    - shows this dialog"
      );
    }
  };

  function handleInput(string) {
    var parameters = parseCommand(string);
    var fn;

    print("> " + string);
    if(parameters[0] && parameters[0][0] === '/' && (fn = commands[parameters[0].substring(1)])) {
      fn.apply(null, parameters.slice(1), 1);
    } else {
      print("! Unknown command.");
    }
  }

  function sendJSON(object) {
    var string = JSON.stringify(object);
    if(socket.readyState === WebSocket.OPEN) {
      try {
        socket.send(string);
        print("> " + string);
      } catch(err) {
        print("! " + string);
      }
    } else {
      print("= Not connected.");
    }
  }

  function print(string) {
    var scrolledDown = (consoleElement.scrollHeight * 0.95 - consoleElement.scrollTop <= consoleElement.clientHeight);
    var lines = string.split('\n');
    var frag = document.createDocumentFragment();

    lines.forEach(function(line) {
      if(line === "") {
        return;
      }
      frag.appendChild(document.createTextNode(line));
      frag.appendChild(document.createElement("br"));
    });

    consoleElement.appendChild(frag);

    if(scrolledDown) {
      consoleElement.scrollTop = consoleElement.scrollHeight;
    }
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

      if(string[index] != '"' && string[index] != '\'') {
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
