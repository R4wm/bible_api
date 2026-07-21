import { useEffect, useMemo, useRef, useState } from "react";

declare const Chart: any;

const defaultScope = "synonyms:write";
const suggestionPageSize = 20;

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

type RecentPage = {
  book: string;
  chapter: number;
  ts?: number;
};

type SearchHistory = {
  query: string;
  ts?: number;
};

type VerseOpenMode = "new-tab" | "same-tab";

type VerseSelection = {
  verses: number[];
  label: string;
};

type View = "books" | "chapters" | "reading" | "search";
type SettingsTab = "preferences" | "history";

type GoogleProfile = { picture?: string; name?: string; email?: string };

const decodeJwtPayload = (jwt: string): any => {
  try {
    const seg = jwt.split(".")[1].replace(/-/g, "+").replace(/_/g, "/");
    return JSON.parse(atob(seg));
  } catch { return {}; }
};

const defaultVerseOpenMode: VerseOpenMode = "new-tab";

const storedVerseOpenMode = (): VerseOpenMode =>
  localStorage.getItem("bible-verse-open-mode") === "same-tab" ? "same-tab" : defaultVerseOpenMode;

const isVerseOpenMode = (value: unknown): value is VerseOpenMode =>
  value === "new-tab" || value === "same-tab";

