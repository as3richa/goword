(function() {
  "use strict";

  var socket;
  var grid = null;

  function sendJSON(object) {
    var string = JSON.stringify(object);
    if(socket.readyState === WebSocket.OPEN) {
      try {
        socket.send(string);
      } catch(err) {
        myConsole.print("! " + string);
      }
    } else {
      myConsole.print("= Not connected.");
    }
  }

  var commands = {};

  commands.join = {
    arity: 1,
    help: ["<lobby name>", "attempts to join a lobby"],
    callback: function(lobbyName) {
      sendJSON({
        "command": "join",
        "lobbyName": lobbyName
      });
    }
  };

  commands.part = {
    arity: 0,
    help: ["", "attempts to leave the currently-joined lobby"],
    callback: function() {
      sendJSON({"command": "part"});
    }
  };

  commands.ready = {
    arity: 0,
    help: ["", "indicates that you're ready to play again"],
    callback: function() {
      sendJSON({"command": "ready"});
    }
  };

  commands.guess = {
    arity: 1,
    help: ["word", "guesses a word in-game"],
    callback: function(word) {
      sendJSON({"command": "guess", "word": word});
    }
  };

  window.addEventListener("load", function() {
    initializeConsole(commands, "guess");

    var proto = (window.location.protocol === "http:") ? "ws:" : "wss:";
    var path = proto + "//" + window.location.hostname + ":" + window.location.port + "/engine";

    myConsole.print("= Connecting...");

    socket = new WebSocket(path);

    socket.addEventListener("error", function() {
      myConsole.print("= Network error.");
    });

    socket.addEventListener("close", function() {
      myConsole.print("= Connection closed.");
    });

    socket.addEventListener("message", function(message) {
      var data;
      try {
        data = JSON.parse(message.data);
      } catch(e) {
        return;
      }

      if(data.ok) {
        switch(data.command) {
          case "nick":
            myConsole.print("< " + data.message, "#080");
            break;

          case "join":
          case "part":
            myConsole.print("< " + data.message, "#080");
            myConsole.print("< **Players:**");
            if(data.metadata.players) {
              for(var player in data.metadata.players) {
                myConsole.print("< ~ " + player);
              }
            }
            break;

          case "state":
            var description = (data.message) ? data.message + ". ": "";
            switch(data.metadata.state) {
              case "awaitingPlayers":
                description += "Waiting for more players...";
                break;

              case "betweenGames":
                description += "Waiting on the next game; " + data.metadata.remaining + "s left - **/ready** to start early";
                break;

              case "countdown":
                description += "Next game starts in " + data.metadata.remaining + "s";
                break;

              case "inGame":
                description += "Round in progress; " + data.metadata.remaining + "s left";
                break;
            }
            myConsole.print("< " + description, "#080");

            grid = data.metadata.grid;

            if(grid) {
              for(var i = 0; i < 4; i ++) {
                var line = "< **";
                for(var j = 0; j < 4; j ++) {
                  if(j > 0) {
                    line += " ";
                  }

                  line += "[";
                  if(grid[i][j].length == 1) {
                    line += " ";
                  }
                  line += grid[i][j];
                  line += "]";
                }
                line += "**";

                myConsole.print(line);
              }
            }
            break;

          case "result":
            console.log(data.metadata);
            myConsole.print("< **Results:**");
            for(var playerName in data.metadata) {
              var result = data.metadata[playerName];
              myConsole.print("< **" + playerName + "** made " + result.words.length + " guesses for a total of " + result.total + " point(s)");

              var listing = "<";
              for(var k = 0; k < result.words.length; k ++) {
                var pair = result.words[k];
                listing += " ";
                if(pair[1] === 0) {
                  listing += pair[0];
                } else {
                  listing += "**" + pair[0] + " (" + pair[1] + ")**";
                }
              }

              myConsole.print(listing);
            }
            break;

          case "ready":
            myConsole.print("< " + data.message, "#080");
            break;
        }
      } else {
        myConsole.print("< " + data.message, "#e00");
      }
    });
  });
})();
