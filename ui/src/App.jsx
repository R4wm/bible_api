import { useEffect, useMemo, useState } from "react";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8000";

const apiUrl = (path) => `${API_BASE}${path}`;

const titleCase = (value) =>
  value
    .toLowerCase()
    .split(" ")
    .map((word) => (word ? word[0].toUpperCase() + word.slice(1) : word))
    .join(" ");

export default function App() {
  const [books, setBooks] = useState([]);
  const [chapters, setChapters] = useState([]);
  const [verses, setVerses] = useState([]);
  const [selectedBook, setSelectedBook] = useState("");
  const [selectedChapter, setSelectedChapter] = useState("");
  const [selectedVerse, setSelectedVerse] = useState("all");
  const [reading, setReading] = useState([]);
  const [readingStatus, setReadingStatus] = useState("Select a book to start.");
  const [chapterVerses, setChapterVerses] = useState([]);
  const [openVerses, setOpenVerses] = useState([]);

  const [query, setQuery] = useState("");
  const [searchResults, setSearchResults] = useState([]);
  const [searchStatus, setSearchStatus] = useState("");
  const [suggestQuery, setSuggestQuery] = useState("");
  const [suggestions, setSuggestions] = useState([]);

  useEffect(() => {
    fetch(apiUrl("/bible/list_books?json=true"))
      .then((res) => res.json())
      .then((data) => {
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
    fetch(apiUrl(`/bible/list_chapters/${encodeURIComponent(selectedBook)}?json=true`))
      .then((res) => res.json())
      .then((data) => {
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
    fetch(apiUrl(`/bible/${encodeURIComponent(selectedBook)}/${encodeURIComponent(selectedChapter)}?json=true`))
      .then((res) => res.json())
      .then((data) => {
        const verseCount = data.Verses ? data.Verses.length : 0;
        setVerses(Array.from({ length: verseCount }, (_, i) => i + 1));
        setSelectedVerse("all");
        const fullChapter = (data.Verses || []).map((text, index) => ({ number: index + 1, text }));
        setChapterVerses(fullChapter);
        setReading(fullChapter);
        setReadingStatus("");
      })
      .catch(() => {
        setVerses([]);
        setReading([]);
        setReadingStatus("Failed to load chapter.");
      });
  }, [selectedBook, selectedChapter]);

  useEffect(() => {
    if (!selectedBook || !selectedChapter) {
      return;
    }

    if (selectedVerse === "all") {
      if (chapterVerses.length > 0) {
        setReading(chapterVerses);
        setReadingStatus("");
      }
      return;
    }

    if (!selectedBook || !selectedChapter || selectedVerse === "all") {
      return;
    }

    setReadingStatus("Loading verse...");
    fetch(
      apiUrl(
        `/bible/${encodeURIComponent(selectedBook)}/${encodeURIComponent(selectedChapter)}/${encodeURIComponent(
          selectedVerse
        )}?json=true`
      )
    )
      .then((res) => res.json())
      .then((data) => {
        const parsed = [];
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
  }, [selectedBook, selectedChapter, selectedVerse, chapterVerses]);

  const runSearch = async () => {
    if (!query.trim()) {
      return;
    }
    setSearchStatus("Searching...");
    setSearchResults([]);
    try {
      const res = await fetch(apiUrl(`/bible/v2/search?q=${encodeURIComponent(query)}`));
      if (!res.ok) {
        throw new Error("Search failed");
      }
      const data = await res.json();
      setSearchResults(data?.data?.results || []);
      setSearchStatus("");
    } catch {
      setSearchStatus("Search failed. Check the API.");
    }
  };

  useEffect(() => {
    if (!suggestQuery.trim()) {
      setSuggestions([]);
      return;
    }

    const timer = window.setTimeout(async () => {
      try {
        const res = await fetch(apiUrl(`/bible/v2/suggest?q=${encodeURIComponent(suggestQuery)}`));
        if (!res.ok) {
          return;
        }
        const data = await res.json();
        setSuggestions(data?.data?.suggestions || []);
      } catch {
        setSuggestions([]);
      }
    }, 200);

    return () => window.clearTimeout(timer);
  }, [suggestQuery]);

  const addOpenVerse = (item) => {
    const key = `${item.book}-${item.chapter}-${item.verse}`;
    setOpenVerses((prev) => {
      if (prev.some((entry) => entry.key === key)) {
        return prev;
      }
      return [
        {
          key,
          book: item.book,
          chapter: item.chapter,
          verse: item.verse,
          text: item.text
        },
        ...prev
      ];
    });
  };

  const removeOpenVerse = (key) => {
    setOpenVerses((prev) => prev.filter((entry) => entry.key !== key));
  };

  const safeHighlight = useMemo(() => {
    return (item) => {
      if (!item.highlight || item.highlight.length === 0) {
        return item.text;
      }
      return item.highlight[0]
        .replace(/<em>/g, "<mark>")
        .replace(/<\/em>/g, "</mark>");
    };
  }, []);

  const bookMetrics = useMemo(() => {
    const counts = new Map();
    searchResults.forEach((item) => {
      const book = item.book || "UNKNOWN";
      counts.set(book, (counts.get(book) || 0) + 1);
    });
    const rows = Array.from(counts.entries())
      .map(([book, count]) => ({ book, count }))
      .sort((a, b) => b.count - a.count);
    const max = rows[0]?.count || 0;
    return { rows, max, total: searchResults.length };
  }, [searchResults]);

  return (
    <div className="app">
      <header className="hero">
        <div className="metrics">
          <p className="eyebrow">Search Metrics</p>
          {bookMetrics.rows.length === 0 ? (
            <p className="subhead">Run a search to populate the study dashboard.</p>
          ) : (
            <>
              <div className="metric-grid">
                <div className="metric-card">
                  <span>Query</span>
                  <strong>{query ? `"${query}"` : "—"}</strong>
                </div>
                <div className="metric-card">
                  <span>Total Results</span>
                  <strong>{bookMetrics.total}</strong>
                </div>
                <div className="metric-card">
                  <span>Books Matched</span>
                  <strong>{bookMetrics.rows.length}</strong>
                </div>
                <div className="metric-card">
                  <span>Top Book</span>
                  <strong>{bookMetrics.rows[0]?.book || "—"}</strong>
                </div>
              </div>
              <div className="metric-focus">
                <div className="metric-line">
                  <span>Concentration</span>
                  <strong>
                    {Math.round((bookMetrics.rows[0]?.count / bookMetrics.total) * 100)}%
                  </strong>
                </div>
                <div className="stat-bar">
                  <span
                    style={{
                      width: `${bookMetrics.rows[0]?.count && bookMetrics.total
                        ? (bookMetrics.rows[0].count / bookMetrics.total) * 100
                        : 0}%`
                    }}
                  />
                </div>
              </div>
              <div className="chart compact scroll">
                {bookMetrics.rows.map((row) => (
                  <div key={row.book} className="chart-row">
                    <div className="chart-label">{row.book}</div>
                    <div className="chart-bar">
                      <span
                        style={{
                          width: `${bookMetrics.max ? (row.count / bookMetrics.max) * 100 : 0}%`
                        }}
                      />
                    </div>
                    <div className="chart-value">{row.count}</div>
                  </div>
                ))}
              </div>
            </>
          )}
          <div className="hint">
            API: {API_BASE} · Set <code>VITE_API_BASE_URL</code>
          </div>
        </div>
      </header>

      <section className="dashboard">
        <div className="panel">
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
            <div className="results">
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
            <div className="panel-header">
              <h3>Open Verses</h3>
              <p>Collect verses from search to compare and study.</p>
            </div>
            <div className="open-verses">
              {openVerses.length === 0 ? (
                <div className="empty">No verses pinned yet.</div>
              ) : null}
              {openVerses.map((verse) => (
                <article key={verse.key} className="result open-verse">
                  <div className="meta">
                    {verse.book} {verse.chapter}:{verse.verse}
                  </div>
                  <div className="text">{verse.text}</div>
                  <button className="ghost" onClick={() => removeOpenVerse(verse.key)}>
                    Remove
                  </button>
                </article>
              ))}
            </div>
          </div>
        </div>

        <div className="panel">
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
            <button onClick={runSearch}>Search</button>
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
          {searchStatus ? <div className="hint">{searchStatus}</div> : null}
          <div className="results">
            {searchResults.length === 0 && !searchStatus ? (
              <div className="empty">No results yet.</div>
            ) : null}
            {searchResults.map((item) => (
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
                <button className="ghost" onClick={() => addOpenVerse(item)}>
                  Pin verse
                </button>
              </article>
            ))}
          </div>
        </div>
      </section>

      <footer className="footer">
        <span>Powered by Bible API</span>
        <span>Docs: {API_BASE}/docs</span>
      </footer>
    </div>
  );
}
