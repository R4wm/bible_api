import { useEffect, useMemo, useRef, useState } from "react";

declare const Chart: any;

const defaultScope = "synonyms:write";
const suggestionPageSize = 20;
const maxNoteLength = 10000;
const defaultSearchResultsPerPage = 50;
const searchResultsPerPageOptions = [10, 25, 50, 100];

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

type UserSettings = {
  font?: string;
  theme?: "light" | "dark";
  verse_open_mode?: VerseOpenMode;
  show_continue_reading?: boolean;
  search_results_per_page?: number;
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

type VerseNote = {
  book: string;
  chapter: number;
  verse: number;
  text: string;
  updated_at?: number;
};

type VerseOpenMode = "new-tab" | "same-tab";

type VerseSelection = {
  verses: number[];
  label: string;
};

type View = "books" | "chapters" | "reading" | "search";
type SettingsTab = "preferences" | "history";
type SearchMatchMode = "any" | "all" | "phrase";

type RunSearchOptions = {
  match?: SearchMatchMode;
  caseSensitive?: boolean;
  updateURL?: boolean;
};
type MapID = "ministry-of-jesus" | "pauls-missionary-journeys" | "kingdoms-of-saul-david-solomon";

const MAPS: Record<MapID, { title: string; description: string; src: string; alt: string; sourceURL: string; credit: string; license: string; width?: number; height?: number }> = {
  "ministry-of-jesus": {
    title: "The Ministry of Jesus",
    description: "A map of locations associated with Jesus’ ministry.",
    src: "/v2/maps/the-ministry-of-jesus.svg",
    alt: "Map of locations associated with the ministry of Jesus",
    sourceURL: "https://commons.wikimedia.org/wiki/File:The_Ministry_of_Jesus.svg",
    credit: "Map by DEGA MD",
    license: "CC BY-NC-SA 4.0",
  },
  "pauls-missionary-journeys": {
    title: "Paul’s Missionary Journeys",
    description: "A map of Paul’s three missionary journeys and his journey to Rome.",
    src: "/v2/maps/pauls-missionary-journeys.png",
    alt: "English map of Paul’s three missionary journeys and his journey to Rome",
    sourceURL: "https://commons.wikimedia.org/wiki/File:Biblica_Open_Bible_Map_16_17_Paul_missionary_journeys_map.png",
    credit: "Map by Biblica, Inc. and Biblica Open Study Bible Resources",
    license: "CC BY-SA 4.0",
  },
  "kingdoms-of-saul-david-solomon": {
    title: "The Kingdoms of Saul, David, and Solomon",
    description: "A map of the united monarchy, including the kingdom of Saul and places central to 1 Samuel.",
    src: "/v2/maps/kingdoms-of-saul-david-solomon.webp",
    alt: "Map of the kingdoms of Saul, David, and Solomon",
    sourceURL: "https://commons.wikimedia.org/wiki/File:Biblica_Open_Bible_Map_06_The_Kingdoms_of_Saul_David_and_Solomon.png",
    credit: "Map by Biblica, Inc. and Biblica Open Study Bible Resources",
    license: "CC BY-SA 4.0",
    width: 1600,
    height: 2374,
  },
};

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

const parseSearchMatch = (value: string | null): SearchMatchMode => {
  if (value === "all" || value === "phrase") return value;
  return "any";
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

const v2SearchURL = (
  q: string,
  match: SearchMatchMode,
  caseSensitive: boolean,
  from = 0,
): string => {
  const params = new URLSearchParams({
    q,
    match,
    from: String(from),
  });
  if (caseSensitive) params.set("case_sensitive", "true");
  return `/v2/?${params.toString()}`;
};

export default function App() {
  const [token, setToken] = useState("");
  const [config, setConfig] = useState<AuthConfig>({});

  const [view, setView] = useState<View>("books");
  const [selectedMapID, setSelectedMapID] = useState<MapID | "">("");
  const [books, setBooks] = useState<string[]>([]);
  const [selectedBook, setSelectedBook] = useState("");
  const [chapters, setChapters] = useState<number[]>([]);
  const [selectedChapter, setSelectedChapter] = useState(0);
  const [selectedVerses, setSelectedVerses] = useState<number[]>([]);
  const [selectedVerseLabel, setSelectedVerseLabel] = useState("");
  const [reading, setReading] = useState<ReadVerse[]>([]);

  const [query, setQuery] = useState("");
  const [searchMatch, setSearchMatch] = useState<SearchMatchMode>("any");
  const [caseSensitive, setCaseSensitive] = useState(false);
  const [suggestQuery, setSuggestQuery] = useState("");
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [suggestionOffset, setSuggestionOffset] = useState(0);
  const [suggestionsLoading, setSuggestionsLoading] = useState(false);
  const [hasMoreSuggestions, setHasMoreSuggestions] = useState(false);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [bookCounts, setBookCounts] = useState<Record<string, number>>({});
  const [searchTotal, setSearchTotal] = useState(0);
  const [searchOffset, setSearchOffset] = useState(0);
  const [searchResultsPerPage, setSearchResultsPerPage] = useState(defaultSearchResultsPerPage);
  const [searchPerformed, setSearchPerformed] = useState(false);
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
  const [showContinueReading, setShowContinueReading] = useState(true);
  const [recent, setRecent] = useState<RecentPage[]>([]);
  const [searchHistory, setSearchHistory] = useState<SearchHistory[]>([]);
  const [notes, setNotes] = useState<VerseNote[]>([]);
  const [noteVerse, setNoteVerse] = useState<number | null>(null);
  const [noteText, setNoteText] = useState("");
  const [noteSaving, setNoteSaving] = useState(false);
  const [noteError, setNoteError] = useState("");
  const [notesMode, setNotesMode] = useState(false);
  const [notesEditingEnabled, setNotesEditingEnabled] = useState(true);
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
  const runSearchRef = useRef<
    (searchTerm?: string, from?: number, pageSize?: number, options?: RunSearchOptions) => Promise<void>
  >(async () => {});
  const [linkCopied, setLinkCopied] = useState(false);
  const searchResultsRef = useRef<HTMLElement>(null);
  const chaptersRef = useRef<HTMLElement>(null);
  const readingRef = useRef<HTMLElement>(null);
  const mapsRef = useRef<HTMLElement>(null);

  const fontClasses = ["font-blackletter", "font-renaissance", "font-serif"];
  const selectedVerseSet = useMemo(() => new Set(selectedVerses), [selectedVerses]);
  const notesByVerse = useMemo(
    () => new Map(notes.map((note) => [note.verse, note])),
    [notes],
  );

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
      setNotes([]);
      setNoteVerse(null);
      setNotesMode(false);
      setNotesEditingEnabled(true);
      setSettingsTab("preferences");
      setSearchResultsPerPage(defaultSearchResultsPerPage);
      return;
    }
    fetch("/user/settings")
      .then((res) => (res.ok ? res.json() : null))
      .then((data: UserSettings | null) => {
        if (data?.font) setFont(data.font);
        if (data?.theme === "light" || data?.theme === "dark") setTheme(data.theme);
        if (isVerseOpenMode(data?.verse_open_mode)) setVerseOpenMode(data.verse_open_mode);
        if (typeof data?.show_continue_reading === "boolean") setShowContinueReading(data.show_continue_reading);
        if (searchResultsPerPageOptions.includes(data?.search_results_per_page || 0)) {
          setSearchResultsPerPage(data!.search_results_per_page!);
        }
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

  // Notes are loaded only for the open chapter and only after sign-in.
  useEffect(() => {
    if (!hasToken || view !== "reading" || !selectedBook || !selectedChapter) {
      setNotes([]);
      setNoteVerse(null);
      return;
    }
    let cancelled = false;
    fetch(`/user/notes?book=${encodeURIComponent(selectedBook)}&chapter=${selectedChapter}`)
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (!cancelled) {
          setNotes(Array.isArray(data?.notes) ? data.notes : []);
          const editingEnabled = data?.editing_enabled !== false;
          setNotesEditingEnabled(editingEnabled);
          if (!editingEnabled) setNotesMode(false);
        }
      })
      .catch(() => { if (!cancelled) setNotes([]); });
    return () => { cancelled = true; };
  }, [hasToken, view, selectedBook, selectedChapter]);

  // Push settings changes to the server, but only after hydration so we never
  // overwrite saved preferences with the local defaults on first load.
  useEffect(() => {
    if (!hasToken || !settingsHydrated.current) return;
    fetch("/user/settings", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        font,
        theme,
        verse_open_mode: verseOpenMode,
        show_continue_reading: showContinueReading,
        search_results_per_page: searchResultsPerPage,
      }),
    }).catch(() => {});
  }, [font, theme, verseOpenMode, showContinueReading, searchResultsPerPage, hasToken]);

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
    if (!searchPerformed || !chartRef.current || typeof Chart === "undefined") return;

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
  }, [searchPerformed, bookCounts, theme]);

  const scrollToSection = (section: React.RefObject<HTMLElement | null>) => {
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    section.current?.scrollIntoView({ behavior: reducedMotion ? "auto" : "smooth", block: "start" });
  };

  useEffect(() => {
    if (view !== "search" || loading || !searchPerformed) return;
    const frame = window.requestAnimationFrame(() => scrollToSection(searchResultsRef));
    return () => window.cancelAnimationFrame(frame);
  }, [view, loading, searchPerformed, results]);

  useEffect(() => {
    if (view !== "chapters" || !selectedBook || chapters.length === 0) return;
    const frame = window.requestAnimationFrame(() => scrollToSection(chaptersRef));
    return () => window.cancelAnimationFrame(frame);
  }, [view, selectedBook, chapters]);

  useEffect(() => {
    if (view !== "reading" || !selectedBook || selectedChapter === 0 || reading.length === 0) return;
    const frame = window.requestAnimationFrame(() => scrollToSection(readingRef));
    return () => window.cancelAnimationFrame(frame);
  }, [view, selectedBook, selectedChapter, reading]);

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

  const pushSearchLocation = (
    q: string,
    match: SearchMatchMode,
    caseSensitive: boolean,
    from = 0,
  ) => {
    const url = v2SearchURL(q, match, caseSensitive, from);
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

  const showMaps = () => {
    scrollToSection(mapsRef);
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
      const q = params.get("q")?.trim();
      if (q) {
        const match = parseSearchMatch(params.get("match"));
        const caseSensitive = params.get("case_sensitive") === "true";
        const from = positiveIntParam(params.get("from"));
        setQuery(q);
        setSearchMatch(match);
        setCaseSensitive(caseSensitive);
        void runSearchRef.current(q, from, defaultSearchResultsPerPage, {
          match,
          caseSensitive,
          updateURL: false,
        });
        return;
      }

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

  const runSearch = async (
    searchTerm?: string,
    from = 0,
    pageSize = searchResultsPerPage,
    options: RunSearchOptions = {},
  ) => {
    const normalizedQuery = (searchTerm ?? query).trim();
    if (!normalizedQuery) return;
    const match = options.match ?? searchMatch;
    const sensitive = options.caseSensitive ?? caseSensitive;
    const updateURL = options.updateURL ?? true;
    setLoading(true);
    setError("");
    setSearchPerformed(true);
    try {
      const params = new URLSearchParams({
        q: normalizedQuery,
        match,
        n: String(pageSize),
        from: String(from),
      });
      if (sensitive) params.set("case_sensitive", "true");
      const resp = await fetch(`/bible/v2/search?${params.toString()}`);
      if (!resp.ok) throw new Error("Search failed");
      const data = await resp.json();
      setResults(data?.data?.results || []);
      setBookCounts(data?.data?.book_counts || {});
      setSearchTotal(Number(data?.meta?.count) || 0);
      setSearchOffset(Number(data?.meta?.from) || 0);
      setQuery(normalizedQuery);
      setSearchMatch(match);
      setCaseSensitive(sensitive);
      setView("search");
      if (updateURL) pushSearchLocation(normalizedQuery, match, sensitive, from);
      if (from === 0) recordSearch(normalizedQuery);
    } catch {
      setError("Search failed. Check the API.");
    } finally {
      setLoading(false);
    }
  };

  runSearchRef.current = runSearch;

  const copySearchLink = async () => {
    const url = `${window.location.origin}${v2SearchURL(query, searchMatch, caseSensitive, searchOffset)}`;
    try {
      await navigator.clipboard.writeText(url);
      setLinkCopied(true);
      window.setTimeout(() => setLinkCopied(false), 2000);
    } catch {
      setError("Could not copy link.");
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
    openPage(suggestion.book, suggestion.chapter, {
      verses: [suggestion.verse],
      label: String(suggestion.verse),
    });
  };

  const logout = async () => {
    await fetch("/auth/logout", { method: "POST" });
    setToken("");
    setProfile({});
    localStorage.removeItem("bible-profile");
    setProfileOpen(false);
  };

  const openNote = (verse: number) => {
    if (!hasToken) return;
    setNoteVerse(verse);
    setNoteText(notesByVerse.get(verse)?.text || "");
    setNoteError("");
  };

  const saveNote = async () => {
    if (!noteVerse || !selectedBook || !selectedChapter) return;
    const text = noteText.trim();
    if (!text) {
      setNoteError("Write a note before saving.");
      return;
    }
    setNoteSaving(true);
    setNoteError("");
    try {
      const response = await fetch("/user/notes", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ book: selectedBook, chapter: selectedChapter, verse: noteVerse, text }),
      });
      if (!response.ok) {
        if (response.status === 503) {
          setNotesEditingEnabled(false);
          setNotesMode(false);
          setNoteError("Note editing is temporarily unavailable while storage is near capacity.");
          return;
        }
        throw new Error("Unable to save note");
      }
      const saved: VerseNote = await response.json();
      setNotes((current) => [...current.filter((note) => note.verse !== saved.verse), saved]);
      setNoteText(saved.text);
      setNoteVerse(null);
    } catch {
      setNoteError("Unable to save your note. Please try again.");
    } finally {
      setNoteSaving(false);
    }
  };

  const safeHighlight = useMemo(() => {
    return (item: SearchResult) => {
      if (!item.highlight || item.highlight.length === 0) return item.text;
      return item.highlight[0].replace(/<em>/g, "<mark>").replace(/<\/em>/g, "</mark>");
    };
  }, []);

  const updateSearchResultsPerPage = (value: number) => {
    setSearchResultsPerPage(value);
    if (searchPerformed && query.trim()) runSearch(query, 0, value);
  };

  const continueReading = hasToken && showContinueReading && (
    <div className="recent-row">
      <span className="recent-label">Continue reading</span>
      {recent.length === 0 ? (
        <span className="recent-empty">No pages read yet.</span>
      ) : (
        recent.slice(0, 5).map((page) => (
          <button
            key={`${page.book}-${page.chapter}`}
            className="chip"
            onClick={() => openPage(page.book, page.chapter)}
          >
            {BOOK_ABBREVIATIONS[page.book] ?? page.book} {page.chapter}
          </button>
        ))
      )}
    </div>
  );

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
        <div className="search-controls">
          <div className="search-bar">
          <input
            ref={searchInputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") runSearch(); }}
            placeholder="Search the text"
          />
          <button onClick={() => runSearch()} disabled={loading}>
            {loading ? "..." : "search"}
          </button>
          </div>
          <fieldset className="search-options">
            <legend>Match</legend>
            <label>
              <input type="radio" name="search-match" value="any" checked={searchMatch === "any"} onChange={() => setSearchMatch("any")} />
              Any words
            </label>
            <label>
              <input type="radio" name="search-match" value="all" checked={searchMatch === "all"} onChange={() => setSearchMatch("all")} />
              All words
            </label>
            <label>
              <input type="radio" name="search-match" value="phrase" checked={searchMatch === "phrase"} onChange={() => setSearchMatch("phrase")} />
              Exact phrase
            </label>
            <label>
              <input type="checkbox" checked={caseSensitive} onChange={(event) => setCaseSensitive(event.target.checked)} />
              Case sensitive
            </label>
          </fieldset>
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
          <button onClick={() => { showMaps(); setMenuOpen(false); }}>Maps</button>
          <a href="/docs">Docs</a>
          <a href="/donate">Donations</a>
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
              {hasToken && (
                <>
                  <label className="settings-toggle">
                    <input
                      type="checkbox"
                      checked={showContinueReading}
                      onChange={(event) => setShowContinueReading(event.target.checked)}
                    />
                    Show Continue Reading
                  </label>
                  <label className="settings-select">
                    Search results per page
                    <select
                      value={searchResultsPerPage}
                      onChange={(event) => updateSearchResultsPerPage(Number(event.target.value))}
                    >
                      {searchResultsPerPageOptions.map((count) => (
                        <option key={count} value={count}>{count}</option>
                      ))}
                    </select>
                  </label>
                </>
              )}
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

      {/* Keep search feedback next to the controls that produced it. */}
      {searchPerformed && (
        <section className="search-results" ref={searchResultsRef} aria-live="polite">
          <div className="search-results-header">
            <h2>Search Results</h2>
            <button type="button" className="copy-link-btn" onClick={() => void copySearchLink()}>
              {linkCopied ? "Link copied" : "Copy link"}
            </button>
          </div>
          <p className="search-result-summary">
            {searchTotal === 0
              ? "No matches"
              : `Showing ${searchOffset + 1}–${Math.min(searchOffset + results.length, searchTotal)} of ${searchTotal.toLocaleString()} matches`}
          </p>
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
                  onClick={(event) => {
                    if (verseOpenMode === "same-tab") {
                      event.preventDefault();
                      openPage(item.book, item.chapter, { verses: [item.verse], label: String(item.verse) });
                    }
                  }}
                >
                  {item.book} {item.chapter}:{item.verse}
                </a>
              </div>
              <div className="text" dangerouslySetInnerHTML={{ __html: safeHighlight(item) }} />
            </div>
          ))}
          {searchTotal > searchResultsPerPage && (
            <nav className="search-pagination" aria-label="Search result pages">
              <button
                onClick={() => runSearch(query, Math.max(0, searchOffset - searchResultsPerPage))}
                disabled={loading || searchOffset === 0}
              >
                Previous
              </button>
              <span>
                Page {Math.floor(searchOffset / searchResultsPerPage) + 1} of {Math.ceil(searchTotal / searchResultsPerPage)}
              </span>
              <button
                onClick={() => runSearch(query, searchOffset + searchResultsPerPage)}
                disabled={loading || searchOffset + searchResultsPerPage >= searchTotal}
              >
                Next
              </button>
            </nav>
          )}
        </section>
      )}

      {/* Navigation */}
      {selectedBook && (
        <div className="nav-bar">
          <button className="nav-btn" onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}>Books Menu</button>
          {selectedChapter > 0 && (
            <button className="nav-btn" onClick={() => loadChapters(selectedBook)}>{selectedBook}</button>
          )}
          {selectedChapter > 1 && (
            <button className="nav-btn" onClick={() => loadChapter(selectedChapter - 1)}>&lt;</button>
          )}
          {selectedChapter < chapters.length && (
            <button className="nav-btn" onClick={() => loadChapter(selectedChapter + 1)}>&gt;</button>
          )}
        </div>
      )}

      {/* Books list */}
      <>
        {continueReading}
        <h2>Books</h2>
        <div className="books-grid">
          {books.map((book) => (
            <button key={book} className="block" onClick={() => loadChapters(book)}>
              {book}
            </button>
          ))}
        </div>
      </>

      {/* Chapters list */}
      {selectedBook && (
        <section ref={chaptersRef} className="reader-section">
          <h2>{selectedBook}</h2>
          {continueReading}
          <div className="chapters-grid">
            {chapters.map((ch) => (
              <button key={ch} className="block" onClick={() => loadChapter(ch)}>
                {ch}
              </button>
            ))}
          </div>
        </section>
      )}

      {/* Reading */}
      {selectedChapter > 0 && (
        <section ref={readingRef} className="reader-section reader-layout">
          <div className="reader-content">
            <div className="reader-heading">
              <h2>{selectedBook} {selectedChapter}{selectedVerseLabel ? `:${selectedVerseLabel}` : ""}</h2>
              {hasToken && (
                <button
                  className={`notes-mode-toggle${notesMode ? " active" : ""}`}
                  aria-pressed={notesMode}
                  disabled={!notesEditingEnabled}
                  onClick={() => {
                    setNotesMode((enabled) => {
                      if (enabled) setNoteVerse(null);
                      return !enabled;
                    });
                  }}
                >
                  {notesMode ? "Exit notes mode" : "Notes mode"}
                </button>
              )}
            </div>
            {hasToken && !notesEditingEnabled && <p className="notes-unavailable">Note editing is temporarily unavailable while storage is near capacity.</p>}
            {hasToken && notesMode && <p className="notes-hint">Notes mode is on — select a verse to add or edit a private note.</p>}
            {reading.map((v) => (
              <div
                key={v.number}
                className="verse-row"
              >
                <p
                  id={`verse-${v.number}`}
                  className={`verse-text${selectedVerseSet.has(v.number) ? " selected-verse" : ""}${notesMode ? " note-selectable" : ""}`}
                  onClick={notesMode ? () => openNote(v.number) : undefined}
                >
                  <span className="verse-num">{v.number}</span> {v.text}
                </p>
                {notesByVerse.has(v.number) && (
                  <button
                    className="note-indicator"
                    aria-label={`Open note for verse ${v.number}`}
                    title="Open note"
                    onClick={(event) => { event.stopPropagation(); openNote(v.number); }}
                  />
                )}
              </div>
            ))}
          </div>
          {hasToken && noteVerse !== null && (
            <div
              className="notes-modal-backdrop"
              onClick={(event) => { if (event.target === event.currentTarget) setNoteVerse(null); }}
            >
              <aside className="notes-sidebar" role="dialog" aria-modal="true" aria-label={`Note for verse ${noteVerse}`}>
                <div className="notes-sidebar-heading">
                  <strong>Note · {BOOK_ABBREVIATIONS[selectedBook] ?? selectedBook} {selectedChapter}:{noteVerse}</strong>
                  <button className="notes-close" onClick={() => setNoteVerse(null)} aria-label="Close note">×</button>
                </div>
                {notesMode ? (
                  <>
                    <textarea
                      value={noteText}
                      onChange={(event) => setNoteText(event.target.value)}
                      placeholder="Write your note…"
                      maxLength={maxNoteLength}
                      autoFocus
                    />
                    <p className="note-character-count">{noteText.length.toLocaleString()} / {maxNoteLength.toLocaleString()}</p>
                    {noteError && <p className="note-error">{noteError}</p>}
                    <button className="note-save" onClick={saveNote} disabled={noteSaving}>
                      {noteSaving ? "Saving…" : "Save note"}
                    </button>
                  </>
                ) : (
                  <>
                    <p className="note-display">{noteText}</p>
                    {noteError && <p className="note-error">{noteError}</p>}
                  </>
                )}
              </aside>
            </div>
          )}
        </section>
      )}

      <section className="maps-view" ref={mapsRef} aria-labelledby="maps-heading">
        <h2 id="maps-heading">Maps</h2>
        <label className="maps-selector">
          Choose a map
          <select value={selectedMapID} onChange={(event) => setSelectedMapID(event.target.value as MapID | "")}>
            <option value="">Select a map</option>
            {Object.entries(MAPS).map(([id, map]) => <option key={id} value={id}>{map.title}</option>)}
          </select>
        </label>
        {selectedMapID && (
          <article className="map-card">
            <h3>{MAPS[selectedMapID].title}</h3>
            <p>{MAPS[selectedMapID].description}</p>
            <a href={MAPS[selectedMapID].src} target="_blank" rel="noopener noreferrer">Open full-size map</a>
            <p className="map-attribution">
              {MAPS[selectedMapID].credit} via{" "}
              <a href={MAPS[selectedMapID].sourceURL} target="_blank" rel="noopener noreferrer">Wikimedia Commons</a>{" "}
              ({MAPS[selectedMapID].license}).
            </p>
            <img
              src={MAPS[selectedMapID].src}
              alt={MAPS[selectedMapID].alt}
              width={MAPS[selectedMapID].width}
              height={MAPS[selectedMapID].height}
              loading="lazy"
              decoding="async"
            />
          </article>
        )}
      </section>

      <div className="footer">
        <span>Powered by OpenSearch</span> | <span>API docs: <a href="/docs">/docs</a></span>
      </div>
    </div>
  );
}
