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

    inputElement.addEventListener("keydown", function(event) {
      var string = inputElement.value;
      if(event.keyCode === 13 && string !== "") {
        event.preventDefault();
        inputElement.value = "";
        handleInput(string);
      }
    });
  }

  function print(string) {
    consoleElement.appendChild(document.createTextNode(string));
  }

  function handleInput(string) {
    print("> " + string + "\n");
  }

  window.addEventListener("load", initialize);
})();
