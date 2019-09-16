package templates

const (
	VerseTemplate = `
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
)
