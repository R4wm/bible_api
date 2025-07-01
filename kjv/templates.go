package kjv

const (
	booksButtonsTemplate = `
<!DOCTYPE html>
<html>
   <head>
      <meta name="viewport" content="width=device-width, initial-scale=1">
      <style>
	 .block {
	 display: block;
	 width: 100%;
	 border: none;
	 background-color: #4CAF50;
	 color: white;
	 padding: 14px 28px;
	 font-size: 16px;
	 cursor: pointer;
	 text-align: center;
	 margin-bottom: 5px;
	 }
	 .block:hover {
	 background-color: #ddd;
	 color: black;
	 }
	 .search-button {
	 display: block;
	 width: 100%;
	 border: none;
	 background-color: #007bff;
	 color: white;
	 padding: 16px 28px;
	 font-size: 18px;
	 cursor: pointer;
	 text-align: center;
	 margin-bottom: 20px;
	 border-radius: 5px;
	 font-weight: bold;
	 }
	 .search-button:hover {
	 background-color: #0056b3;
	 }
	 .nav-button {
	 display: inline-block;
	 width: 48%;
	 border: none;
	 background-color: #6c757d;
	 color: white;
	 padding: 12px 20px;
	 font-size: 14px;
	 cursor: pointer;
	 text-align: center;
	 margin: 1%;
	 border-radius: 3px;
	 text-decoration: none;
	 }
	 .nav-button:hover {
	 background-color: #545b62;
	 }
      </style>
      <title>Books of the Bible</title>
   </head>
   <body style="background-color:{{ .Color }};">
      <div style="padding: 20px;">
         <h1 style="text-align: center; margin-bottom: 30px;">Books of the Bible</h1>
         
         <button class="search-button" onclick="window.location.href='/bible/search';">🔍 Search the Bible</button>
         
         <div style="margin-bottom: 20px;">
            <a href="/bible/random_verse" class="nav-button">🎲 Random Verse</a>
            <a href="/health" class="nav-button">📊 Health Check</a>
         </div>
         
         {{ range $key, $value := .Books }}
         <p><button class="block" onclick="window.location.href= '{{ createLink $value }}';" >{{ $value }}</button></p>
         {{ end }}
      </div>
   </body>
</html>
`
	verseTemplate = `
<!DOCTYPE html>
<html>
  <body style="background-color:{{ .Color }};">
    <h1>
      <center>
	<a href={{.ChapterRef}}>{{ .Verse.Book }} {{ .Verse.Chapter }}</a> : {{ .Verse.Verse }}
      </center>
    </h1>
    <h3>
      <center>{{ .Verse.Text }}</center>
    </h3>
    <center>
      <button onclick="window.location.href = '{{.ListAllBooksLink}}';" class="w3-bar-item w3-button" style="width:33.3%">Books</button>
    </center>
  </body>
</html>
`
	chapterTemplate = `
<html>
<title>{{.BookName}} {{ .Chapter}}</title>
<style>
.btn-group button {
  background-color: gold; /* Green background */
  border: 1px solid green; /* Green border */
  color: black;
  padding: 10px 24px; /* Some padding */
  cursor: pointer; /* Pointer/hand icon */
  float: center; /* Float the buttons side by side */
}
/* Clear floats (clearfix hack) */
.btn-group:after {
  content: "";
  clear: both;
  display: table;
}
.btn-group button:not(:last-child) {
  border-right: none; /* Prevent double borders */
}
/* Add a background color on hover */
.btn-group button:hover {
  background-color: #3e8e41;
}
</style>
  <body style="background-color:{{ .Color }};">
    <h1><center><a href=../{{.BookName}}>{{ .BookName }}</a> {{ .Chapter }}</h1>
  <body>
    {{ range $index, $results := .Verses }}
    <p><b><left><a href={{ verseLink $index }}> {{ add $index 1}}</a> {{ . }} </b></p>
    {{ end }}
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
    <div class="w3-bar">
    <div class="btn-group">
    {{ if .PreviousChapterLink  }}
    <button onclick="window.location.href = '{{.PreviousChapterLink}}';" class="w3-bar-item w3-button" style="width:33.3%"> < </button>
    {{ end }}
    <button onclick="window.location.href = '{{.ListAllBooksLink}}';" class="w3-bar-item w3-button" style="width:33.3%">Books</button>
    {{ if .NextChapterLink  }}
    <button onclick="window.location.href = '{{.NextChapterLink}}';" class="w3-bar-item w3-button" style="width:33.3%"> > </button>
    {{ end }}
    </div>
  </body>
</html>
`

	versesTemplate = `
<html>
<title>{{.HTMLTitle}}</title>
  <style>
    .btn-group button {
    background-color: gold; /* Green background */
    border: 1px solid green; /* Green border */
    color: black;
    padding: 10px 24px; /* Some padding */
    cursor: pointer; /* Pointer/hand icon */
    float: center; /* Float the buttons side by side */
    }
    /* Clear floats (clearfix hack) */
    .btn-group:after {
    content: "";
    clear: both;
    display: table;
    }
    .btn-group button:not(:last-child) {
    border-right: none; /* Prevent double borders */
    }
    /* Add a background color on hover */
    .btn-group button:hover {
    background-color: #3e8e41;
    }
  </style>
  <body style="background-color:{{ .Color }};">
    {{if .StartVerse}}<h1><center><a href=../{{.Chapter}}>{{ .BookName }} {{ .Chapter }}</a>:{{.StartVerse}}-{{.EndVerse}}</h1>
    {{else}}
    <h1><center><a href="../{{.Chapter}}">{{ .BookName }} {{ .Chapter }}</a>:{{.SingleVerse}}
    {{end}}
	<body>
	  {{ range $index, $results := .Verses }}
	  {{ range $verseNum, $verseText := $results}}
	  <p><b><left><a href={{$verseNum}}?json=false> {{$verseNum}}</a> {{$verseText }} </b></p>
	  {{end}}
	  {{end}}
	  <meta name="viewport" content="width=device-width, initial-scale=1">
	  <link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
	  <div class="w3-bar">
	    <div class="btn-group">
	      {{ if .PreviousChapterLink  }}
	      <button onclick="window.location.href = '{{.PreviousChapterLink}}';" class="w3-bar-item w3-button" style="width:33.3%"> < </button>
																	{{ end }}
																	<button onclick="window.location.href = '{{.ListAllBooksLink}}';" class="w3-bar-item w3-button" style="width:33.3%">Books</button>
																	{{ if .NextChapterLink  }}
																	<button onclick="window.location.href = '{{.NextChapterLink}}';" class="w3-bar-item w3-button" style="width:33.3%"> > </button>
	      {{ end }}
	    </div>
	</body>
</html>
	`

	chapterButtonsTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
      .block {
      display: block;
      width: 100%;
      border: none;
      background-color: #4CAF50;
      color: white;
      padding: 14px 28px;
      font-size: 16px;
      cursor: pointer;
      text-align: center;
      }
      .block:hover {
      background-color: #ddd;
      color: black;
      }
    </style>
    <title>{{ .Name }}</title>
  </head>
  <body style="background-color:{{ .Color }};">
    <p><center><h1> {{ .Name }} </h1><center></p>
    <button onclick="window.location.href='../bible/list_books';" class="w3-bar-item w3-button" style="width:33.3%">Books Menu</button>
    {{ range $index, $results := .Links }}
    <p><button class="block" onclick="window.location.href = '{{ $results }}'">{{ add $index 1 }}</button></p>
    {{ end }}
  </body>
</html>
`

	searchResultTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Search Results for "{{.SearchString}}"</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            color: #333;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
            overflow: hidden;
        }
        
        .search-header {
            background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        
        .search-title {
            font-size: 2.2em;
            font-weight: 700;
            margin-bottom: 10px;
            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
        }
        
        .search-query {
            font-size: 1.2em;
            background: rgba(255, 255, 255, 0.2);
            padding: 6px 14px;
            border-radius: 20px;
            display: inline-block;
            margin: 8px 0;
            border: 1px solid rgba(255, 255, 255, 0.3);
        }
        
        .search-stats {
            font-size: 1.1em;
            opacity: 0.9;
            margin-top: 8px;
        }
        
        .search-form {
            background: white;
            padding: 20px 30px;
            display: flex;
            gap: 15px;
            align-items: center;
            flex-wrap: wrap;
            border-bottom: 1px solid rgba(0, 0, 0, 0.1);
        }
        
        .search-input {
            flex: 1;
            min-width: 200px;
            padding: 12px 20px;
            border: 2px solid #e5e7eb;
            border-radius: 25px;
            font-size: 1em;
            transition: all 0.3s ease;
            outline: none;
        }
        
        .search-input:focus {
            border-color: #4f46e5;
            box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
        }
        
        .search-button {
            padding: 12px 24px;
            background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
            color: white;
            border: none;
            border-radius: 25px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
        }
        
        .search-button:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(79, 70, 229, 0.4);
        }
        
        .nav-section {
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        
        .nav-button {
            padding: 10px 20px;
            background: rgba(79, 70, 229, 0.1);
            color: #4f46e5;
            text-decoration: none;
            border-radius: 20px;
            font-weight: 600;
            transition: all 0.3s ease;
            border: 2px solid rgba(79, 70, 229, 0.2);
            font-size: 0.9em;
        }
        
        .nav-button:hover {
            background: #4f46e5;
            color: white;
            transform: translateY(-1px);
        }
        
        .chart-container {
            background: white;
            padding: 30px;
            text-align: center;
            position: relative;
        }
        
        .chart-title {
            font-size: 1.4em;
            font-weight: 600;
            margin-bottom: 20px;
            color: #374151;
        }
        
        .chart-wrapper {
            position: relative;
            height: 400px;
            width: 100%;
            margin: 0 auto;
        }
        
        #myChart {
            position: absolute;
            top: 0;
            left: 0;
            width: 100% !important;
            height: 100% !important;
        }
        
        .results-container {
            background: white;
            padding: 30px;
        }
        
        .verse-item {
            margin-bottom: 20px;
            padding: 16px 0;
            border-bottom: 1px solid #f3f4f6;
            text-align: center;
        }
        
        .verse-item:last-child {
            border-bottom: none;
        }
        
        .verse-reference {
            margin-bottom: 8px;
        }
        
        .verse-reference a {
            color: #4f46e5;
            text-decoration: none;
            font-weight: 600;
            font-size: 1.1em;
            transition: color 0.3s ease;
        }
        
        .verse-reference a:hover {
            color: #7c3aed;
            text-decoration: underline;
        }
        
        .verse-text {
            font-size: 1em;
            line-height: 1.6;
            color: #374151;
            max-width: 800px;
            margin: 0 auto;
        }
        
        .json-breakdown {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
        
        .json-breakdown h4 {
            margin: 0 0 10px 0;
            color: #374151;
            font-size: 1em;
            font-weight: 600;
        }
        
        .json-breakdown pre {
            margin: 0;
            background: white;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 12px;
            font-size: 0.9em;
            line-height: 1.6;
            white-space: pre-wrap;
            word-wrap: break-word;
        }
        
        .json-breakdown code {
            color: #495057;
            font-family: inherit;
        }
        
        .no-results {
            text-align: center;
            padding: 60px 30px;
            color: #6b7280;
            background: white;
        }
        
        .no-results-icon {
            font-size: 4em;
            margin-bottom: 20px;
            opacity: 0.5;
        }
        
        @media (max-width: 768px) {
            .search-header {
                padding: 20px;
            }
            
            .search-title {
                font-size: 1.8em;
            }
            
            .chart-container, .results-container {
                padding: 20px;
            }
            
            .search-form {
                padding: 15px 20px;
                flex-direction: column;
            }
            
            .search-input {
                min-width: 100%;
            }
            
            .nav-section {
                justify-content: center;
            }
        }
        
        .autocomplete-dropdown {
            position: absolute;
            top: 100%;
            left: 0;
            right: 0;
            background: white;
            border: 2px solid #e5e7eb;
            border-top: none;
            border-radius: 0 0 15px 15px;
            max-height: 200px;
            overflow-y: auto;
            z-index: 1000;
            display: none;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        
        .autocomplete-item {
            padding: 12px 20px;
            cursor: pointer;
            border-bottom: 1px solid #f3f4f6;
            transition: background-color 0.2s ease;
        }
        
        .autocomplete-item:hover,
        .autocomplete-item.selected {
            background-color: #f8fafc;
            color: #4f46e5;
        }
        
        .autocomplete-item:last-child {
            border-bottom: none;
        }
        
        .search-input-container {
            position: relative;
            flex: 1;
            min-width: 200px;
        }
    </style>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/3.9.1/chart.min.js"></script>
</head>
<body>
    <div class="container">
        <div class="search-header">
            <h1 class="search-title">Bible Search</h1>
            <div class="search-query">"{{.SearchString}}"</div>
            <div class="search-stats">
                Found {{.Count.overall}} verses
            </div>
        </div>
        
        
        <div class="search-form">
            <form method="GET" action="/bible/search" style="display: flex; gap: 15px; flex: 1; align-items: center; flex-wrap: wrap;">
                <div class="search-input-container">
                    <input type="text" name="q" id="search-input-results" class="search-input" placeholder="Search the Bible..." value="{{.SearchString}}" required autocomplete="off">
                    <div id="autocomplete-dropdown-results" class="autocomplete-dropdown"></div>
                </div>
                <button type="submit" class="search-button">🔍 Search</button>
            </form>
            <div class="nav-section">
                <a href="/bible/random_verse" class="nav-button">🎲 Random</a>
                <a href="/bible/list_books" class="nav-button">📚 Books</a>
            </div>
        </div>
        
        {{if .Verses}}
        <div class="chart-container">
            <h3 class="chart-title">Distribution of "{{.SearchString}}" across Bible books</h3>
            <div class="chart-wrapper">
                <canvas id="myChart"></canvas>
            </div>
        </div>
        
        <div class="results-container">
            {{range .Verses}}
            <div class="verse-item">
                <div class="verse-reference">
                    <a href="{{createLink .}}">{{.Book}} {{.Chapter}}:{{.Verse}}</a>
                </div>
                <div class="verse-text">{{.Text}}</div>
            </div>
            {{end}}
        </div>
        {{else}}
        <div class="no-results">
            <div class="no-results-icon">📖</div>
            <h3>No verses found</h3>
            <p>Try a different search term or check your spelling.</p>
        </div>
        {{end}}
        
        {{if .Verses}}
        <div class="json-breakdown">
            <h4>Book breakdown (JSON):</h4>
            <pre><code>{{range $book, $count := .Count}}{{if ne $book "overall"}}"{{$book}}": {{$count}}
{{end}}{{end}}</code></pre>
        </div>
        {{end}}
    </div>

    {{if .Verses}}
    <script>
        const ctx = document.getElementById('myChart').getContext('2d');
        const myChart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: [
                    "Genesis", "Exodus", "Leviticus", "Numbers", "Deuteronomy", 
                    "Joshua", "Judges", "Ruth", "1 Samuel", "2 Samuel", 
                    "1 Kings", "2 Kings", "1 Chronicles", "2 Chronicles", "Ezra", 
                    "Nehemiah", "Esther", "Job", "Psalms", "Proverbs", 
                    "Ecclesiastes", "Song of Solomon", "Isaiah", "Jeremiah", "Lamentations", 
                    "Ezekiel", "Daniel", "Hosea", "Joel", "Amos", 
                    "Obadiah", "Jonah", "Micah", "Nahum", "Habakkuk", 
                    "Zephaniah", "Haggai", "Zechariah", "Malachi", "Matthew", 
                    "Mark", "Luke", "John", "Acts", "Romans", 
                    "1 Corinthians", "2 Corinthians", "Galatians", "Ephesians", "Philippians", 
                    "Colossians", "1 Thessalonians", "2 Thessalonians", "1 Timothy", "2 Timothy", 
                    "Titus", "Philemon", "Hebrews", "James", "1 Peter", 
                    "2 Peter", "1 John", "2 John", "3 John", "Jude", "Revelation"
                ],
                datasets: [{
                    label: 'Verses containing "{{.SearchString}}"',
                    data: {{.GraphCount}},
                    backgroundColor: function(context) {
                        const index = context.dataIndex;
                        // Old Testament (0-38) - Purple shades
                        if (index <= 38) {
                            return 'rgba(79, 70, 229, 0.6)';
                        }
                        // New Testament (39-65) - Blue shades  
                        return 'rgba(59, 130, 246, 0.6)';
                    },
                    borderColor: function(context) {
                        const index = context.dataIndex;
                        if (index <= 38) {
                            return 'rgba(79, 70, 229, 1)';
                        }
                        return 'rgba(59, 130, 246, 1)';
                    },
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                animation: {
                    duration: 0 // Disable animations to prevent scrolling issues
                },
                plugins: {
                    legend: {
                        display: true,
                        position: 'top'
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.1)'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        },
                        ticks: {
                            maxRotation: 45,
                            minRotation: 0
                        }
                    }
                }
            }
        });
    </script>
    {{end}}
    
    <script>
        function initAutocomplete(inputId, dropdownId) {
            const input = document.getElementById(inputId);
            const dropdown = document.getElementById(dropdownId);
            let selectedIndex = -1;
            let suggestions = [];
            let debounceTimer = null;

            function fetchSuggestions(query) {
                if (query.length < 2) {
                    dropdown.style.display = 'none';
                    return;
                }

                fetch('/bible/autocomplete?q=' + encodeURIComponent(query))
                    .then(response => response.json())
                    .then(data => {
                        suggestions = data || [];
                        displaySuggestions();
                    })
                    .catch(error => {
                        console.error('Autocomplete error:', error);
                        dropdown.style.display = 'none';
                    });
            }

            function displaySuggestions() {
                if (suggestions.length === 0) {
                    dropdown.style.display = 'none';
                    return;
                }

                dropdown.innerHTML = '';
                suggestions.forEach((suggestion, index) => {
                    const item = document.createElement('div');
                    item.className = 'autocomplete-item';
                    item.textContent = suggestion;
                    item.addEventListener('click', () => {
                        input.value = suggestion;
                        dropdown.style.display = 'none';
                        input.form.submit();
                    });
                    dropdown.appendChild(item);
                });

                dropdown.style.display = 'block';
                selectedIndex = -1;
            }

            function selectItem(index) {
                const items = dropdown.querySelectorAll('.autocomplete-item');
                items.forEach(item => item.classList.remove('selected'));
                
                if (index >= 0 && index < items.length) {
                    items[index].classList.add('selected');
                    selectedIndex = index;
                }
            }

            input.addEventListener('input', function() {
                clearTimeout(debounceTimer);
                debounceTimer = setTimeout(() => {
                    fetchSuggestions(this.value.trim());
                }, 300);
            });

            input.addEventListener('keydown', function(e) {
                const items = dropdown.querySelectorAll('.autocomplete-item');
                
                if (e.key === 'ArrowDown') {
                    e.preventDefault();
                    if (selectedIndex < items.length - 1) {
                        selectItem(selectedIndex + 1);
                    }
                } else if (e.key === 'ArrowUp') {
                    e.preventDefault();
                    if (selectedIndex > 0) {
                        selectItem(selectedIndex - 1);
                    }
                } else if (e.key === 'Enter' && selectedIndex >= 0) {
                    e.preventDefault();
                    input.value = suggestions[selectedIndex];
                    dropdown.style.display = 'none';
                    this.form.submit();
                } else if (e.key === 'Escape') {
                    dropdown.style.display = 'none';
                    selectedIndex = -1;
                }
            });

            document.addEventListener('click', function(e) {
                if (!input.contains(e.target) && !dropdown.contains(e.target)) {
                    dropdown.style.display = 'none';
                }
            });
        }

        // Initialize autocomplete for the search results form
        document.addEventListener('DOMContentLoaded', function() {
            initAutocomplete('search-input-results', 'autocomplete-dropdown-results');
        });
    </script>
</body>
</html>
`

	searchFormTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bible Search</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            color: #333;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .container {
            max-width: 600px;
            width: 100%;
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
            padding: 50px;
            text-align: center;
        }
        
        .search-icon {
            font-size: 4em;
            margin-bottom: 20px;
            opacity: 0.8;
        }
        
        .title {
            font-size: 2.5em;
            font-weight: 700;
            margin-bottom: 10px;
            background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .subtitle {
            font-size: 1.2em;
            color: #6b7280;
            margin-bottom: 40px;
        }
        
        .search-form {
            margin-bottom: 30px;
        }
        
        .search-input {
            width: 100%;
            padding: 16px 24px;
            font-size: 1.1em;
            border: 2px solid #e5e7eb;
            border-radius: 50px;
            outline: none;
            transition: all 0.3s ease;
            margin-bottom: 20px;
        }
        
        .search-input:focus {
            border-color: #4f46e5;
            box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.1);
        }
        
        .search-button {
            width: 100%;
            padding: 16px 24px;
            font-size: 1.1em;
            background: linear-gradient(135deg, #4f46e5 0%, #7c3aed 100%);
            color: white;
            border: none;
            border-radius: 50px;
            cursor: pointer;
            font-weight: 600;
            transition: all 0.3s ease;
        }
        
        .search-button:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(79, 70, 229, 0.3);
        }
        
        .suggestions {
            margin-bottom: 30px;
        }
        
        .suggestions-title {
            font-size: 1.1em;
            font-weight: 600;
            margin-bottom: 15px;
            color: #374151;
        }
        
        .suggestion-tags {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            justify-content: center;
        }
        
        .suggestion-tag {
            padding: 8px 16px;
            background: rgba(79, 70, 229, 0.1);
            color: #4f46e5;
            border-radius: 20px;
            text-decoration: none;
            font-size: 0.9em;
            transition: all 0.3s ease;
            border: 2px solid rgba(79, 70, 229, 0.2);
        }
        
        .suggestion-tag:hover {
            background: #4f46e5;
            color: white;
            transform: translateY(-1px);
        }
        
        .nav-section {
            display: flex;
            gap: 15px;
            justify-content: center;
            flex-wrap: wrap;
        }
        
        .nav-button {
            padding: 12px 24px;
            background: rgba(107, 114, 128, 0.1);
            color: #6b7280;
            text-decoration: none;
            border-radius: 25px;
            font-weight: 600;
            transition: all 0.3s ease;
            border: 2px solid rgba(107, 114, 128, 0.2);
        }
        
        .nav-button:hover {
            background: #6b7280;
            color: white;
            transform: translateY(-1px);
        }
        
        .autocomplete-dropdown {
            position: absolute;
            top: 100%;
            left: 0;
            right: 0;
            background: white;
            border: 2px solid #e5e7eb;
            border-top: none;
            border-radius: 0 0 15px 15px;
            max-height: 200px;
            overflow-y: auto;
            z-index: 1000;
            display: none;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }
        
        .autocomplete-item {
            padding: 12px 20px;
            cursor: pointer;
            border-bottom: 1px solid #f3f4f6;
            transition: background-color 0.2s ease;
        }
        
        .autocomplete-item:hover,
        .autocomplete-item.selected {
            background-color: #f8fafc;
            color: #4f46e5;
        }
        
        .autocomplete-item:last-child {
            border-bottom: none;
        }
        
        .search-input-container {
            position: relative;
            flex: 1;
            min-width: 200px;
        }
        
        @media (max-width: 768px) {
            .container {
                padding: 30px;
            }
            
            .title {
                font-size: 2em;
            }
            
            .suggestion-tags {
                flex-direction: column;
                align-items: center;
            }
            
            .nav-section {
                flex-direction: column;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="search-icon">🔍</div>
        <h1 class="title">Search the Bible</h1>
        <p class="subtitle">Find verses, stories, and wisdom from God's Word</p>
        
        <form class="search-form" method="GET" action="/bible/search">
            <div class="search-input-container">
                <input 
                    type="text" 
                    name="q" 
                    id="search-input-main"
                    class="search-input" 
                    placeholder="Enter words, phrases, or topics..." 
                    required
                    autofocus
                    autocomplete="off"
                >
                <div id="autocomplete-dropdown-main" class="autocomplete-dropdown"></div>
            </div>
            <button type="submit" class="search-button">🔍 Search Scripture</button>
        </form>
        
        <div class="suggestions">
            <h3 class="suggestions-title">Popular searches:</h3>
            <div class="suggestion-tags">
                <a href="/bible/search?q=love" class="suggestion-tag">love</a>
                <a href="/bible/search?q=faith" class="suggestion-tag">faith</a>
                <a href="/bible/search?q=hope" class="suggestion-tag">hope</a>
                <a href="/bible/search?q=peace" class="suggestion-tag">peace</a>
                <a href="/bible/search?q=joy" class="suggestion-tag">joy</a>
                <a href="/bible/search?q=grace" class="suggestion-tag">grace</a>
                <a href="/bible/search?q=forgiveness" class="suggestion-tag">forgiveness</a>
                <a href="/bible/search?q=wisdom" class="suggestion-tag">wisdom</a>
            </div>
        </div>
        
        <div class="nav-section">
            <a href="/bible/list_books" class="nav-button">📚 Browse Books</a>
            <a href="/bible/random_verse" class="nav-button">🎲 Random Verse</a>
        </div>
    </div>

    <script>
        function initAutocomplete(inputId, dropdownId) {
            const input = document.getElementById(inputId);
            const dropdown = document.getElementById(dropdownId);
            let selectedIndex = -1;
            let suggestions = [];
            let debounceTimer = null;

            function fetchSuggestions(query) {
                if (query.length < 2) {
                    dropdown.style.display = 'none';
                    return;
                }

                fetch('/bible/autocomplete?q=' + encodeURIComponent(query))
                    .then(response => response.json())
                    .then(data => {
                        suggestions = data || [];
                        displaySuggestions();
                    })
                    .catch(error => {
                        console.error('Autocomplete error:', error);
                        dropdown.style.display = 'none';
                    });
            }

            function displaySuggestions() {
                if (suggestions.length === 0) {
                    dropdown.style.display = 'none';
                    return;
                }

                dropdown.innerHTML = '';
                suggestions.forEach((suggestion, index) => {
                    const item = document.createElement('div');
                    item.className = 'autocomplete-item';
                    item.textContent = suggestion;
                    item.addEventListener('click', () => {
                        input.value = suggestion;
                        dropdown.style.display = 'none';
                        input.form.submit();
                    });
                    dropdown.appendChild(item);
                });

                dropdown.style.display = 'block';
                selectedIndex = -1;
            }

            function selectItem(index) {
                const items = dropdown.querySelectorAll('.autocomplete-item');
                items.forEach(item => item.classList.remove('selected'));
                
                if (index >= 0 && index < items.length) {
                    items[index].classList.add('selected');
                    selectedIndex = index;
                }
            }

            input.addEventListener('input', function() {
                clearTimeout(debounceTimer);
                debounceTimer = setTimeout(() => {
                    fetchSuggestions(this.value.trim());
                }, 300);
            });

            input.addEventListener('keydown', function(e) {
                const items = dropdown.querySelectorAll('.autocomplete-item');
                
                if (e.key === 'ArrowDown') {
                    e.preventDefault();
                    if (selectedIndex < items.length - 1) {
                        selectItem(selectedIndex + 1);
                    }
                } else if (e.key === 'ArrowUp') {
                    e.preventDefault();
                    if (selectedIndex > 0) {
                        selectItem(selectedIndex - 1);
                    }
                } else if (e.key === 'Enter' && selectedIndex >= 0) {
                    e.preventDefault();
                    input.value = suggestions[selectedIndex];
                    dropdown.style.display = 'none';
                    this.form.submit();
                } else if (e.key === 'Escape') {
                    dropdown.style.display = 'none';
                    selectedIndex = -1;
                }
            });

            document.addEventListener('click', function(e) {
                if (!input.contains(e.target) && !dropdown.contains(e.target)) {
                    dropdown.style.display = 'none';
                }
            });
        }

        // Initialize autocomplete for the main search form
        document.addEventListener('DOMContentLoaded', function() {
            initAutocomplete('search-input-main', 'autocomplete-dropdown-main');
        });
    </script>
</body>
</html>
`
)
