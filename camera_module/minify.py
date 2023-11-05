#!/usr/bin/python3

import sys
import htmlmin

if len(sys.argv) <= 2:
    print("Usage: minify.py input.html output.html")
    sys.exit(1)

input_file = open(sys.argv[1], "r")
output_file = open(sys.argv[2], "w")

html = input_file.read()
input_file.close()

html_minified = htmlmin.minify(html, remove_empty_space=True, remove_comments=True)
output_file.write(html_minified)

output_file.close()
print("Done.")
