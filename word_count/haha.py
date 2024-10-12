#!/usr/bin/env python3

import json


mapper = {"{book_name}": "{book_name_count}"}
something = {}
list_of_words = []
something['the'] = [{"GEN" : {"count": 100},
                    "GEN 1": {"count": 10},
                     "GEN 1:1": {"count": 1},},
                    ]
exodus = {"EXO": {"count": 200},
          "EXO 1": {"count": 20},
          "EXO 1:1": {"count": 2},}
total = {"total": {"count": 2225}}

something['the'].append(exodus)
something['the'].append(total)

list_of_words.append(something)

result = json.dumps(list_of_words, indent=8)

print(result)

