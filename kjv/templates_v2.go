package kjv

import "html/template"

const v2PageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Bible API v2</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@300;400;600;700&family=Roboto+Mono:wght@400;600&display=swap" rel="stylesheet">
  <script src="https://accounts.google.com/gsi/client" async defer></script>
  <style>
    :root {
      --bg: #f4f0e8;
      --ink: #1b1b1f;
      --accent: #d67b2e;
      --accent-dark: #7a3d0b;
      --panel: #ffffff;
      --muted: #6a6a72;
      --border: #e0d7c8;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "IBM Plex Sans", sans-serif;
      color: var(--ink);
      background: radial-gradient(circle at 20% 20%, #fff8ee, #f4f0e8 40%, #efe5d3 100%);
      min-height: 100vh;
    }
    header {
      padding: 32px 48px 16px;
      display: flex;
      flex-wrap: wrap;
      gap: 16px;
      align-items: center;
      justify-content: space-between;
    }
    header h1 {
      margin: 0;
      font-size: 32px;
      letter-spacing: -0.02em;
    }
    header p {
      margin: 4px 0 0;
      color: var(--muted);
    }
    main {
      display: grid;
      grid-template-columns: minmax(260px, 360px) 1fr;
      gap: 24px;
      padding: 16px 48px 48px;
    }
    .panel {
      background: var(--panel);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 20px;
      box-shadow: 0 12px 30px rgba(0,0,0,0.06);
    }
    .panel h2 {
      margin-top: 0;
      font-size: 18px;
      text-transform: uppercase;
      letter-spacing: 0.1em;
      color: var(--accent-dark);
    }
    .token {
      font-family: "Roboto Mono", monospace;
      font-size: 12px;
      background: #fbf6ef;
      border: 1px dashed var(--border);
      padding: 12px;
      border-radius: 10px;
      word-break: break-all;
    }
    button {
      border: none;
      background: var(--accent);
      color: white;
      padding: 10px 14px;
      border-radius: 10px;
      cursor: pointer;
      font-weight: 600;
    }
    button.secondary {
      background: transparent;
      color: var(--accent-dark);
      border: 1px solid var(--accent-dark);
    }
    .field {
      display: flex;
      gap: 8px;
      align-items: center;
      margin-bottom: 10px;
    }
    input[type="text"] {
      width: 100%;
      padding: 10px 12px;
      border-radius: 10px;
      border: 1px solid var(--border);
      font-size: 14px;
    }
    .results {
      margin-top: 12px;
      display: grid;
      gap: 10px;
    }
    .result {
      padding: 12px;
      border-radius: 12px;
      border: 1px solid var(--border);
      background: #fffdf9;
    }
    .result .meta {
      font-size: 12px;
      color: var(--muted);
      margin-bottom: 6px;
    }
    .highlight {
      background: #ffe2c4;
      padding: 0 4px;
      border-radius: 4px;
    }
    .suggestions {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-top: 8px;
    }
    .chip {
      padding: 6px 10px;
      border-radius: 999px;
      background: #fff0dc;
      border: 1px solid var(--border);
      font-size: 12px;
      cursor: pointer;
    }
    @media (max-width: 900px) {
      main {
        grid-template-columns: 1fr;
        padding: 16px 20px 40px;
      }
      header {
        padding: 24px 20px 8px;
      }
    }
  </style>
</head>
<body>
  <header>
    <div>
      <h1>Bible API v2</h1>
      <p>OpenSearch‑powered search + predictive suggestions</p>
    </div>
    <div id="google-button"></div>
  </header>

  <main>
    <section class="panel">
      <h2>Session</h2>
      <p id="session-status">Not signed in.</p>
      <div class="token" id="token-display">Token will appear here after login.</div>
      <div class="field">
        <button id="logout-btn" class="secondary">Logout</button>
      </div>
    </section>

    <section class="panel">
      <h2>Search</h2>
      <div class="field">
        <input id="search-input" type="text" placeholder="Search the text (e.g., grace, faith, covenant)">
        <button id="search-btn">Search</button>
      </div>
      <div class="field">
        <input id="suggest-input" type="text" placeholder="Typeahead suggestion">
      </div>
      <div class="suggestions" id="suggestions"></div>
      <div class="results" id="results"></div>
    </section>
  </main>

  <script>
    const tokenDisplay = document.getElementById("token-display");
    const sessionStatus = document.getElementById("session-status");
    const resultsEl = document.getElementById("results");
    const suggestionsEl = document.getElementById("suggestions");
    const searchInput = document.getElementById("search-input");
    const suggestInput = document.getElementById("suggest-input");

    let activeToken = "";

    function setToken(token) {
      activeToken = token || "";
      if (activeToken) {
        tokenDisplay.textContent = activeToken;
        sessionStatus.textContent = "Signed in. Token ready for protected calls.";
      } else {
        tokenDisplay.textContent = "Token will appear here after login.";
        sessionStatus.textContent = "Not signed in.";
      }
    }

    async function fetchSession() {
      try {
        const resp = await fetch("/auth/me");
        if (!resp.ok) {
          setToken("");
          return;
        }
        const data = await resp.json();
        setToken(data.token || "");
      } catch (err) {
        setToken("");
      }
    }

    function renderResults(items) {
      resultsEl.innerHTML = "";
      if (!items || items.length === 0) {
        resultsEl.innerHTML = "<div class='result'>No results yet.</div>";
        return;
      }
      items.forEach(item => {
        const div = document.createElement("div");
        div.className = "result";
        const meta = document.createElement("div");
        meta.className = "meta";
        meta.textContent = item.book + " " + item.chapter + ":" + item.verse;
        const text = document.createElement("div");
        if (item.highlight && item.highlight.length) {
          text.innerHTML = item.highlight[0].replace(/<em>/g, "<span class='highlight'>").replace(/<\/em>/g, "</span>");
        } else {
          text.textContent = item.text;
        }
        div.appendChild(meta);
        div.appendChild(text);
        resultsEl.appendChild(div);
      });
    }

    async function runSearch() {
      const q = searchInput.value.trim();
      if (!q) return;
      resultsEl.innerHTML = "<div class='result'>Searching...</div>";
      const resp = await fetch("/bible/v2/search?q=" + encodeURIComponent(q));
      const data = await resp.json();
      renderResults(data.data && data.data.results ? data.data.results : []);
    }

    let suggestTimer;
    async function runSuggest(value) {
      if (!value) {
        suggestionsEl.innerHTML = "";
        return;
      }
      const resp = await fetch("/bible/v2/suggest?q=" + encodeURIComponent(value));
      if (!resp.ok) return;
      const data = await resp.json();
      const suggestions = (data.data && data.data.suggestions) || [];
      suggestionsEl.innerHTML = "";
      suggestions.slice(0, 12).forEach(s => {
        const chip = document.createElement("div");
        chip.className = "chip";
        chip.textContent = s.text;
        chip.addEventListener("click", () => {
          searchInput.value = s.text;
          runSearch();
        });
        suggestionsEl.appendChild(chip);
      });
    }

    document.getElementById("search-btn").addEventListener("click", runSearch);
    searchInput.addEventListener("keydown", (e) => {
      if (e.key === "Enter") runSearch();
    });
    suggestInput.addEventListener("input", (e) => {
      clearTimeout(suggestTimer);
      suggestTimer = setTimeout(() => runSuggest(e.target.value.trim()), 200);
    });
    document.getElementById("logout-btn").addEventListener("click", async () => {
      await fetch("/auth/logout", { method: "POST" });
      setToken("");
    });

    window.handleCredentialResponse = async function(response) {
      const resp = await fetch("/auth/google/token", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id_token: response.credential, scope: "synonyms:write" })
      });
      if (!resp.ok) {
        alert("Login failed.");
        return;
      }
      const data = await resp.json();
      setToken(data.token || "");
    };

    window.onload = () => {
      fetchSession();
      google.accounts.id.initialize({
        client_id: "{{.GoogleClientID}}",
        callback: handleCredentialResponse
      });
      google.accounts.id.renderButton(
        document.getElementById("google-button"),
        { theme: "outline", size: "large" }
      );
    };
  </script>
</body>
</html>`

func (app *App) v2Template() (*template.Template, error) {
	return template.New("v2").Parse(v2PageTemplate)
}
