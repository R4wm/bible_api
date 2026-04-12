package kjv

const (
	searchBarCSS = `
      #tfheader{
      background-color:#c3dfef;
      }
      #tfnewsearch{
      float:right;
      padding:20px;
      }
      .tftextinput{
      margin: 0;
      padding: 5px 15px;
      font-family: Arial, Helvetica, sans-serif;
      font-size:14px;
      border:1px solid #0076a3; border-right:0px;
      border-top-left-radius: 5px 5px;
      border-bottom-left-radius: 5px 5px;
      }
      .tfbutton {
      margin: 0;
      padding: 5px 15px;
      font-family: Arial, Helvetica, sans-serif;
      font-size:14px;
      outline: none;
      cursor: pointer;
      text-align: center;
      text-decoration: none;
      color: #ffffff;
      border: solid 1px #0076a3; border-right:0px;
      background: #0095cd;
      background: -webkit-gradient(linear, left top, left bottom, from(#00adee), to(#0078a5));
      background: -moz-linear-gradient(top,  #00adee,  #0078a5);
      border-top-right-radius: 5px 5px;
      border-bottom-right-radius: 5px 5px;
      }
      .tfbutton:hover {
      text-decoration: none;
      background: #007ead;
      background: -webkit-gradient(linear, left top, left bottom, from(#0095cc), to(#00678e));
      background: -moz-linear-gradient(top,  #0095cc,  #00678e);
      }
      .tfbutton::-moz-focus-inner {
      border: 0;
      }
      .tfclear{
      clear:both;
      }
`

	searchBarHTML = `
    <div id="tfheader">
      <form id="tfnewsearch" method="get" action="/bible/search">
        <input type="text" class="tftextinput" name="q" size="21" maxlength="120"
               list="search-suggestions" autocomplete="off"><datalist id="search-suggestions"></datalist><input type="submit" value="search" class="tfbutton">
      </form>
      <div class="tfclear"></div>
    </div>
    <script>
    (function() {
      var input = document.querySelector('.tftextinput');
      var datalist = document.getElementById('search-suggestions');
      var timer;
      input.addEventListener('input', function() {
        clearTimeout(timer);
        var val = input.value.trim();
        if (val.length < 2) { datalist.innerHTML = ''; return; }
        timer = setTimeout(function() {
          fetch('/bible/suggest?q=' + encodeURIComponent(val))
            .then(function(r) { return r.json(); })
            .then(function(words) {
              datalist.innerHTML = '';
              words.forEach(function(w) {
                var opt = document.createElement('option');
                opt.value = w;
                datalist.appendChild(opt);
              });
            })
            .catch(function() { datalist.innerHTML = ''; });
        }, 250);
      });
    })();
    </script>
`

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
	 }
	 .block:hover {
	 background-color: #ddd;
	 color: black;
	 }
` + searchBarCSS + `
      </style>
      <title>Books of the Bible</title>
   </head>
   <body style="background-color:{{ .Color }};">
` + searchBarHTML + `
      {{ range $key, $value := .Books }}
      <p><button class="block" onclick="window.location.href= '{{ createLink $value }}';" >{{ $value }}</button></p>
      {{ end }}
   </body>
</html>
`
	verseTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
` + searchBarCSS + `
    </style>
  </head>
  <body style="background-color:{{ .Color }};">
` + searchBarHTML + `
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
<!DOCTYPE html>
<html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
  <title>{{.BookName}} {{ .Chapter}}</title>
  <style>
.btn-group button {
  background-color: gold;
  border: 1px solid green;
  color: black;
  padding: 10px 24px;
  cursor: pointer;
  float: center;
}
.btn-group:after {
  content: "";
  clear: both;
  display: table;
}
.btn-group button:not(:last-child) {
  border-right: none;
}
.btn-group button:hover {
  background-color: #3e8e41;
}
` + searchBarCSS + `
  </style>
</head>
  <body style="background-color:{{ .Color }};">
` + searchBarHTML + `
    <h1><center><a href=../{{.BookName}}>{{ .BookName }}</a> {{ .Chapter }}</h1>
    {{ range $index, $results := .Verses }}
    <p><b><left><a href={{ verseLink $index }}> {{ add $index 1}}</a> {{ . }} </b></p>
    {{ end }}
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
<!DOCTYPE html>
<html>
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
  <title>{{.HTMLTitle}}</title>
  <style>
    .btn-group button {
    background-color: gold;
    border: 1px solid green;
    color: black;
    padding: 10px 24px;
    cursor: pointer;
    float: center;
    }
    .btn-group:after {
    content: "";
    clear: both;
    display: table;
    }
    .btn-group button:not(:last-child) {
    border-right: none;
    }
    .btn-group button:hover {
    background-color: #3e8e41;
    }
` + searchBarCSS + `
  </style>
</head>
  <body style="background-color:{{ .Color }};">
` + searchBarHTML + `
    {{if .StartVerse}}<h1><center><a href=../{{.Chapter}}>{{ .BookName }} {{ .Chapter }}</a>:{{.StartVerse}}-{{.EndVerse}}</h1>
    {{else}}
    <h1><center><a href="../{{.Chapter}}">{{ .BookName }} {{ .Chapter }}</a>:{{.SingleVerse}}
    {{end}}
	  {{ range $index, $results := .Verses }}
	  {{ range $verseNum, $verseText := $results}}
	  <p><b><left><a href={{$verseNum}}?json=false> {{$verseNum}}</a> {{$verseText }} </b></p>
	  {{end}}
	  {{end}}
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
` + searchBarCSS + `
    </style>
    <title>{{ .Name }}</title>
  </head>
  <body style="background-color:{{ .Color }};">
` + searchBarHTML + `
    <p><center><h1> {{ .Name }} </h1><center></p>
    <button onclick="window.location.href='../bible/list_books';" class="w3-bar-item w3-button" style="width:33.3%">Books Menu</button>
    {{ range $index, $results := .Links }}
    <p><button class="block" onclick="window.location.href = '{{ $results }}'">{{ add $index 1 }}</button></p>
    {{ end }}
  </body>
</html>
`

	searchResultTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <title>Bible Search: "{{.SearchString}}"</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="ROBOTS" content="NOINDEX, NOFOLLOW" />
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.7/dist/chart.umd.min.js"></script>
    <style type="text/css">
