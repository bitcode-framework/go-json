/* go-json playground — loads the WASM build of the language engine and runs
 * programs entirely in the browser. No server, no network calls. */
(() => {
  'use strict';

  const sourceEl = document.getElementById('source');
  const outputEl = document.getElementById('output');
  const samplesEl = document.getElementById('samples');
  const runBtn = document.getElementById('run');
  const clearBtn = document.getElementById('clear');
  const engineEl = document.getElementById('engine');

  const SAMPLES = {
    'Hello World': `{
  "name": "hello",
  "go_json": "1",
  "steps": [
    {"let": "greeting", "value": "Hello, World!"},
    {"return": "greeting"}
  ]
}`,

    'Factorial (recursion)': `{
  "name": "math_demo",
  "go_json": "1",
  "functions": {
    "factorial": {
      "params": {"n": "int"},
      "returns": "int",
      "steps": [
        {"if": "n <= 1", "then": [{"return": 1}]},
        {"let": "sub", "call": "factorial", "with": {"n": "n - 1"}},
        {"return": "n * sub"}
      ]
    }
  },
  "steps": [
    {"let": "result", "call": "factorial", "with": {"n": "10"}},
    {"return": "result"}
  ]
}`,

    'JSONC with comments': `{
  // JSONC allows comments and trailing commas
  "name": "discount",
  "go_json": "1",
  "functions": {
    "price": {
      "params": {"amount": "float", "tier": "string",},
      "steps": [
        /* Gold gets 20% off */
        {"if": "tier == 'gold'", "then": [{"return": "amount * 0.8"}]},
        {"if": "tier == 'silver'", "then": [{"return": "amount * 0.9"}]},
        {"return": "amount"}
      ]
    }
  },
  "steps": [
    {"let": "gold", "call": "price", "with": {"amount": "100", "tier": "'gold'"}},
    {"let": "silver", "call": "price", "with": {"amount": "100", "tier": "'silver'"}},
    {"let": "summary", "expr": "{gold: gold, silver: silver}"},
    {"return": "summary"}
  ]
}`,

    'Loops and lambdas': `{
  "name": "lambda_demo",
  "go_json": "1",
  "steps": [
    {"let": "nums", "value": [1, 2, 3, 4, 5]},
    {"let": "doubled", "expr": "mapFn(nums, fn(x) => x * 2)"},
    {"let": "evens", "expr": "filterFn(nums, fn(x) => int(x) % 2 == 0)"},
    {"let": "total", "expr": "reduceFn(nums, fn(acc, x) => acc + x, 0)"},
    {"let": "out", "expr": "{doubled: doubled, evens: evens, total: total}"},
    {"return": "out"}
  ]
}`,

    'Structs and methods': `{
  "name": "struct_demo",
  "go_json": "1",
  "structs": {
    "Counter": {
      "fields": {"value": "int"},
      "methods": {
        "increment": {
          "params": {"by": "int"},
          "steps": [
            {"set": "self.value", "expr": "self.value + by"},
            {"return": "self.value"}
          ]
        }
      }
    }
  },
  "steps": [
    {"let": "c", "new": "Counter", "with": {"value": 0}},
    {"let": "a", "call": "c.increment", "with": {"by": "5"}},
    {"let": "b", "call": "c.increment", "with": {"by": "3"}},
    {"return": "b"}
  ]
}`,

    'Pattern matching': `{
  "name": "match_demo",
  "go_json": "1",
  "steps": [
    {"let": "label", "value": "unset"},
    {"let": "resp", "value": {"status": 404, "body": "missing"}},
    {"match": "resp", "cases": [
      {"pattern": {"status": 200}, "then": [{"set": "label", "value": "success"}]},
      {"pattern": {"status": 404}, "then": [{"set": "label", "value": "not found"}]},
      {"pattern": "_", "then": [{"set": "label", "value": "error"}]}
    ]},
    {"return": "label"}
  ]
}`,

    'Higher-order functions': `{
  "name": "stdlib_demo",
  "go_json": "1",
  "steps": [
    {"let": "words", "value": ["alpha", "beta", "gamma"]},
    {"let": "loud", "expr": "mapFn(words, fn(w) => upper(w))"},
    {"let": "joined", "expr": "join(loud, ' / ')"},
    {"let": "squared", "expr": "mapFn([1, 2, 3], fn(x) => x * x)"},
    {"let": "out", "expr": "{joined: joined, squared: squared}"},
    {"return": "out"}
  ]
}`,

    'Error handling': `{
  "name": "error_demo",
  "go_json": "1",
  "steps": [
    {"try": [
      {"error": "'something went wrong'"}
    ], "catch": {
      "as": "err",
      "steps": [
        {"let": "msg", "expr": "'caught: ' + err.message"},
        {"return": "msg"}
      ]
    }}
  ]
}`
  };

  for (const name of Object.keys(SAMPLES)) {
    const opt = document.createElement('option');
    opt.value = name;
    opt.textContent = name;
    samplesEl.appendChild(opt);
  }

  samplesEl.addEventListener('change', () => {
    const name = samplesEl.value;
    if (name && SAMPLES[name]) {
      sourceEl.value = SAMPLES[name];
      outputEl.innerHTML = '<span class="empty">Press Run to execute the program.</span>';
    }
  });

  // Once the user edits the source, the loaded example no longer matches, so
  // clear the selection instead of showing a stale example name.
  sourceEl.addEventListener('input', () => {
    const name = samplesEl.value;
    if (name && SAMPLES[name] && sourceEl.value !== SAMPLES[name]) {
      samplesEl.value = '';
    }
  });

  function render(text, cls) {
    outputEl.textContent = '';
    const span = document.createElement('span');
    if (cls) span.className = cls;
    span.textContent = text;
    outputEl.appendChild(span);
  }

  function renderResult(raw) {
    let parsed;
    try {
      parsed = JSON.parse(raw);
    } catch {
      render(raw, 'err');
      return;
    }

    if (!parsed.ok) {
      render(parsed.error, 'err');
      return;
    }

    outputEl.textContent = '';

    const value = document.createElement('span');
    value.className = 'ok';
    value.textContent = typeof parsed.value === 'string'
      ? JSON.stringify(parsed.value)
      : JSON.stringify(parsed.value, null, 2);
    outputEl.appendChild(value);

    if (parsed.steps !== undefined) {
      const meta = document.createElement('div');
      meta.className = 'meta';
      const bits = [];
      if (parsed.steps) bits.push(parsed.steps + ' steps');
      if (parsed.duration) bits.push(parsed.duration);
      meta.textContent = bits.join(' · ');
      outputEl.appendChild(meta);
    }
  }

  function execute() {
    const src = sourceEl.value.trim();
    if (!src) {
      render('Nothing to run — the program is empty.', 'err');
      return;
    }
    if (typeof window.goJsonRun !== 'function') {
      render('The engine is still loading. Give it a moment and try again.', 'err');
      return;
    }

    runBtn.disabled = true;
    try {
      renderResult(window.goJsonRun(src));
    } catch (err) {
      render('Playground error: ' + (err && err.message ? err.message : String(err)), 'err');
    } finally {
      runBtn.disabled = false;
    }
  }

  runBtn.addEventListener('click', execute);
  clearBtn.addEventListener('click', () => {
    outputEl.innerHTML = '<span class="empty">Press Run to execute the program.</span>';
  });

  sourceEl.addEventListener('keydown', (e) => {
    // Ctrl/Cmd+Enter runs; Tab inserts two spaces instead of moving focus.
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      execute();
      return;
    }
    if (e.key === 'Tab') {
      e.preventDefault();
      const { selectionStart: start, selectionEnd: end } = sourceEl;
      sourceEl.value = sourceEl.value.slice(0, start) + '  ' + sourceEl.value.slice(end);
      sourceEl.selectionStart = sourceEl.selectionEnd = start + 2;
    }
  });

  sourceEl.value = SAMPLES['Hello World'];
  samplesEl.value = 'Hello World';

  // Boot the WASM engine.
  (async () => {
    try {
      if (typeof Go !== 'function') {
        throw new Error('wasm_exec.js did not load');
      }
      const go = new Go();
      const result = await WebAssembly.instantiateStreaming
        ? await WebAssembly.instantiateStreaming(fetch('go-json.wasm'), go.importObject)
        : await WebAssembly.instantiate(await (await fetch('go-json.wasm')).arrayBuffer(), go.importObject);

      go.run(result.instance);
      // go.run blocks in a goroutine; poll briefly for the registered hook.
      for (let i = 0; i < 100; i++) {
        if (typeof window.goJsonRun === 'function') {
          engineEl.textContent = 'engine ready · runs locally in your browser';
          return;
        }
        await new Promise((r) => setTimeout(r, 50));
      }
      throw new Error('engine did not register its entry point');
    } catch (err) {
      engineEl.textContent = 'engine failed to load';
      render('Failed to load the WebAssembly engine: ' + (err && err.message ? err.message : String(err)), 'err');
    }
  })();
})();
