// predict function for the go-json script-plugins example.
//
// go-json exposes a script import as a proxy with call(name, ...args). The
// runtime injects `params` (with params.args holding the positional arguments)
// and `params.__function` naming the function to run. Return a value from the
// function that matches __function; go-json uses that as the call result.
function predict(features) {
  var total = 0;
  for (var i = 0; i < features.length; i++) {
    total += features[i];
  }
  return { score: total, label: total >= 10 ? "high" : "low" };
}

// Dispatch on the injected function name.
var __result = (function () {
  var name = params.__function;
  var args = params.args || [];
  if (name === "predict") {
    return predict(args[0]);
  }
  throw new Error("unknown function: " + name);
})();

__result;