` + searchBarCSS + `
      body { font-family: Arial, Helvetica, sans-serif; margin: 0; }
      .chart-container {
        position: relative;
        width: 95%;
        max-width: 1200px;
        margin: 20px auto;
        background: #fff;
        border: 1px solid #ddd;
        border-radius: 4px;
        padding: 16px;
      }
      .results-header {
        text-align: center;
        margin: 16px 0;
        font-size: 18px;
      }
      .results-header span {
        font-weight: bold;
        color: #0076a3;
      }
      .result-item {
        padding: 8px 16px;
        border-bottom: 1px solid #eee;
      }
      .result-item:hover {
        background: #f5f5f5;
      }
      .result-ref {
        font-weight: bold;
      }
      .result-ref a {
        color: #0076a3;
        text-decoration: none;
      }
      .result-ref a:hover {
        text-decoration: underline;
      }
      .result-text {
        color: #333;
        margin-top: 2px;
      }
      .results-list {
        max-width: 1200px;
        margin: 0 auto 40px;
      }
    </style>
  </head>
  <body>
` + searchBarHTML + `

    <div class="results-header">
      Results for <span>"{{.SearchString}}"</span>
    </div>

    <div class="chart-container">
      <canvas id="myChart"></canvas>
    </div>

    <div class="results-list">
    {{range .Verses }}
      <div class="result-item">
        <div class="result-ref"><a href="{{ createLink .}}?json=false">{{ .Book }} {{ .Chapter }}:{{ .Verse}}</a></div>
        <div class="result-text">{{ .Text }}</div>
      </div>
    {{ end }}
    </div>

    <script>
var data = {{.GraphCount}};
var labels = [
  "Genesis","Exodus","Leviticus","Numbers","Deuteronomy",
  "Joshua","Judges","Ruth","1 Samuel","2 Samuel",
  "1 Kings","2 Kings","1 Chronicles","2 Chronicles",
  "Ezra","Nehemiah","Esther","Job","Psalms","Proverbs",
  "Ecclesiastes","Song of Solomon","Isaiah","Jeremiah","Lamentations",
  "Ezekiel","Daniel","Hosea","Joel","Amos",
  "Obadiah","Jonah","Micah","Nahum","Habakkuk",
  "Zephaniah","Haggai","Zechariah","Malachi",
  "Matthew","Mark","Luke","John","Acts",
  "Romans","1 Corinthians","2 Corinthians","Galatians","Ephesians",
  "Philippians","Colossians","1 Thessalonians","2 Thessalonians",
  "1 Timothy","2 Timothy","Titus","Philemon","Hebrews",
  "James","1 Peter","2 Peter","1 John","2 John","3 John",
  "Jude","Revelation"
];

// Color by testament: OT = warm coral, NT = teal
var colors = data.map(function(val, i) {
  if (val === 0) return 'rgba(200,200,200,0.3)';
  return i < 39 ? 'rgba(220, 88, 42, 0.7)' : 'rgba(34, 139, 134, 0.7)';
});
var borders = data.map(function(val, i) {
  if (val === 0) return 'rgba(200,200,200,0.5)';
  return i < 39 ? 'rgba(180, 60, 20, 1)' : 'rgba(20, 100, 100, 1)';
});

new Chart(document.getElementById('myChart'), {
  type: 'bar',
  data: {
    labels: labels,
    datasets: [{
      label: '"{{.SearchString}}" occurrences by book',
      data: data,
      backgroundColor: colors,
      borderColor: borders,
      borderWidth: 1,
      borderRadius: 2
    }]
  },
  options: {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          title: function(items) { return items[0].label; },
          label: function(item) {
            var v = item.raw;
            return v + ' match' + (v !== 1 ? 'es' : '');
          }
        }
      }
    },
    scales: {
      x: {
        ticks: {
          maxRotation: 90,
          minRotation: 45,
          font: { size: 10 }
        },
        grid: { display: false }
      },
      y: {
        beginAtZero: true,
        ticks: {
          precision: 0
        },
        title: {
          display: true,
          text: 'Matches'
        }
      }
    }
  }
});

// Make chart container tall enough to read labels
document.querySelector('.chart-container').style.height =
  Math.max(350, window.innerWidth < 768 ? 300 : 450) + 'px';
</script>
  </body>
</html>
`
)
