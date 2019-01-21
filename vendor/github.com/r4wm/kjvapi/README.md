# kjvapi

## Purpose
The idea is to be able to query the bible from a RESTful API using go webserver as the backend

The idea so far is
>nginx -> go web api -> redis 

returning json formated results (super fast)

To get the redis in memory db with kjv text, please see [kjv redis](https://github/r4wm/kjv)

## Testing
```bash
@arch-lt ~/go/src/github.com/r4wm/kjvapi> go test
PASS
ok  	github.com/r4wm/kjvapi	0.001s
@arch-lt ~/go/src/github.com/r4wm/kjvapi> 
```

## Example (future usage)
pretending for now the request fields came via http request, the verses would be fetched from redis db

source
```golang
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/r4wm/kjvapi"
)

func main() {
	var verses []kjvapi.KJVVerse
	//Pretend the data came from http.Request
	book := "Genesis"
	chapter := 1
    
    //Assume verses came from redis lookup
	verses = append(verses, kjvapi.KJVVerse{Verse: 1,
		Text: "In the beginning God created the heaven and the earth."})

	verses = append(verses, kjvapi.KJVVerse{Verse: 2,
		Text: "And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters."})

	result := kjvapi.GetChapter(book, chapter, verses)
	jsonQuery, err := json.MarshalIndent(result, "", "  ")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", jsonQuery)
}
```
result

```bash
@arch-lt ~/go/src/github.com/r4wm> go run main.go
{
  "book": "Genesis",
  "Chapters": [
    {
      "chapter": 1,
      "Verses": [
        {
          "verse": 1,
          "text": "In the beginning God created the heaven and the earth."
        },
        {
          "verse": 2,
          "text": "And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters."
        }
      ]
    }
  ]
}
@arch-lt ~/go/src/github.com/r4wm>
```

## Create your own KJV database
```go
package main

import (
	"fmt"

	"github.com/r4wm/kjvapi"
)

func main() {
	fmt.Println("Starting kjv database generation")
	kjvapi.CreateKJVDB("/tmp/kjv.db")
}
```
