(function(){
  "use strict";

  function marginBox(elem, left, top, width, height, margin) {
    var wrapper = document.createElement("div");
    wrapper.style.boxSizing = "border-box";
    wrapper.style.position = "absolute";
    wrapper.style.left = left;
    wrapper.style.top = top;
    wrapper.style.width = width;
    wrapper.style.height = height;
    wrapper.style.padding = margin;

    elem.style.boxSizing = "border-box";
    elem.style.width = "100%";
    elem.style.height = "100%";
    elem.style.margin = "0";
    wrapper.appendChild(elem);

    return wrapper;
  }

  var defaultBorder = "2px solid #777";
  var defaultMargin = "0.8%";

  var aspectRatio = 25 / 20;

  var flashDuration = 5000;

  var stateDescriptions = {
    "awaitingPlayers": "Waiting for more players...",
    "betweenGames": "Waiting for the next game",
    "countdown": "Game starting momentarily",
    "inGame": "Game in progress"
  };

  var containerElem;

  var nicknameElem;
  var memoElem;
  var stateContainerElem;
  var buttonContainerElem;
  var readyButtonElem;
  var leaveButtonElem;
  var wordlistContainerElem;
  var gridElem;
  var inputContainerElem;
  var inputElem;
  var submitButtonElem;

  var lobbyFormElem;
  var lobbyFormInputElem;
  var lobbyFormSubmitElem;

  var lobbyTimerElem;
  var lobbyStateElem;
  var lobbyControlsElem;
  var lobbyReadyButtonElem;
  var lobbyPartButtonElem;
  var lobbyPlayerListElem;

  var gridContext;

  var nickname;
  var lobby;
  var grid = [
    ["G", "T", "S", "S"],
    ["O", "U", "E", "T"],
    ["C", "X", "I", "M"],
    ["F", "E", "R", "N"]
  ];
  var words = [];
  var nextAsyncEvent = null;

  var flashReset = null;

  window.addEventListener("load", function() {
    containerElem = document.createElement("div");
    containerElem.style.boxSizing = "border-box";
    containerElem.style.position = "absolute";

    nicknameElem = document.createElement("div");
    nicknameElem.appendChild(document.createTextNode(""));
    nicknameElem.style.textAlign = "center";
    nicknameElem.style.fontWeight = "bold";
    containerElem.appendChild(marginBox(nicknameElem, "2%", "2.5%", "32%", "5%", defaultMargin));

    memoElem = document.createElement("div");
    memoElem.appendChild(document.createTextNode(""));
    memoElem.style.textAlign = "center";
    memoElem.style.fontWeight = "bold";
    containerElem.appendChild(marginBox(memoElem, "34%", "3.5%", "64%", "5%", defaultMargin));

    stateContainerElem = document.createElement("div");
    stateContainerElem.style.border = defaultBorder;
    stateContainerElem.style.overfloxX = "hidden";
    stateContainerElem.style.overflowY = "scroll";
    containerElem.appendChild(marginBox(stateContainerElem, "2%", "7.5%", "32%", "40%", defaultMargin));

    wordlistContainerElem = document.createElement("div");
    wordlistContainerElem.style.border = defaultBorder;
    wordlistContainerElem.style.overflowY = "scroll";
    wordlistContainerElem.style.padding = "5%";
    containerElem.appendChild(marginBox(wordlistContainerElem, "2%", "47.5%", "32%", "50%", defaultMargin));

    gridElem = document.createElement("canvas");
    containerElem.appendChild(marginBox(gridElem, "34%", "7.5%", "64%", "80%", defaultMargin));

    inputContainerElem = document.createElement("div");
    inputContainerElem.position = "relative";

    inputElem = document.createElement("input");
    inputElem.style.boxSize = "border-box";
    inputElem.type = "text";
    inputElem.style.width = "80%";
    inputElem.style.height = "100%";
    inputElem.style.margin = "0px";
    inputElem.style.border = defaultBorder;
    inputElem.style.outline = "none";
    inputElem.style.position = "absolute";
    inputElem.style.left = "0";
    inputElem.style.top = "0";
    inputElem.style.paddingLeft = "2%";
    inputElem.placeholder = "Word";
    inputElem.addEventListener("keydown", function(e) {
      if(e.keyCode === 13) {
        e.preventDefault();
        submitButtonElem.click();
      }
    });
    inputContainerElem.appendChild(inputElem);

    submitButtonElem = document.createElement("button");
    submitButtonElem.style.boxSizing = "border-box";
    submitButtonElem.innerHTML = "Submit";
    submitButtonElem.style.width = "20%";
    submitButtonElem.style.height = "100%";
    submitButtonElem.style.position = "absolute";
    submitButtonElem.style.right = "0";
    submitButtonElem.style.top = "0";
    submitButtonElem.addEventListener("click", function() {
      word(inputElem.value);
      inputElem.value = "";
    });
    inputContainerElem.appendChild(submitButtonElem);

    containerElem.appendChild(marginBox(inputContainerElem, "34%", "87.5%", "64%", "9%", defaultMargin));

    lobbyFormElem = document.createElement("div");
    lobbyFormElem.style.padding = "5%";
    lobbyFormElem.style.width = "100%";
    lobbyFormElem.style.height = "100%";

    lobbyFormInputElem = document.createElement("input");
    lobbyFormInputElem.style.width = "70%";
    lobbyFormInputElem.style.height = "20%";
    lobbyFormInputElem.style.paddingLeft = "2%";
    lobbyFormInputElem.type = "text";
    lobbyFormInputElem.placeholder = "Lobby Name";
    lobbyFormElem.appendChild(lobbyFormInputElem);
    lobbyFormInputElem.addEventListener("keydown", function(e) {
      if(e.keyCode === 13) {
        e.preventDefault();
        lobbyFormSubmitElem.click();
      }
    });

    lobbyFormSubmitElem = document.createElement("button");
    lobbyFormSubmitElem.innerHTML = "Join";
    lobbyFormSubmitElem.style.width = "30%";
    lobbyFormSubmitElem.style.height = "20%";
    lobbyFormSubmitElem.style.fontFamily = "Roboto";
    lobbyFormElem.appendChild(lobbyFormSubmitElem);
    lobbyFormSubmitElem.addEventListener("click", function() {
      joinLobby(lobbyFormInputElem.value);
      lobbyFormInputElem.value = "";
    });

    stateContainerElem.appendChild(lobbyFormElem);

    lobbyTimerElem = document.createElement("div");
    lobbyTimerElem.appendChild(document.createTextNode("0:00"));
    lobbyTimerElem.style.textAlign = "center";
    lobbyTimerElem.style.padding = "5%";
    lobbyTimerElem.style.paddingBottom = "0";
    stateContainerElem.appendChild(lobbyTimerElem);

    lobbyStateElem = document.createElement("div");
    lobbyStateElem.appendChild(document.createTextNode(""));
    lobbyStateElem.style.textAlign = "center";
    lobbyStateElem.style.paddingBottom = "0";
    stateContainerElem.appendChild(lobbyStateElem);

    lobbyControlsElem = document.createElement("div");
    lobbyControlsElem.style.padding = "5%";
    lobbyControlsElem.style.paddingBottom = "0";
    lobbyControlsElem.style.width = "100%";
    lobbyControlsElem.style.height = "20%";

    lobbyReadyButtonElem = document.createElement("button");
    lobbyReadyButtonElem.innerHTML = "Ready";
    lobbyReadyButtonElem.style.width = "50%";
    lobbyReadyButtonElem.style.height = "100%";
    lobbyReadyButtonElem.addEventListener("click", ready);
    lobbyControlsElem.appendChild(lobbyReadyButtonElem);

    lobbyPartButtonElem = document.createElement("button");
    lobbyPartButtonElem.innerHTML = "Leave";
    lobbyPartButtonElem.style.width = "50%";
    lobbyPartButtonElem.style.height = "100%";
    lobbyPartButtonElem.addEventListener("click", partLobby);
    lobbyControlsElem.appendChild(lobbyPartButtonElem);

    stateContainerElem.appendChild(lobbyControlsElem);

    lobbyPlayerListElem = document.createElement("div");
    lobbyPlayerListElem.style.padding = "5%";
    lobbyPlayerListElem.style.width = "100%";
    stateContainerElem.appendChild(lobbyPlayerListElem);

    document.body.appendChild(containerElem);

    gridContext = gridElem.getContext("2d");

    resizeGame();
    renderInterface();
    setInterval(updateTimer, 200);
  });

  function resizeGame() {
    var viewport = verge.viewport();

    var width = Math.min(viewport.width, viewport.height * aspectRatio);
    var height = (width / aspectRatio);

    containerElem.style.width = width + "px";
    containerElem.style.height = height + "px";
    containerElem.style.left = ((viewport.width - width) / 2) + "px";
    containerElem.style.top = ((viewport.height - height) / 2) + "px";

    nicknameElem.style.fontSize = (0.03 * height) + "px";
    memoElem.style.fontSize = (0.018 * height) + "px";

    inputElem.style.fontSize = (0.06 * height) + "px";
    submitButtonElem.style.fontSize = (0.03 * height) + "px";

    lobbyFormInputElem.style.fontSize = (0.02 * height) + "px";
    lobbyFormSubmitElem.style.fontSize = (0.02 * height) + "px";

    lobbyTimerElem.style.fontSize = (0.06 * height) + "px";
    lobbyStateElem.style.fontSize = (0.016 * height) + "px";
    lobbyReadyButtonElem.style.fontSize = (0.025 * height) + "px";
    lobbyPartButtonElem.style.fontSize = (0.025 * height) + "px";
    lobbyPlayerListElem.style.fontSize = (0.022 * height) + "px";

    wordlistContainerElem.style.fontSize = (0.018 * height) + "px";

    gridElem.width = gridElem.offsetWidth;
    gridElem.height = gridElem.offsetHeight;

    renderGrid();
  }

  function renderGrid() {
    gridElem.width = gridElem.width;

    var squareSize = gridElem.width / 4;
    var squarePadding = squareSize / 20;
    var squareRounding = squareSize / 10;
    var glyphSize = squareSize * 0.7;

    gridContext.strokeStyle = "#777";
    gridContext.lineWidth = 2;
    gridContext.fillStyle = "#000";
    gridContext.font = glyphSize + "px sans";
    gridContext.textAlign = "center";
    gridContext.textBaseline = "middle";

    for(var i = 0; i < 4; i ++) {
      for(var j = 0; j < 4; j ++) {
        drawRoundedRectangle(gridContext,
          j * squareSize + squarePadding, i * squareSize + squarePadding,
          squareSize - 2 * squarePadding, squareSize - 2 * squarePadding, squareRounding);

        gridContext.font = (grid[i][j].length == 1 ? glyphSize : (0.75) * glyphSize) + "px sans";
        gridContext.fillText(grid[i][j], (j + 0.5) * squareSize, (i + 0.5) * squareSize);
      }
    }
  }

  function renderInterface() {
    nicknameElem.removeChild(nicknameElem.firstChild);
    nicknameElem.appendChild(document.createTextNode(nickname));

    if(lobby && lobby.state === "inGame" && lobby.players[nickname].playing) {
      inputElem.disabled = submitButtonElem.disabled = false;
    } else {
      inputElem.disabled = submitButtonElem.disabled = true;
      inputElem.value = "";
    }

    if(lobby && lobby.state === "betweenGames") {
      lobbyReadyButtonElem.disabled = false;
    } else {
      lobbyReadyButtonElem.disabled = true;
    }

    if(lobby) {
      lobbyFormElem.style.display = "none";
      lobbyTimerElem.style.display = "block";
      lobbyStateElem.style.display = "block";
      lobbyControlsElem.style.display = "block";
      lobbyPlayerListElem.style.display = "block";
    } else {
      lobbyFormElem.style.display = "block";
      lobbyTimerElem.style.display = "none";
      lobbyStateElem.style.display = "none";
      lobbyControlsElem.style.display = "none";
      lobbyPlayerListElem.style.display = "none";
    }

    if(lobby) {
      while(lobbyPlayerListElem.hasChildNodes()) {
        lobbyPlayerListElem.removeChild(lobbyPlayerListElem.lastChild);
      }

      var playerNames = Object.keys(lobby.players);
      playerNames.sort();

      playerNames.forEach(function(playerName) {
        var player = lobby.players[playerName];

        lobbyPlayerListElem.appendChild(document.createElement("div"));
        lobbyPlayerListElem.lastChild.style.paddingBottom = "2.5%";
        lobbyPlayerListElem.lastChild.appendChild(document.createTextNode(playerName + " (" + player.score + ")"));
        if(playerName === nickname) {
          lobbyPlayerListElem.lastChild.style.fontWeight = "bold";
        }
      });

      lobbyStateElem.removeChild(lobbyStateElem.lastChild);
      lobbyStateElem.appendChild(document.createTextNode(stateDescriptions[lobby.state]));
    }

    renderGrid();
    renderWordlists();
    updateTimer();
  }

  function renderWordlists() {
    while(wordlistContainerElem.hasChildNodes()) {
      wordlistContainerElem.removeChild(wordlistContainerElem.lastChild);
    }

    if(words.length > 0) {
      var text = "";
      words.forEach(function(word) {
        if(text !== "") {
          text += " ";
        }
        text += word;
      });

      wordlistContainerElem.appendChild(document.createElement("p"));
      wordlistContainerElem.lastChild.style.fontWeight = "bold";
      wordlistContainerElem.lastChild.style.marginTop = "0";
      wordlistContainerElem.lastChild.appendChild(document.createTextNode(text));
    } else if(lobby) {
      var playerNames = Object.keys(lobby.players);
      playerNames.sort();
      playerNames.forEach(function(playerName) {
        var player = lobby.players[playerName];
        if(player.result) {
          wordlistContainerElem.appendChild(document.createElement("p"));
          wordlistContainerElem.lastChild.style.fontWeight = "bold";
          wordlistContainerElem.lastChild.style.marginTop = "0";
          wordlistContainerElem.lastChild.appendChild(document.createTextNode(playerName + "'s words"));
          wordlistContainerElem.appendChild(document.createElement("p"));
          wordlistContainerElem.lastChild.style.marginTop = "0";

          player.result.words.forEach(function(word) {
            var wordElem = document.createElement("span");
            wordElem.appendChild(document.createTextNode(word.word));
            if(word.points === -1) {
              wordElem.style.color = "red";
              wordElem.appendChild(document.createTextNode(" (-1)"));
            } else if(word.points === 0) {
              wordElem.style.textDecoration = "line-through";
            } else {
              wordElem.style.fontWeight = "bold";
              wordElem.appendChild(document.createTextNode(" (" + word.points + ")"));
            }
            if(wordlistContainerElem.lastChild.children.length > 0) {
              wordlistContainerElem.lastChild.appendChild(document.createTextNode(" "));
            }
            wordlistContainerElem.lastChild.appendChild(wordElem);
          });
        }
      });
    }
  }

  function flashMessage(message, err) {
    memoElem.removeChild(memoElem.firstChild);

    var messageNode = document.createTextNode(message);
    if(err) {
      memoElem.appendChild(document.createElement("span"));
      memoElem.lastChild.style.color = "#f00";
      memoElem.lastChild.appendChild(messageNode);
    } else {
      memoElem.appendChild(messageNode);
    }

    if(flashReset) {
      clearTimeout(flashReset);
    }
    flashReset = setTimeout(clearMessage, flashDuration);
  }

  function clearMessage() {
    memoElem.removeChild(memoElem.firstChild);
    memoElem.appendChild(document.createTextNode(lobby ? "Lobby " + lobby.name : "Goword"));
    flashReset = null;
  }

  function updateTimer() {
    var stamp;
    if(!nextAsyncEvent || Date.now() > nextAsyncEvent) {
      stamp = "0:00";
    } else {
      var delta = nextAsyncEvent - Date.now();
      delta = Math.ceil(delta / 1000);
      stamp = "";
      stamp += Math.floor(delta / 60);
      stamp += ":";
      delta %= 60;
      stamp += Math.floor(delta / 10);
      delta %= 10;
      stamp += Math.floor(delta);
    }
    lobbyTimerElem.removeChild(lobbyTimerElem.firstChild);
    lobbyTimerElem.appendChild(document.createTextNode(stamp));
  }

  window.addEventListener("resize", resizeGame);

  window.updateInterface = function(data) {
    if(data.type === "state") {
      nickname = data.nickname;
      lobby = data.lobby;
      if(data.lobby && data.lobby.grid) {
        grid = data.lobby.grid;
      }
      nextAsyncEvent = (lobby && lobby.secondsRemaining) ? (Date.now() + lobby.secondsRemaining * 1000) : null;

      if(!lobby || lobby.state !== "inGame") {
        words = [];
      }
    } else if(data.type === "word") {
      words.push(data.word);
    }

    if(data.message) {
      flashMessage(data.message, data.type === "error");
    }

    renderInterface();
  };
})();
