function initializeConsole(commands, defaultCommand) {
  "use strict";

  var helpDialog;
  commands.help = {
    arity: 0,
    help: ["", "displays the help dialog"],
    callback: function() {
      myConsole.print(helpDialog);
    }
  };

  (function() {
    var commandNames = Object.keys(commands);
    commandNames.sort();

    var helpPrefixes = {};
    var longestPrefixLength = 0;

    commandNames.forEach(function(name) {
      helpPrefixes[name] = "/" + name;
      if(commands[name].help[0] !== "") {
        helpPrefixes[name] += " " + commands[name].help[0];
      }

      if(helpPrefixes[name].length > longestPrefixLength) {
        longestPrefixLength = helpPrefixes[name].length;
      }
    });

    helpDialog = "< **Commands:**\n";
    commandNames.forEach(function(name) {
      helpDialog += "< ~ " + "**" + helpPrefixes[name] + "**" + " ".repeat(longestPrefixLength - helpPrefixes[name].length) + " - " + commands[name].help[1] + "\n";
    });
  })();

  var consoleElement = document.getElementById("console");
  var inputElement = document.getElementById("input");

  var myConsole = {};

  myConsole.print = function(string, color) {
    if(typeof color === "undefined") {
      color = null;
    }

    var scrolledDown = (consoleElement.scrollHeight * 0.95 - consoleElement.scrollTop <= consoleElement.clientHeight);

    var lines = string.split('\n');
    var frag = document.createDocumentFragment();

    lines.forEach(function(line) {
      if(line === "") {
        return;
      }

      var split = line.split("**");

      for(var i = 0; i < split.length; i ++) {
        if(split[i] !== "") {
          if(i % 2 === 0) {
            frag.appendChild(document.createTextNode(split[i]));
          } else {
            frag.appendChild(document.createElement("span"));
            frag.lastChild.style.fontWeight = "bold";
            frag.lastChild.appendChild(document.createTextNode(split[i]));
          }
        }
      }

      frag.appendChild(document.createElement("br"));
    });

    if(color !== null) {
      consoleElement.appendChild(document.createElement("span"));
      consoleElement.lastChild.style.color = color;
      consoleElement.lastChild.appendChild(frag);
    } else {
      consoleElement.appendChild(frag);
    }

    if(scrolledDown) {
      consoleElement.scrollTop = consoleElement.scrollHeight;
    }
  };

  function handleInput(string) {
    var parameters = parseCommand(string);
    var name, cmd, fn, arity;

    myConsole.print("> " + string);

    if(parameters[0] && parameters[0][0] === '/') {
      name = parameters[0].substring(1);
      parameters = parameters.slice(1);
    } else {
      name = defaultCommand;
    }

    cmd = commands[name];

    if(cmd) {
      fn = cmd.callback;
      arity = cmd.arity;
      if(arity === parameters.length) {
        fn.apply(myConsole, parameters, 1);
      } else {
        myConsole.print("! " + "**" + name + "**" + " takes exactly " + arity + " parameter(s)", "#e00");
      }
    } else {
      myConsole.print("! Unknown command", "#e00");
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

  inputElement.addEventListener("keydown", function(event) {
    var string = inputElement.value;
    if(event.keyCode === 13 && string !== "") {
      event.preventDefault();
      inputElement.value = "";
      handleInput(string);
    }
  });

  window.myConsole = myConsole;
}
