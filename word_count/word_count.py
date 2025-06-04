#!/usr/bin/env python3


import sqlite3
mapper = {"word":[{"book_name": "count"},
                  {"book_chapter": "count"},
                  {"book_chapter_verse": "count"},
                  {"total_count": "count"},
                  ]
          }

con = sqlite3.connect("../data/kjv.db")
cur = con.cursor()
res = cur.execute("select * from kjv limit 1")
print(res.fetchone())
