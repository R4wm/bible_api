# Bible API
- A raw high performance RESTful api written in golang
- King James Version Pure Cambridge Text
- No adds, No distractions, not ever.
- Easy navigation
	- [Simple book listing and buttons choice](https://mintz5.duckdns.org/bible/list_books)
	- [Random Verse Generator](https://mintz5.duckdns.org/bible/random_verse)
	- [All pages support json output](https://mintz5.duckdns.org/bible/random_verse?json=true)
		- provide argument: `?json=true`
	- Forward chapter button (if applicable)
	- Previous chapter button (if applicable)
	- Books link button in Chapter selection
	- [Supports verse ranges](https://mintz5.duckdns.org/bible/EPHESIANS/2/8-9)
	- Search feature
		- Example: `https://mintz5.duckdns.org/bible/search?q=heart`
        - Yearly Reading schedule for Old Testament and New Testament for every day
        - Monthly Reading schedule for Proverbs and Psalsm for every day


To use public version of running API, visit the [bible_api](https://mintz5.duckdns.org/bible/list_books)
## Docker
```bash

@fake:[~/github/bible_api]:$ docker build -t something:000 -f ./Dockerfile .

0 @fake:[~/github/bible_api]:$ docker image ls  | tail -2 
ubuntu                                                             20.04          20fffa419e3a   5 weeks ago          72.8MB

0 @fake:[~/github/bible_api]:$ docker run -it -d -p 8000:8000 --name something_c bf306415b1f5
303d470aac6f9c375e29d1c4073ab633a88752bef56ab02fd0fe8e025bb30cd4
0 @fake:[~/github/bible_api]:$ 
0 @fake:[~/github/bible_api]:$ curl http://localhost:8000/bible/list_books | tail -5 
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  8080    0  8080    0     0  3945k      0 --:--:-- --:--:-- --:--:-- 3945k
      
      <p><button class="block" onclick="window.location.href= 'REVELATION?json=false';" >REVELATION</button></p>
      
   </body>
</html>
0 @fake:[~/github/bible_api]:$ 

```
## TODO:
Swipe to next chapter
Move from sqlite3 to elasticsearch
Detailed search analytics
