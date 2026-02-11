import { useEffect, useMemo, useState } from "react";

const defaultScope = "synonyms:write";

type SearchResult = {
  book: string;
  chapter: number;
  verse: number;
  text: string;
  highlight?: string[];
};

type Suggestion = {
  text: string;
  score: number;
};

type AuthConfig = {
  google_client_id?: string;
};

export default function App() {
  const [token, setToken] = useState<string>("");
  const [status, setStatus] = useState<string>("Checking session...");
  const [query, setQuery] = useState<string>("");
  const [suggestQuery, setSuggestQuery] = useState<string>("");
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");
  const [config, setConfig] = useState<AuthConfig>({});

  const hasToken = token.length > 0;

  useEffect(() => {
    fetch("/auth/config")
      .then((res) => res.json())
      .then((data) => setConfig(data))
      .catch(() => setConfig({}));
  }, []);

  useEffect(() => {
    fetch("/auth/me")
      .then((res) => {
        if (!res.ok) {
          setStatus("Not signed in.");
          return null;
        }
        return res.json();
      })
      .then((data) => {
        if (data?.token) {
          setToken(data.token);
          setStatus("Signed in. Token ready.");
        }
      })
      .catch(() => setStatus("Not signed in."));
  }, []);

  useEffect(() => {
    if (!config.google_client_id) {
      return;
    }

    let cancelled = false;
    const initGoogle = () => {
      if (cancelled) {
        return;
      }
      const google = (window as any).google;
      if (!google?.accounts?.id) {
        window.setTimeout(initGoogle, 300);
        return;
      }
      google.accounts.id.initialize({
        client_id: config.google_client_id,
        callback: async (response: { credential: string }) => {
          const resp = await fetch("/auth/google/token", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              id_token: response.credential,
              scope: defaultScope
            })
          });
          if (!resp.ok) {
            setStatus("Login failed.");
            return;
          }
          const data = await resp.json();
          setToken(data.token || "");
          setStatus("Signed in. Token ready.");
        }
      });
      google.accounts.id.renderButton(document.getElementById("google-button"), {
        theme: "outline",
        size: "large"
      });
    };

    initGoogle();
    return () => {
      cancelled = true;
    };
  }, [config.google_client_id]);

  const runSearch = async () => {
    if (!query.trim()) {
      return;
    }
    setLoading(true);
    setError("");
    try {
      const resp = await fetch(`/bible/v2/search?q=${encodeURIComponent(query)}`);
      if (!resp.ok) {
        throw new Error("Search failed");
      }
      const data = await resp.json();
      setResults(data?.data?.results || []);
    } catch (err) {
      setError("Search failed. Check the API.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!suggestQuery.trim()) {
      setSuggestions([]);
      return;
    }

    const timer = window.setTimeout(async () => {
      try {
        const resp = await fetch(`/bible/v2/suggest?q=${encodeURIComponent(suggestQuery)}`);
        if (!resp.ok) {
          return;
        }
        const data = await resp.json();
        setSuggestions(data?.data?.suggestions || []);
      } catch {
        setSuggestions([]);
      }
    }, 200);

    return () => window.clearTimeout(timer);
  }, [suggestQuery]);

  const logout = async () => {
    await fetch("/auth/logout", { method: "POST" });
    setToken("");
    setStatus("Not signed in.");
  };

  const safeHighlight = useMemo(() => {
    return (item: SearchResult) => {
      if (!item.highlight || item.highlight.length === 0) {
        return item.text;
      }
      return item.highlight[0]
        .replace(/<em>/g, "<mark>")
        .replace(/<\/em>/g, "</mark>");
    };
  }, []);

  return (
    <div className="app">
      <header className="hero">
        <div>
          <p className="eyebrow">Bible API v2</p>
          <h1>Deep study search with OpenSearch</h1>
          <p className="subhead">
            Sign in, explore scripture, and tune your research with fast predictive search.
          </p>
        </div>
        <div className="login-card">
          <div id="google-button" className="google-button" />
          <p className="status">{status}</p>
          <div className="token">
            {hasToken ? token : "Token will appear here after login."}
          </div>
          <button className="ghost" onClick={logout}>
            Logout
          </button>
        </div>
      </header>

      <section className="panel">
        <div className="panel-header">
          <h2>Search</h2>
          <p>Try: grace, covenant, kingdom, faith, mercy.</p>
        </div>
        <div className="fields">
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search the text"
          />
          <button onClick={runSearch} disabled={loading}>
            {loading ? "Searching..." : "Search"}
          </button>
        </div>
        <div className="fields">
          <input
            value={suggestQuery}
            onChange={(event) => setSuggestQuery(event.target.value)}
            placeholder="Predictive suggestion"
          />
        </div>
        <div className="suggestions">
          {suggestions.map((suggestion) => (
            <button
              key={`${suggestion.text}-${suggestion.score}`}
              className="chip"
              onClick={() => {
                setQuery(suggestion.text);
                runSearch();
              }}
            >
              {suggestion.text}
            </button>
          ))}
        </div>
        {error && <div className="error">{error}</div>}
        <div className="results">
          {results.length === 0 && !loading ? (
            <div className="empty">No results yet.</div>
          ) : null}
          {results.map((item) => (
            <article
              key={`${item.book}-${item.chapter}-${item.verse}`}
              className="result"
            >
              <div className="meta">
                {item.book} {item.chapter}:{item.verse}
              </div>
              <div
                className="text"
                dangerouslySetInnerHTML={{ __html: safeHighlight(item) }}
              />
            </article>
          ))}
        </div>
      </section>

      <footer className="footer">
        <span>Powered by OpenSearch</span>
        <span>API docs: /docs</span>
      </footer>
    </div>
  );
}
