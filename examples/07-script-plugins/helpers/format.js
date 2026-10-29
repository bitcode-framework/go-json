// format function for the go-json script-plugins example.
//
// Same contract as predict.js: go-json injects `params` with params.args
// (the arguments for this call) and params.__function (the function to run).
function format(prediction) {
  var label = prediction.label;
  var score = prediction.score;
  return label.toUpperCase() + " (score " + score + ")";
}

var __result = (function () {
  var name = params.__function;
  var args = params.args || [];
  if (name === "format") {
    return format(args[0]);
  }
  throw new Error("unknown function: " + name);
})();

__result;
