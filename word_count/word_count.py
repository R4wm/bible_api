#!/usr/bin/env python3
<<<<<<< HEAD


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
=======
import sqlite3

con = sqlite3.connect("/home/rmintz/kjv.db")
cur = con.cursor()


execute_str = "select text from kjv"
res = cur.execute(execute_str)

results = res.fetchall()

for i in results:
    print("this is i: ", i )
    i[0].split()
    
    breakpoint()

    
print(results)
>>>>>>> 55b93a2 (clean up and add Makefile)
