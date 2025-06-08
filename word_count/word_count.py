#!/usr/bin/env python3
import sqlite3

con = sqlite3.connect("/home/rmintz/kjv.db")
cur = con.cursor()


execute_str = "select text from kjv"
res = cur.execute(execute_str)

results = res.fetchall()

for i in results:
    print("this is i: ", i )
    i[0].split()
    
    # breakpoint()

    
print(results)
