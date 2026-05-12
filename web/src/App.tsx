import { useEffect, useMemo, useRef, useState } from "react";

declare const Chart: any;

const defaultScope = "synonyms:write";

const BOOKS_CANONICAL_ORDER = [
  "GENESIS","EXODUS","LEVITICUS","NUMBERS","DEUTERONOMY",
  "JOSHUA","JUDGES","RUTH","1SAMUEL","2SAMUEL",
  "1KINGS","2KINGS","1CHRONICLES","2CHRONICLES",
  "EZRA","NEHEMIAH","ESTHER","JOB","PSALMS","PROVERBS",
  "ECCLESIASTES","SONG OF SOLOMON","ISAIAH","JEREMIAH","LAMENTATIONS",
  "EZEKIEL","DANIEL","HOSEA","JOEL","AMOS",
  "OBADIAH","JONAH","MICAH","NAHUM","HABAKKUK",
  "ZEPHANIAH","HAGGAI","ZECHARIAH","MALACHI",
  "MATTHEW","MARK","LUKE","JOHN","ACTS",
  "ROMANS","1CORINTHIANS","2CORINTHIANS","GALATIANS","EPHESIANS",
  "PHILIPPIANS","COLOSSIANS","1THESSALONIANS","2THESSALONIANS",
  "1TIMOTHY","2TIMOTHY","TITUS","PHILEMON","HEBREWS",
  "JAMES","1PETER","2PETER","1JOHN","2JOHN","3JOHN",
  "JUDE","REVELATION",
];

const BOOK_ABBREVIATIONS: Record<string, string> = {
  "GENESIS": "Gen", "EXODUS": "Ex", "LEVITICUS": "Lev", "NUMBERS": "Num", "DEUTERONOMY": "Deut",
  "JOSHUA": "Josh", "JUDGES": "Judg", "RUTH": "Ruth", "1SAMUEL": "1 Sam", "2SAMUEL": "2 Sam",
  "1KINGS": "1 Kgs", "2KINGS": "2 Kgs", "1CHRONICLES": "1 Chr", "2CHRONICLES": "2 Chr",
  "EZRA": "Ezra", "NEHEMIAH": "Neh", "ESTHER": "Esth", "JOB": "Job", "PSALMS": "Ps", "PROVERBS": "Prov",
  "ECCLESIASTES": "Eccl", "SONG OF SOLOMON": "Song", "ISAIAH": "Isa", "JEREMIAH": "Jer", "LAMENTATIONS": "Lam",
  "EZEKIEL": "Ezek", "DANIEL": "Dan", "HOSEA": "Hos", "JOEL": "Joel", "AMOS": "Amos",
  "OBADIAH": "Obad", "JONAH": "Jonah", "MICAH": "Mic", "NAHUM": "Nah", "HABAKKUK": "Hab",
  "ZEPHANIAH": "Zeph", "HAGGAI": "Hag", "ZECHARIAH": "Zech", "MALACHI": "Mal",
  "MATTHEW": "Matt", "MARK": "Mark", "LUKE": "Luke", "JOHN": "John", "ACTS": "Acts",
  "ROMANS": "Rom", "1CORINTHIANS": "1 Cor", "2CORINTHIANS": "2 Cor", "GALATIANS": "Gal", "EPHESIANS": "Eph",
  "PHILIPPIANS": "Phil", "COLOSSIANS": "Col", "1THESSALONIANS": "1 Thess", "2THESSALONIANS": "2 Thess",
  "1TIMOTHY": "1 Tim", "2TIMOTHY": "2 Tim", "TITUS": "Titus", "PHILEMON": "Phlm", "HEBREWS": "Heb",
  "JAMES": "Jas", "1PETER": "1 Pet", "2PETER": "2 Pet", "1JOHN": "1 John", "2JOHN": "2 John", "3JOHN": "3 John",
  "JUDE": "Jude", "REVELATION": "Rev",
};

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
  book?: string;
  chapter?: number;
  verse?: number;
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

type ReadVerse = {
  number: number;
  text: string;
};

type View = "books" | "chapters" | "reading" | "search";

type GoogleProfile = { picture?: string; name?: string; email?: string };

