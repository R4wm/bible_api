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
## TODO:
Swipe to next chapter
Move from sqlite3 to elasticsearch
Detailed search analytics