const positiveIntParam = (value: string | null): number => {
  const parsed = Number.parseInt(value || "", 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
};

const formatHistoryTime = (timestamp?: number): string =>
  timestamp ? new Date(timestamp * 1000).toLocaleString() : "";

const parseVerseSelection = (value: string | null): VerseSelection => {
  const raw = (value || "").trim();
  if (!raw) return { verses: [], label: "" };

  const selected = new Set<number>();
  const normalized = raw.replace(/[–—]/g, "-");

  normalized.split(",").forEach((part) => {
    const token = part.trim();
    if (!token) return;

    const range = token.match(/^(\d+)\s*-\s*(\d+)$/);
    if (range) {
      const start = Number.parseInt(range[1], 10);
      const end = Number.parseInt(range[2], 10);
      const low = Math.min(start, end);
      const high = Math.max(start, end);
      for (let n = low; n <= high && selected.size < 250; n += 1) {
        if (n > 0) selected.add(n);
      }
      return;
    }

    const single = token.match(/^\d+$/);
    if (single) {
      const n = Number.parseInt(token, 10);
      if (n > 0) selected.add(n);
    }
  });

  const verses = Array.from(selected).sort((a, b) => a - b);
  return {
    verses,
    label: verses.length > 0 ? raw.replace(/\s+/g, "") : "",
  };
};

const v2ReaderURL = (book?: string, chapter?: number, verse?: number | string): string => {
  const params = new URLSearchParams();
  if (book) params.set("book", book);
  if (chapter && chapter > 0) params.set("chapter", String(chapter));
  if (typeof verse === "number" && verse > 0) params.set("verse", String(verse));
  if (typeof verse === "string" && verse.trim()) params.set("verse", verse.trim());
  const query = params.toString();
  return query ? `/v2/?${query}` : "/v2/";
};

export default function App() {
  const [token, setToken] = useState("");
  const [config, setConfig] = useState<AuthConfig>({});

  const [view, setView] = useState<View>("books");
  const [books, setBooks] = useState<string[]>([]);
  const [selectedBook, setSelectedBook] = useState("");
  const [chapters, setChapters] = useState<number[]>([]);
  const [selectedChapter, setSelectedChapter] = useState(0);
  const [selectedVerses, setSelectedVerses] = useState<number[]>([]);
  const [selectedVerseLabel, setSelectedVerseLabel] = useState("");
  const [reading, setReading] = useState<ReadVerse[]>([]);

  const [query, setQuery] = useState("");
  const [suggestQuery, setSuggestQuery] = useState("");
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [suggestionOffset, setSuggestionOffset] = useState(0);
  const [suggestionsLoading, setSuggestionsLoading] = useState(false);
  const [hasMoreSuggestions, setHasMoreSuggestions] = useState(false);
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
  const [verseOpenMode, setVerseOpenMode] = useState<VerseOpenMode>(storedVerseOpenMode);
  const [recent, setRecent] = useState<RecentPage[]>([]);
  const [searchHistory, setSearchHistory] = useState<SearchHistory[]>([]);
  const [settingsTab, setSettingsTab] = useState<SettingsTab>("preferences");
  // True once we've pulled server-side settings for the logged-in user, so we
  // don't push local defaults back up before knowing what the server has.
  const settingsHydrated = useRef(false);

  const hasToken = token.length > 0;
  const searchInputRef = useRef<HTMLInputElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const menuBtnRef = useRef<HTMLButtonElement>(null);
  const settingsRef = useRef<HTMLDivElement>(null);
  const profileRef = useRef<HTMLDivElement>(null);
  const profileBtnRef = useRef<HTMLButtonElement>(null);
  const chartRef = useRef<HTMLCanvasElement>(null);
  const chartInstance = useRef<any>(null);
  const suggestionsSentinelRef = useRef<HTMLDivElement>(null);
  const suggestionRequestRef = useRef(0);

  const fontClasses = ["font-blackletter", "font-renaissance", "font-serif"];
  const selectedVerseSet = useMemo(() => new Set(selectedVerses), [selectedVerses]);

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
    localStorage.setItem("bible-verse-open-mode", verseOpenMode);
  }, [verseOpenMode]);

  // On login, pull the user's saved settings + reading history from the server.
  // Server settings win over local defaults so preferences follow the account.
  useEffect(() => {
    if (!hasToken) {
      settingsHydrated.current = false;
      setRecent([]);
      setSearchHistory([]);
      setSettingsTab("preferences");
      return;
    }
    fetch("/user/settings")
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data?.font) setFont(data.font);
        if (data?.theme === "light" || data?.theme === "dark") setTheme(data.theme);
        if (isVerseOpenMode(data?.verse_open_mode)) setVerseOpenMode(data.verse_open_mode);
      })
      .catch(() => {})
      .finally(() => { settingsHydrated.current = true; });
    fetch("/user/history")
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (Array.isArray(data?.pages)) setRecent(data.pages);
        if (Array.isArray(data?.searches)) setSearchHistory(data.searches);
      })
      .catch(() => {});
  }, [hasToken]);

  // Push settings changes to the server, but only after hydration so we never
  // overwrite saved preferences with the local defaults on first load.
  useEffect(() => {
    if (!hasToken || !settingsHydrated.current) return;
    fetch("/user/settings", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ font, theme, verse_open_mode: verseOpenMode }),
    }).catch(() => {});
  }, [font, theme, verseOpenMode, hasToken]);

  useEffect(() => {
    if (view === "reading" && selectedBook && selectedChapter) {
      const abbr = BOOK_ABBREVIATIONS[selectedBook] ?? selectedBook;
      document.title = `${abbr} ${selectedChapter}${selectedVerseLabel ? `:${selectedVerseLabel}` : ""}`;
    } else {
      document.title = "Bible API";
    }
  }, [view, selectedBook, selectedChapter, selectedVerseLabel]);

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

  useEffect(() => {
    if (view !== "reading" || selectedVerses.length === 0 || reading.length === 0) return;
    const firstSelectedVerse = selectedVerses[0];
    window.setTimeout(() => {
      document.getElementById(`verse-${firstSelectedVerse}`)?.scrollIntoView({
        block: "center",
        behavior: "smooth",
      });
    }, 50);
  }, [view, selectedVerses, reading]);

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

  const pushReaderLocation = (book?: string, chapter?: number, verse?: string) => {
    const url = v2ReaderURL(book, chapter, verse);
    if (`${window.location.pathname}${window.location.search}` !== url) {
      window.history.pushState(null, "", url);
    }
  };

  const showBooks = (updateURL = true) => {
    setSelectedBook("");
    setSelectedChapter(0);
    setSelectedVerses([]);
    setSelectedVerseLabel("");
    setView("books");
    if (updateURL) pushReaderLocation();
  };

  const loadChapters = (book: string, updateURL = true) => {
    setSelectedBook(book);
    setSelectedChapter(0);
    setSelectedVerses([]);
    setSelectedVerseLabel("");
    setView("chapters");
    if (updateURL) pushReaderLocation(book);
    fetch(`/bible/list_chapters/${encodeURIComponent(book)}?json=true`)
      .then((res) => res.json())
      .then((data: ChaptersResponse) => setChapters(data.Chapters || []))
      .catch(() => setChapters([]));
  };

  // Record a read page to the server and reflect the returned list locally.
  // No-op when signed out so anonymous activity is never retained.
  const recordRead = (book: string, chapter: number) => {
    if (!hasToken) return;
    fetch("/user/history", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ book, chapter }),
    })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => { if (Array.isArray(data?.pages)) setRecent(data.pages); })
      .catch(() => {});
  };

  const recordSearch = (searchQuery: string) => {
    if (!hasToken) return;
    fetch("/user/history/searches", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ query: searchQuery }),
    })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => { if (Array.isArray(data?.searches)) setSearchHistory(data.searches); })
      .catch(() => {});
  };

  const loadVerses = (book: string, chapter: number) => {
    fetch(`/bible/${encodeURIComponent(book)}/${chapter}?json=true`)
      .then((res) => res.json())
      .then((data: ChapterResponse) => {
        setReading((data.Verses || []).map((text, i) => ({ number: i + 1, text })));
      })
      .catch(() => setReading([]));
    recordRead(book, chapter);
  };

  const loadChapter = (chapter: number) => {
    setSelectedChapter(chapter);
    setSelectedVerses([]);
    setSelectedVerseLabel("");
    setView("reading");
    pushReaderLocation(selectedBook, chapter);
    loadVerses(selectedBook, chapter);
  };

  // Jump straight to a (book, chapter) — used by the "Continue reading" list,
  // which may target a book whose chapters aren't loaded yet.
  const openPage = (
    book: string,
    chapter: number,
    selection: VerseSelection = { verses: [], label: "" },
    updateURL = true,
  ) => {
    setSelectedBook(book);
    setSelectedChapter(chapter);
    setSelectedVerses(selection.verses);
    setSelectedVerseLabel(selection.label);
    setView("reading");
    if (updateURL) pushReaderLocation(book, chapter, selection.label);
    fetch(`/bible/list_chapters/${encodeURIComponent(book)}?json=true`)
      .then((res) => res.json())
      .then((data: ChaptersResponse) => setChapters(data.Chapters || []))
      .catch(() => setChapters([]));
    loadVerses(book, chapter);
  };

  // Restore a book, chapter, and optional verse selection from the URL on
  // startup and whenever the user navigates with the browser controls.
  useEffect(() => {
    const restoreLocation = () => {
      const params = new URLSearchParams(window.location.search);
      const book = params.get("book")?.trim();
      const chapter = positiveIntParam(params.get("chapter"));
      const selection = parseVerseSelection(params.get("verse") || params.get("verses"));

      if (!book) {
        showBooks(false);
      } else if (chapter > 0) {
        openPage(book, chapter, selection, false);
      } else {
        loadChapters(book, false);
      }
    };

    restoreLocation();
    window.addEventListener("popstate", restoreLocation);
    return () => window.removeEventListener("popstate", restoreLocation);
  }, []);

  const runSearch = async (searchTerm = query) => {
    const normalizedQuery = searchTerm.trim();
    if (!normalizedQuery) return;
    setLoading(true);
    setError("");
    try {
      const resp = await fetch(`/bible/v2/search?q=${encodeURIComponent(normalizedQuery)}`);
      if (!resp.ok) throw new Error("Search failed");
      const data = await resp.json();
      setResults(data?.data?.results || []);
      setBookCounts(data?.data?.book_counts || {});
      setView("search");
      recordSearch(normalizedQuery);
    } catch {
      setError("Search failed. Check the API.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const query = suggestQuery.trim();
    const requestID = ++suggestionRequestRef.current;

    if (!query) {
      setSuggestions([]);
      setSuggestionOffset(0);
      setHasMoreSuggestions(false);
      setSuggestionsLoading(false);
      return;
    }

    setSuggestions([]);
    setSuggestionOffset(0);
    setHasMoreSuggestions(false);
    const timer = window.setTimeout(async () => {
      setSuggestionsLoading(true);
      try {
        const resp = await fetch(`/bible/v2/suggest?q=${encodeURIComponent(query)}&n=${suggestionPageSize}&from=0`);
        if (!resp.ok) return;
        const data = await resp.json();
        if (requestID !== suggestionRequestRef.current) return;
        const nextSuggestions = data?.data?.suggestions || [];
        setSuggestions(nextSuggestions);
        setSuggestionOffset(nextSuggestions.length);
        setHasMoreSuggestions(nextSuggestions.length === suggestionPageSize);
      } catch {
        if (requestID === suggestionRequestRef.current) {
          setSuggestions([]);
          setHasMoreSuggestions(false);
        }
      } finally {
        if (requestID === suggestionRequestRef.current) setSuggestionsLoading(false);
      }
    }, 200);
    return () => window.clearTimeout(timer);
  }, [suggestQuery]);

  useEffect(() => {
    const sentinel = suggestionsSentinelRef.current;
    const query = suggestQuery.trim();
    if (!sentinel || !query || !hasMoreSuggestions || suggestionsLoading) return;

    const observer = new IntersectionObserver((entries) => {
      if (!entries[0].isIntersecting) return;
      observer.unobserve(sentinel);
      const requestID = suggestionRequestRef.current;
      const from = suggestionOffset;
      setSuggestionsLoading(true);

      fetch(`/bible/v2/suggest?q=${encodeURIComponent(query)}&n=${suggestionPageSize}&from=${from}`)
        .then((resp) => (resp.ok ? resp.json() : Promise.reject(new Error("Suggestion request failed"))))
        .then((data) => {
          if (requestID !== suggestionRequestRef.current) return;
          const nextSuggestions: Suggestion[] = data?.data?.suggestions || [];
          setSuggestions((current) => {
            const known = new Set(current.map((s) => `${s.book}-${s.chapter}-${s.verse}-${s.text}`));
            return [...current, ...nextSuggestions.filter((s) => !known.has(`${s.book}-${s.chapter}-${s.verse}-${s.text}`))];
          });
          setSuggestionOffset(from + nextSuggestions.length);
          setHasMoreSuggestions(nextSuggestions.length === suggestionPageSize);
        })
        .catch(() => {
          if (requestID === suggestionRequestRef.current) setHasMoreSuggestions(false);
        })
        .finally(() => {
          if (requestID === suggestionRequestRef.current) setSuggestionsLoading(false);
        });
    }, { rootMargin: "240px" });

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasMoreSuggestions, suggestionOffset, suggestionsLoading, suggestQuery]);

  const openSuggestion = (suggestion?: Suggestion) => {
    if (!suggestion?.book || !suggestion.chapter || !suggestion.verse) return;
    const url = v2ReaderURL(suggestion.book, suggestion.chapter, suggestion.verse);
    if (verseOpenMode === "new-tab") {
      window.open(url, "_blank", "noopener,noreferrer");
      return;
    }
    window.location.href = url;
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
          <button onClick={() => { showBooks(); setMenuOpen(false); }}>Books</button>
          <button onClick={() => { setView("search"); searchInputRef.current?.focus(); setMenuOpen(false); }}>Search</button>
          <a href="/docs">Docs</a>
          <button onClick={() => setSettingsOpen(!settingsOpen)}>Settings</button>
          <a href="/bible/list_books">Open Classic</a>
        </div>
      )}

      {/* Settings panel */}
      {settingsOpen && (
        <div className="settings-panel open" ref={settingsRef}>
          {hasToken && (
            <div className="settings-tabs" role="tablist" aria-label="Settings sections">
              <button role="tab" aria-selected={settingsTab === "preferences"} onClick={() => setSettingsTab("preferences")}>Preferences</button>
              <button role="tab" aria-selected={settingsTab === "history"} onClick={() => setSettingsTab("history")}>History</button>
            </div>
          )}
          {settingsTab === "preferences" && (
            <>
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
              <strong>Verse links</strong>
              {[
                { value: "new-tab", label: "Open in new tab" },
                { value: "same-tab", label: "Use current tab" },
              ].map((opt) => (
                <label key={opt.value}>
                  <input
                    type="radio"
                    name="verse-open-mode"
                    value={opt.value}
                    checked={verseOpenMode === opt.value}
                    onChange={() => setVerseOpenMode(opt.value as VerseOpenMode)}
                  />
                  {opt.label}
                </label>
              ))}
            </>
          )}
          {hasToken && settingsTab === "history" && (
            <div className="history-tab" role="tabpanel">
              <strong>Searches</strong>
              {searchHistory.length === 0 && <p className="history-empty">No searches yet.</p>}
              {searchHistory.map((entry) => (
                <button
                  key={`${entry.query}-${entry.ts}`}
                  className="history-entry"
                  onClick={() => {
                    setQuery(entry.query);
                    runSearch(entry.query);
                    setSettingsOpen(false);
                  }}
                >
                  <span>{entry.query}</span>
                  <time>{formatHistoryTime(entry.ts)}</time>
                </button>
              ))}
              <strong>Reading and navigation</strong>
              {recent.length === 0 && <p className="history-empty">No pages read yet.</p>}
              {recent.map((page) => (
                <button
                  key={`${page.book}-${page.chapter}-${page.ts}`}
                  className="history-entry"
                  onClick={() => {
                    openPage(page.book, page.chapter);
                    setSettingsOpen(false);
                  }}
                >
                  <span>{BOOK_ABBREVIATIONS[page.book] ?? page.book} {page.chapter}</span>
                  <time>{formatHistoryTime(page.ts)}</time>
                </button>
              ))}
            </div>
          )}
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
        <>
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
          {hasMoreSuggestions && <div className="suggestions-sentinel" ref={suggestionsSentinelRef} aria-busy={suggestionsLoading} />}
        </>
      )}
      {error && <div className="error">{error}</div>}

      {/* Navigation */}
      {view !== "books" && (
        <div className="nav-bar">
          <button className="nav-btn" onClick={() => showBooks()}>Books Menu</button>
          {view === "reading" && (
            <button className="nav-btn" onClick={() => loadChapters(selectedBook)}>{selectedBook}</button>
          )}
          {view === "reading" && selectedChapter > 1 && (
            <button className="nav-btn" onClick={() => loadChapter(selectedChapter - 1)}>&lt;</button>
          )}
          {view === "reading" && selectedChapter < chapters.length && (
            <button className="nav-btn" onClick={() => loadChapter(selectedChapter + 1)}>&gt;</button>
          )}
        </div>
      )}

      {/* Continue reading: last pages read (signed-in users) */}
      {view === "books" && recent.length > 0 && (
        <div className="recent-row">
          <span className="recent-label">Continue reading</span>
          {recent.slice(0, 5).map((p) => (
            <button
              key={`${p.book}-${p.chapter}`}
              className="chip"
              onClick={() => openPage(p.book, p.chapter)}
            >
              {BOOK_ABBREVIATIONS[p.book] ?? p.book} {p.chapter}
            </button>
          ))}
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
          <h2>{selectedBook} {selectedChapter}{selectedVerseLabel ? `:${selectedVerseLabel}` : ""}</h2>
          {reading.map((v) => (
            <p
              key={v.number}
              id={`verse-${v.number}`}
              className={`verse-text${selectedVerseSet.has(v.number) ? " selected-verse" : ""}`}
            >
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
                <a
                  href={v2ReaderURL(item.book, item.chapter, item.verse)}
                  target={verseOpenMode === "new-tab" ? "_blank" : undefined}
                  rel={verseOpenMode === "new-tab" ? "noopener noreferrer" : undefined}
                >
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
