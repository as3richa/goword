(function() {
  "use strict";

  /** @const */ var aspectRatio = 4 / 5;

  var gameElement;

  function initialize() {
    gameElement = document.getElementById("game");
  }

  function resizeGame() {
    var viewport = verge.viewport();
    var gameWidth, gameHeight;

    if(viewport.width / viewport.height < aspectRatio) {
      gameWidth = viewport.width;
      gameHeight = gameWidth / aspectRatio;
    } else {
      gameHeight = viewport.height;
      gameWidth = gameHeight * aspectRatio;
    }

    gameElement.style.width = gameWidth + "px";
    gameElement.style.height = gameHeight + "px";
  }

  window.addEventListener("load", initialize);
  window.addEventListener("load", resizeGame);
  window.addEventListener("resize", resizeGame);
})();
