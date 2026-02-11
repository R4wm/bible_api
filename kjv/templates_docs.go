package kjv

import "html/template"

const docsPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Bible API Docs</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;600&family=Source+Code+Pro:wght@400;600&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #f8f5f1;
      --ink: #1d1b1a;
      --accent: #2f6b58;
      --panel: #ffffff;
      --border: #e0d7c8;
      --muted: #6d6a68;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Space Grotesk", sans-serif;
      color: var(--ink);
      background: linear-gradient(135deg, #f8f5f1 0%, #efe7db 100%);
      min-height: 100vh;
    }
    header {
      padding: 28px 40px 12px;
      display: flex;
      justify-content: space-between;
      align-items: flex-end;
      flex-wrap: wrap;
      gap: 10px;
    }
    header h1 {
      margin: 0;
      font-size: 28px;
    }
    header a {
      color: var(--accent);
      text-decoration: none;
      font-weight: 600;
    }
    main {
      padding: 12px 40px 40px;
    }
    .card {
      background: var(--panel);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 16px;
      margin-bottom: 12px;
      display: grid;
      grid-template-columns: 90px 1fr;
      gap: 12px;
      align-items: center;
    }
    .method {
      font-family: "Source Code Pro", monospace;
      font-size: 12px;
      background: #e8f1ed;
      color: #205142;
      padding: 6px 8px;
      border-radius: 8px;
      text-align: center;
      font-weight: 600;
    }
    .path {
      font-family: "Source Code Pro", monospace;
      font-size: 14px;
    }
    .desc {
      color: var(--muted);
      margin-top: 4px;
      font-size: 13px;
    }
    .auth {
      margin-top: 6px;
      color: #8a4b12;
      font-size: 12px;
    }
    @media (max-width: 720px) {
      header, main { padding: 20px; }
      .card { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <header>
    <div>
      <h1>Bible API Docs</h1>
      <div class="desc">Quick reference for available endpoints</div>
    </div>
    <a href="/docs.json">View JSON</a>
  </header>
  <main>
    {{range .}}
    <div class="card">
      <div class="method">{{.Method}}</div>
      <div>
        <div class="path">{{.Path}}</div>
        <div class="desc">{{.Description}}</div>
        {{if .Auth}}<div class="auth">Auth: {{.Auth}}</div>{{end}}
      </div>
    </div>
    {{end}}
  </main>
</body>
</html>`

func (app *App) docsTemplate() (*template.Template, error) {
	return template.New("docs").Parse(docsPageTemplate)
}
