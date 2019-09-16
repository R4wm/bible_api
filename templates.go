package kjv

const (
	lastCardinalVerseNum = 31101

	verseTemplate = `
<!DOCTYPE html>
<html>
   <body style="background-color:{{ .Color }};">
      <h1>
	 <center>{{ .Verse.Book }} {{ .Verse.Chapter }}:{{ .Verse.Verse }} </center>
      </h1>
      <h3>
	 <center>{{ .Verse.Text }}</center>
	 <center><p><button class="block" onclick="window.location.href={{.ChapterRef}};">{{ .Verse.Book }} {{ .Verse.Chapter }}</button></p\
></center>
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
    <h1><center>{{ .BookName }} {{ .Chapter }}</h1>
  <body>
    {{ range $index, $results := .Verses }}
    <p><b><left><a href={{ verseLink $index }}> {{ add $index 1}}</a> {{ . }} </b></p>
    {{ end }}
    <div class="btn-group">
    {{ if .PreviousChapterLink  }}
    <p><button class="block" onclick="window.location.href={{.PreviousChapterLink}}">Previous Chapter</button></p>
    {{ end }}
    <p><button class="block" onclick="window.location.href={{.ListAllBooksLink}}">Books</button></p>
    {{ if .NextChapterLink  }}
    <p><button class="block" onclick="window.location.href={{.NextChapterLink}}">Next Chapter</button></p>
    {{ end }}
    </div>
  </body>
</html>
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
      </style>
      <title>Books of the Bible</title>
   </head>
   <body style="background-color:{{ .Color }};">
      {{ range $key, $value := .Books }}
      <p><button class="block" onclick="window.location.href={{ createLink $value }}">{{ $value }}</button></p>
      {{ end }}
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
       <p><button class="block" onclick="window.location.href={{ $results }}">{{ add $index 1 }}</button></p>
     {{ end }}
   </body>
</html>
`
)
