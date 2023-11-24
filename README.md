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

### Docker
Try this: 
```bash
[2023-11-24 17:10:36] rmintz@node-1:~/github/bible_api 
$ docker build -t bible_api . 
[+] Building 178.2s (11/11) FINISHED                                                                                                                                          
 => [internal] load .dockerignore                                                                                                                                        0.0s
 => => transferring context: 2B                                                                                                                                          0.0s
 => [internal] load build definition from Dockerfile                                                                                                                     0.0s
 => => transferring dockerfile: 899B                                                                                                                                     0.0s
 => [internal] load metadata for docker.io/library/golang:latest                                                                                                         0.3s
 => [1/7] FROM docker.io/library/golang:latest@sha256:9baee0edab4139ae9b108fffabb8e2e98a67f0b259fd25283c2a084bd74fea0d                                                   0.0s
 => [internal] load build context                                                                                                                                        0.0s
 => => transferring context: 10.58kB                                                                                                                                     0.0s
 => CACHED [2/7] RUN mkdir -p /go/src/bible_api                                                                                                                          0.0s
 => CACHED [3/7] WORKDIR /go/src/bible_api                                                                                                                               0.0s
 => [4/7] COPY . /go/src/bible_api                                                                                                                                       0.1s
 => [5/7] RUN go build -o /bible_api ./cmd/bible_api.go                                                                                                                 33.0s
 => [6/7] RUN /bible_api -createDB                                                                                                                                     144.0s 
 => exporting to image                                                                                                                                                   0.6s 
 => => exporting layers                                                                                                                                                  0.6s 
 => => writing image sha256:60fd50afd48ddbe859af058c985ffff598ab834c4b4d6d8239f584d0d6e77cc0                                                                             0.0s 
 => => naming to docker.io/library/bible_api                                                                                                                             0.0s 
[2023-11-24 17:13:35] rmintz@node-1:~/github/bible_api                                                                                                                        
$ docker run -it -d -p8000:8000  bible_api                                                                                                             
06d5aadd5271b515bf8f6de7ac27040fa4a3b4e46cb70b054b2976d3a5868118
[2023-11-24 17:21:11] rmintz@node-1:~/github/bible_api 
$ curl -s localhost:8000/bible/rom/1?json=true| jq . 
{
  "BookName": "ROMANS",
  "Chapter": 1,
  "Verses": [
    "Paul, a servant of Jesus Christ, called [to be] an apostle, separated unto the gospel of God,",
    "(Which he had promised afore by his prophets in the holy scriptures,)",
    "Concerning his Son Jesus Christ our Lord, which was made of the seed of David according to the flesh;",
    "And declared [to be] the Son of God with power, according to the spirit of holiness, by the resurrection from the dead:",
    "By whom we have received grace and apostleship, for obedience to the faith among all nations, for his name:",
    "Among whom are ye also the called of Jesus Christ:",
    "To all that be in Rome, beloved of God, called [to be] saints: Grace to you and peace from God our Father, and the Lord Jesus Christ.",
    "First, I thank my God through Jesus Christ for you all, that your faith is spoken of throughout the whole world.",
    "For God is my witness, whom I serve with my spirit in the gospel of his Son, that without ceasing I make mention of you always in my prayers;",
    "Making request, if by any means now at length I might have a prosperous journey by the will of God to come unto you.",
    "For I long to see you, that I may impart unto you some spiritual gift, to the end ye may be established;",
    "That is, that I may be comforted together with you by the mutual faith both of you and me.",
    "Now I would not have you ignorant, brethren, that oftentimes I purposed to come unto you, (but was let hitherto,) that I might have some fruit among you also, even as among other Gentiles.",
    "I am debtor both to the Greeks, and to the Barbarians; both to the wise, and to the unwise.",
    "So, as much as in me is, I am ready to preach the gospel to you that are at Rome also.",
    "For I am not ashamed of the gospel of Christ: for it is the power of God unto salvation to every one that believeth; to the Jew first, and also to the Greek.",
    "For therein is the righteousness of God revealed from faith to faith: as it is written, The just shall live by faith.",
    "For the wrath of God is revealed from heaven against all ungodliness and unrighteousness of men, who hold the truth in unrighteousness;",
    "Because that which may be known of God is manifest in them; for God hath shewed [it] unto them.",
    "For the invisible things of him from the creation of the world are clearly seen, being understood by the things that are made, [even] his eternal power and Godhead; so that they are without excuse:",
    "Because that, when they knew God, they glorified [him] not as God, neither were thankful; but became vain in their imaginations, and their foolish heart was darkened.",
    "Professing themselves to be wise, they became fools,",
    "And changed the glory of the uncorruptible God into an image made like to corruptible man, and to birds, and fourfooted beasts, and creeping things.",
    "Wherefore God also gave them up to uncleanness through the lusts of their own hearts, to dishonour their own bodies between themselves:",
    "Who changed the truth of God into a lie, and worshipped and served the creature more than the Creator, who is blessed for ever. Amen.",
    "For this cause God gave them up unto vile affections: for even their women did change the natural use into that which is against nature:",
    "And likewise also the men, leaving the natural use of the woman, burned in their lust one toward another; men with men working that which is unseemly, and receiving in themselves that recompence of their error which was meet.",
    "And even as they did not like to retain God in [their] knowledge, God gave them over to a reprobate mind, to do those things which are not convenient;",
    "Being filled with all unrighteousness, fornication, wickedness, covetousness, maliciousness; full of envy, murder, debate, deceit, malignity; whisperers,",
    "Backbiters, haters of God, despiteful, proud, boasters, inventors of evil things, disobedient to parents,",
    "Without understanding, covenantbreakers, without natural affection, implacable, unmerciful:",
    "Who knowing the judgment of God, that they which commit such things are worthy of death, not only do the same, but have pleasure in them that do them."
  ],
  "Color": "SpringGreen",
  "NextChapterLink": "2?json=false",
  "PreviousChapterLink": "",
  "ListAllBooksLink": "../list_books?json=false"
}
[2023-11-24 17:21:36] rmintz@node-1:~/github/bible_api 
$ 

```