const decodeJwtPayload = (jwt: string): any => {
  try {
    const seg = jwt.split(".")[1].replace(/-/g, "+").replace(/_/g, "/");
    return JSON.parse(atob(seg));
  } catch { return {}; }
};

export default function App() {
  const [token, setToken] = useState("");
  const [config, setConfig] = useState<AuthConfig>({});

  const [view, setView] = useState<View>("books");
  const [books, setBooks] = useState<string[]>([]);
  const [selectedBook, setSelectedBook] = useState("");
  const [chapters, setChapters] = useState<number[]>([]);
  const [selectedChapter, setSelectedChapter] = useState(0);
  const [reading, setReading] = useState<ReadVerse[]>([]);

  const [query, setQuery] = useState("");
  const [suggestQuery, setSuggestQuery] = useState("");
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [bookCounts, setBookCounts] = useState<Record<string, number>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [menuOpen, setMenuOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [profileOpen, setProfileOpen] = useState(false);
  const [profile, setProfile] = useState<GoogleProfile>(() => {
    try { return JSON.parse(localStorage.getItem("bible-profile") || "{}"); }
    catch { return {}; }
  });
  const [font, setFont] = useState(() => localStorage.getItem("bible-font") || "renaissance");
  const [theme, setTheme] = useState<"light" | "dark">(
    () => localStorage.getItem("bible-theme") === "light" ? "light" : "dark"
  );

  const hasToken = token.length > 0;
  const searchInputRef = useRef<HTMLInputElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const menuBtnRef = useRef<HTMLButtonElement>(null);
  const settingsRef = useRef<HTMLDivElement>(null);
  const profileRef = useRef<HTMLDivElement>(null);
  const profileBtnRef = useRef<HTMLButtonElement>(null);
  const chartRef = useRef<HTMLCanvasElement>(null);
  const chartInstance = useRef<any>(null);

  const fontClasses = ["font-blackletter", "font-renaissance", "font-serif"];

  // Apply font class on mount and when font changes
  useEffect(() => {
    fontClasses.forEach((c) => document.body.classList.remove(c));
    if (font !== "default") document.body.classList.add("font-" + font);
    localStorage.setItem("bible-font", font);
  }, [font]);

  useEffect(() => {
    document.body.classList.remove("theme-dark");
    if (theme === "dark") document.body.classList.add("theme-dark");
    localStorage.setItem("bible-theme", theme);
  }, [theme]);

  useEffect(() => {
    if (view === "reading" && selectedBook && selectedChapter) {
      const abbr = BOOK_ABBREVIATIONS[selectedBook] ?? selectedBook;
      document.title = `${abbr} ${selectedChapter}`;
    } else {
      document.title = "Bible API";
    }
  }, [view, selectedBook, selectedChapter]);

  // Close menu on Escape and outside click
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") { setMenuOpen(false); setSettingsOpen(false); setProfileOpen(false); }
    };
    const onClick = (e: MouseEvent) => {
      const target = e.target as Node;
      const insideMenu = (menuRef.current && menuRef.current.contains(target)) ||
        (menuBtnRef.current && menuBtnRef.current.contains(target)) ||
        (settingsRef.current && settingsRef.current.contains(target));
      const insideProfile = (profileRef.current && profileRef.current.contains(target)) ||
        (profileBtnRef.current && profileBtnRef.current.contains(target));
      if (!insideMenu && !insideProfile) {
        setMenuOpen(false);
        setSettingsOpen(false);
        setProfileOpen(false);
      }
    };
    document.addEventListener("keydown", onKey);
    document.addEventListener("click", onClick);
    return () => {
      document.removeEventListener("keydown", onKey);
      document.removeEventListener("click", onClick);
    };
  }, []);

  // Render chart when search results change
  useEffect(() => {
    if (view !== "search" || !chartRef.current || typeof Chart === "undefined") return;

    const axisColor = theme === "dark" ? "#ddd" : "#666";
    const gridColor = theme === "dark" ? "rgba(255,255,255,0.1)" : "rgba(0,0,0,0.1)";

    const counts = BOOKS_CANONICAL_ORDER.map((b) => bookCounts[b] || 0);
    const colors = counts.map((val, i) => {
      if (val === 0) return "rgba(200,200,200,0.3)";
      return i < 39 ? "rgba(220, 88, 42, 0.7)" : "rgba(34, 139, 134, 0.7)";
    });
    const borders = counts.map((val, i) => {
      if (val === 0) return "rgba(200,200,200,0.5)";
      return i < 39 ? "rgba(180, 60, 20, 1)" : "rgba(20, 100, 100, 1)";
    });

    chartInstance.current?.destroy();
    chartInstance.current = new Chart(chartRef.current, {
      type: "bar",
      data: {
        labels: BOOKS_CANONICAL_ORDER,
        datasets: [{
          label: "Matches by book",
          data: counts,
          backgroundColor: colors,
          borderColor: borders,
          borderWidth: 1,
          borderRadius: 2,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (item: any) => {
                const v = item.raw;
                return v + " match" + (v !== 1 ? "es" : "");
              },
            },
          },
        },
        scales: {
          x: {
            ticks: { maxRotation: 90, minRotation: 45, font: { size: 10 }, color: axisColor },
            grid: { display: false },
          },
          y: {
            beginAtZero: true,
            ticks: { precision: 0, color: axisColor },
            grid: { color: gridColor },
            title: { display: true, text: "Matches", color: axisColor },
          },
        },
      },
    });

    return () => {
      chartInstance.current?.destroy();
      chartInstance.current = null;
    };
  }, [view, bookCounts, theme]);

  useEffect(() => {
    fetch("/auth/config")
      .then((res) => res.json())
      .then((data) => setConfig(data))
      .catch(() => setConfig({}));
  }, []);

  useEffect(() => {
    fetch("/auth/me")
      .then((res) => {
        if (!res.ok) return null;
        return res.json();
      })
      .then((data) => {
        if (!data) return;
        if (data.token) setToken(data.token);
        if (data.picture || data.name || data.email) {
          const fromServer: GoogleProfile = {
            picture: data.picture || undefined,
            name: data.name || undefined,
            email: data.email || undefined,
          };
          setProfile(fromServer);
          localStorage.setItem("bible-profile", JSON.stringify(fromServer));
        }
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!config.google_client_id) return;
    let cancelled = false;
    const initGoogle = () => {
      if (cancelled) return;
      const google = (window as any).google;
      if (!google?.accounts?.id) {
        window.setTimeout(initGoogle, 300);
        return;
      }
      google.accounts.id.initialize({
        client_id: config.google_client_id,
        callback: async (response: { credential: string }) => {
          const claims = decodeJwtPayload(response.credential);
          const newProfile: GoogleProfile = {
            picture: claims.picture,
            name: claims.name,
            email: claims.email,
          };
          localStorage.setItem("bible-profile", JSON.stringify(newProfile));
          setProfile(newProfile);
          const resp = await fetch("/auth/google/token", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id_token: response.credential, scope: defaultScope }),
          });
          if (!resp.ok) { setError("Login failed."); return; }
          const data = await resp.json();
          setToken(data.token || "");
        },
      });
      google.accounts.id.renderButton(document.getElementById("google-button"), {
        theme: "outline",
        size: "large",
      });
    };
    initGoogle();
    return () => { cancelled = true; };
  }, [config.google_client_id, hasToken, profileOpen]);

  useEffect(() => {
    fetch("/bible/list_books?json=true")
      .then((res) => res.json())
      .then((data: BooksResponse) => setBooks(data.Books || []))
      .catch(() => setBooks([]));
  }, []);

  const loadChapters = (book: string) => {
    setSelectedBook(book);
    setView("chapters");
    fetch(`/bible/list_chapters/${encodeURIComponent(book)}?json=true`)
      .then((res) => res.json())
      .then((data: ChaptersResponse) => setChapters(data.Chapters || []))
      .catch(() => setChapters([]));
  };

  const loadChapter = (chapter: number) => {
    setSelectedChapter(chapter);
    setView("reading");
    fetch(`/bible/${encodeURIComponent(selectedBook)}/${chapter}?json=true`)
      .then((res) => res.json())
      .then((data: ChapterResponse) => {
        setReading((data.Verses || []).map((text, i) => ({ number: i + 1, text })));
      })
      .catch(() => setReading([]));
  };

  const runSearch = async () => {
    if (!query.trim()) return;
    setLoading(true);
    setError("");
    try {
      const resp = await fetch(`/bible/v2/search?q=${encodeURIComponent(query)}`);
      if (!resp.ok) throw new Error("Search failed");
      const data = await resp.json();
      setResults(data?.data?.results || []);
      setBookCounts(data?.data?.book_counts || {});
      setView("search");
    } catch {
      setError("Search failed. Check the API.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!suggestQuery.trim()) { setSuggestions([]); return; }
    const timer = window.setTimeout(async () => {
      try {
        const resp = await fetch(`/bible/v2/suggest?q=${encodeURIComponent(suggestQuery)}`);
        if (!resp.ok) return;
        const data = await resp.json();
        setSuggestions(data?.data?.suggestions || []);
      } catch { setSuggestions([]); }
    }, 200);
    return () => window.clearTimeout(timer);
  }, [suggestQuery]);

  const openSuggestion = (suggestion?: Suggestion) => {
    if (!suggestion?.book || !suggestion.chapter || !suggestion.verse) return;
    window.open(
      `/bible/${encodeURIComponent(suggestion.book)}/${suggestion.chapter}/${suggestion.verse}`,
      "_blank",
      "noopener,noreferrer",
    );
  };

  const logout = async () => {
    await fetch("/auth/logout", { method: "POST" });
    setToken("");
    setProfile({});
    localStorage.removeItem("bible-profile");
    setProfileOpen(false);
  };

  const safeHighlight = useMemo(() => {
    return (item: SearchResult) => {
      if (!item.highlight || item.highlight.length === 0) return item.text;
      return item.highlight[0].replace(/<em>/g, "<mark>").replace(/<\/em>/g, "</mark>");
    };
  }, []);

  return (
    <div className="app">
      {/* Top bar: hamburger + search */}
      <div className="top-bar">
        <button
          className="hamburger"
          ref={menuBtnRef}
          aria-label="Open navigation menu"
          aria-expanded={menuOpen}
          onClick={() => setMenuOpen(!menuOpen)}
        >
          &#9776;
        </button>
        <div className="search-bar">
          <input
            ref={searchInputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") runSearch(); }}
            placeholder="Search the text"
          />
          <button onClick={runSearch} disabled={loading}>
            {loading ? "..." : "search"}
          </button>
        </div>
        <button
          className="profile-btn"
          ref={profileBtnRef}
          aria-label={hasToken ? "Open profile menu" : "Sign in"}
          aria-expanded={profileOpen}
          onClick={() => setProfileOpen(!profileOpen)}
        >
          {hasToken && profile.picture ? (
            <img src={profile.picture} alt="" referrerPolicy="no-referrer" />
          ) : hasToken && profile.name ? (
            <span className="profile-initial">{profile.name.charAt(0).toUpperCase()}</span>
          ) : (
            <span className="profile-icon" aria-hidden="true">&#x1F464;</span>
          )}
        </button>
      </div>

      {/* Profile panel */}
      {profileOpen && (
        <div className="profile-panel" ref={profileRef}>
          <div id="google-button" style={{ display: hasToken ? "none" : "block" }} />
          {hasToken && (
            <>
              {profile.name && <div className="profile-name">{profile.name}</div>}
              {profile.email && <div className="profile-email">{profile.email}</div>}
              <button className="profile-logout" onClick={logout}>Logout</button>
            </>
          )}
        </div>
      )}

      {/* Menu panel */}
      {menuOpen && (
        <div className="menu-panel" ref={menuRef}>
          <button onClick={() => { setView("books"); setMenuOpen(false); }}>Books</button>
          <button onClick={() => { setView("search"); searchInputRef.current?.focus(); setMenuOpen(false); }}>Search</button>
          <a href="/docs">Docs</a>
          <button onClick={() => setSettingsOpen(!settingsOpen)}>Settings</button>
          <a href="/bible/list_books">Open Classic</a>
        </div>
      )}

      {/* Settings panel */}
      {settingsOpen && (
        <div className="settings-panel open" ref={settingsRef}>
          <strong>Font</strong>
          {[
            { value: "default", label: "Default" },
            { value: "blackletter", label: "Blackletter (Gothic)" },
            { value: "renaissance", label: "Renaissance" },
            { value: "serif", label: "Classic Serif" },
          ].map((opt) => (
            <label key={opt.value}>
              <input
                type="radio"
                name="font-choice"
                value={opt.value}
                checked={font === opt.value}
                onChange={() => setFont(opt.value)}
              />
              {opt.label}
            </label>
          ))}
          <strong>Theme</strong>
          {[
            { value: "light", label: "Light" },
            { value: "dark", label: "Dark" },
          ].map((opt) => (
            <label key={opt.value}>
              <input
                type="radio"
                name="theme-choice"
                value={opt.value}
                checked={theme === opt.value}
                onChange={() => setTheme(opt.value as "light" | "dark")}
              />
              {opt.label}
            </label>
          ))}
        </div>
      )}

      {/* Suggest bar */}
      <div className="search-bar">
        <input
          value={suggestQuery}
          onChange={(e) => setSuggestQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") openSuggestion(suggestions[0]);
          }}
          placeholder="Predictive suggestion"
        />
      </div>
      {suggestions.length > 0 && (
        <div className="suggestions">
          {suggestions.map((s) => (
            <button key={`${s.book ?? ""}-${s.chapter ?? ""}-${s.verse ?? ""}-${s.text}-${s.score}`} className="chip" onClick={() => openSuggestion(s)}>
              {s.book && s.chapter && s.verse && (
                <span className="chip-ref">
                  {BOOK_ABBREVIATIONS[s.book] ?? s.book} {s.chapter}:{s.verse}
                </span>
              )}
              {s.book && s.chapter && s.verse && <span className="chip-separator">—</span>}
              <span className="chip-text">{s.text}</span>
            </button>
          ))}
        </div>
      )}
      {error && <div className="error">{error}</div>}

      {/* Navigation */}
      {view !== "books" && (
        <div className="nav-bar">
          <button className="nav-btn" onClick={() => setView("books")}>Books Menu</button>
          {view === "reading" && (
            <button className="nav-btn" onClick={() => setView("chapters")}>{selectedBook}</button>
          )}
          {view === "reading" && selectedChapter > 1 && (
            <button className="nav-btn" onClick={() => loadChapter(selectedChapter - 1)}>&lt;</button>
          )}
          {view === "reading" && selectedChapter < chapters.length && (
            <button className="nav-btn" onClick={() => loadChapter(selectedChapter + 1)}>&gt;</button>
          )}
        </div>
      )}

      {/* Books list */}
      {view === "books" && (
        <div className="books-grid">
          {books.map((book) => (
            <button key={book} className="block" onClick={() => loadChapters(book)}>
              {book}
            </button>
          ))}
        </div>
      )}

      {/* Chapters list */}
      {view === "chapters" && (
        <>
          <h2>{selectedBook}</h2>
          <div className="chapters-grid">
            {chapters.map((ch) => (
              <button key={ch} className="block" onClick={() => loadChapter(ch)}>
                {ch}
              </button>
            ))}
          </div>
        </>
      )}

      {/* Reading */}
      {view === "reading" && (
        <>
          <h2>{selectedBook} {selectedChapter}</h2>
          {reading.map((v) => (
            <p key={v.number} className="verse-text">
              <span className="verse-num">{v.number}</span> {v.text}
            </p>
          ))}
        </>
      )}

      {/* Search results */}
      {view === "search" && (
        <>
          <h2>Search Results</h2>
          <div className="chart-container">
            <canvas ref={chartRef} />
          </div>
          {results.length === 0 && !loading && <div className="empty">No results.</div>}
          {results.map((item) => (
            <div key={`${item.book}-${item.chapter}-${item.verse}`} className="result">
              <div className="meta">
                <a href={`/bible/${item.book}/${item.chapter}/${item.verse}?json=false`}>
                  {item.book} {item.chapter}:{item.verse}
                </a>
              </div>
              <div className="text" dangerouslySetInnerHTML={{ __html: safeHighlight(item) }} />
            </div>
          ))}
        </>
      )}

      <div className="footer">
        <span>Powered by OpenSearch</span> | <span>API docs: <a href="/docs">/docs</a></span>
      </div>
    </div>
  );
}
