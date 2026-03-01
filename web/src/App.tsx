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

type BooksResponse = {
  Books?: string[];
};

type ChaptersResponse = {
  Chapters?: number[];
  Name?: string;
};

type ChapterResponse = {
  BookName?: string;
  Chapter?: number;
  Verses?: string[];
};

type VerseResponse = {
  BookName?: string;
  Chapter?: number;
  Verses?: Record<string, string>[];
};

type ReadVerse = {
  number: number;
  text: string;
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
  const [books, setBooks] = useState<string[]>([]);
  const [chapters, setChapters] = useState<number[]>([]);
  const [verses, setVerses] = useState<number[]>([]);
  const [selectedBook, setSelectedBook] = useState<string>("");
  const [selectedChapter, setSelectedChapter] = useState<string>("");
  const [selectedVerse, setSelectedVerse] = useState<string>("all");
  const [reading, setReading] = useState<ReadVerse[]>([]);
  const [readingStatus, setReadingStatus] = useState<string>("Select a book to start.");

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

  useEffect(() => {
    fetch("/bible/list_books?json=true")
      .then((res) => res.json())
      .then((data: BooksResponse) => {
        const sorted = [...(data.Books || [])].sort((a, b) => a.localeCompare(b));
        setBooks(sorted);
      })
      .catch(() => setBooks([]));
  }, []);

  useEffect(() => {
    if (!selectedBook) {
      setChapters([]);
      setSelectedChapter("");
      setSelectedVerse("all");
      setReading([]);
      setReadingStatus("Select a book to start.");
      return;
    }

    setReadingStatus("Loading chapters...");
    fetch(`/bible/list_chapters/${encodeURIComponent(selectedBook)}?json=true`)
      .then((res) => res.json())
      .then((data: ChaptersResponse) => {
        setChapters(data.Chapters || []);
        setSelectedChapter("");
        setSelectedVerse("all");
        setReading([]);
        setReadingStatus(`Select a chapter in ${titleCase(data.Name || selectedBook)}.`);
      })
      .catch(() => {
        setChapters([]);
        setReading([]);
        setReadingStatus("Failed to load chapters.");
      });
  }, [selectedBook]);

  useEffect(() => {
    if (!selectedBook || !selectedChapter) {
      setVerses([]);
      setReading([]);
      return;
    }

    setReadingStatus("Loading chapter...");
    fetch(`/bible/${encodeURIComponent(selectedBook)}/${encodeURIComponent(selectedChapter)}?json=true`)
      .then((res) => res.json())
      .then((data: ChapterResponse) => {
        const verseCount = data.Verses ? data.Verses.length : 0;
        setVerses(Array.from({ length: verseCount }, (_, i) => i + 1));
        setSelectedVerse("all");
        setReading((data.Verses || []).map((text, index) => ({ number: index + 1, text })));
        setReadingStatus("");
      })
      .catch(() => {
        setVerses([]);
        setReading([]);
        setReadingStatus("Failed to load chapter.");
      });
  }, [selectedBook, selectedChapter]);

  useEffect(() => {
    if (!selectedBook || !selectedChapter || selectedVerse === "all") {
      return;
    }

    setReadingStatus("Loading verse...");
    fetch(
      `/bible/${encodeURIComponent(selectedBook)}/${encodeURIComponent(selectedChapter)}/${encodeURIComponent(
        selectedVerse
      )}?json=true`
    )
      .then((res) => res.json())
      .then((data: VerseResponse) => {
        const parsed: ReadVerse[] = [];
        (data.Verses || []).forEach((entry) => {
          const [key, value] = Object.entries(entry)[0] || [];
          const num = Number(key);
          if (!Number.isNaN(num) && typeof value === "string") {
            parsed.push({ number: num, text: value });
          }
        });
        setReading(parsed);
        setReadingStatus(parsed.length === 0 ? "Verse not found." : "");
      })
      .catch(() => {
        setReading([]);
        setReadingStatus("Failed to load verse.");
      });
  }, [selectedBook, selectedChapter, selectedVerse]);

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

  const titleCase = useMemo(() => {
    return (value: string) =>
      value
        .toLowerCase()
        .split(" ")
        .map((word) => (word ? word[0].toUpperCase() + word.slice(1) : word))
        .join(" ");
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

      <section className="panel">
        <div className="panel-header">
          <h2>Read</h2>
          <p>Navigate by book, chapter, and verse.</p>
        </div>
        <div className="navigator">
          <div className="select-row">
            <label>
              Book
              <select
                value={selectedBook}
                onChange={(event) => setSelectedBook(event.target.value)}
              >
                <option value="">Select a book</option>
                {books.map((book) => (
                  <option key={book} value={book}>
                    {titleCase(book)}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Chapter
              <select
                value={selectedChapter}
                onChange={(event) => setSelectedChapter(event.target.value)}
                disabled={!selectedBook || chapters.length === 0}
              >
                <option value="">Select chapter</option>
                {chapters.map((chapter) => (
                  <option key={chapter} value={chapter}>
                    {chapter}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Verse
              <select
                value={selectedVerse}
                onChange={(event) => setSelectedVerse(event.target.value)}
                disabled={!selectedChapter || verses.length === 0}
              >
                <option value="all">All verses</option>
                {verses.map((verse) => (
                  <option key={verse} value={verse}>
                    {verse}
                  </option>
                ))}
              </select>
            </label>
          </div>
          {readingStatus ? <div className="hint">{readingStatus}</div> : null}
          <div className="read-results">
            {reading.length === 0 && !readingStatus ? (
              <div className="empty">No verses loaded yet.</div>
            ) : null}
            {reading.map((verse) => (
              <article key={verse.number} className="result">
                <div className="meta">
                  {selectedBook ? titleCase(selectedBook) : ""} {selectedChapter}:
                  {verse.number}
                </div>
                <div className="text">{verse.text}</div>
              </article>
            ))}
          </div>
        </div>
      </section>

      <footer className="footer">
        <span>Powered by OpenSearch</span>
        <span>API docs: /docs</span>
      </footer>
    </div>
  );
}
