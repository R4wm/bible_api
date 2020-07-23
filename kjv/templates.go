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
	 }
	 .block:hover {
	 background-color: #ddd;
	 color: black;
	 }
      </style>
      <title>Books of the Bible</title>
   </head>
   <body style="background-color:{{ .Color }};">
      {{ range $key, $value := .Books }}
      <p><button class="block" onclick="window.location.href= '{{ createLink $value }}';" >{{ $value }}</button></p>
      {{ end }}
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
   </body>
</html>
`
	chapterTemplate = `
<html>
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
<title>{{.BookName}} {{.Chapter}}:{{.SingleVerse}}</title>
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
    {{if .StartVerse}}<h1><center><a href={{.BookName}}>{{ .BookName }} {{ .Chapter }}</a>:{{.StartVerse}}-{{.EndVerse}}</h1>
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
     {{ range $index, $results := .Links }}
       <p><button class="block" onclick="window.location.href = '{{ $results }}'">{{ add $index 1 }}</button></p>
     {{ end }}
   </body>
</html>
`

	searchResultTemplate = `
<!DOCTYPE html>
<html>

      {{range .Verses }}
	 <center>
	   <p> <a href="{{ createLink .}}">{{ .Book }} {{ .Chapter }}:{{ .Verse}} </a></p>
	   <p> {{ .Text }} </p>
	 </center>
      {{ end }}

   </body>
</html>
`
)
