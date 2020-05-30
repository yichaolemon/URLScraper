# URLScraper
Parallel downloading of webpages within the same domain using Go.

## Features
Is fast because of concurrency, but is not very fast because of much needed synchronization for dealing with various
annoying "edge" cases of html. Handles redirects and does a decent job at error logging. Not robust for general purpose use, 
but perhaps decent for recreational and non-mallicious use. You might have to tweak a few things here and there, but should not 
be too bad either. 

## Requirement
Docker for running the code, and python for simple https server. Good internet connection.  

## Sample Usage
**Important**: you should download each term separately, for there is simply too much congestion and it will probably crash your
computer before you know it. I might revamp the logic so that it may mitigate these issues, but until then, below is the suggested usage. 


Here is how you can use the code to download Spring 2020 Yale CS senior projects from https://zoo.cs.yale.edu/classes/cs490/19-20b/index.html.
```
$ mkdir Spring2020 # output_dir 
$ ./run -url "https://zoo.cs.yale.edu/classes/cs490/19-20b/index.html" -filename "index.html"
```
This will take a while, for there are some obscenely big code zipfiles.
If you only want a specific term's project, feel free to modify the `-url` to accomplish that. 
Do this in the `output_dir` root folder. 
```
$ cd /path/to/Spring2020
$ python -m SimpleHTTPServer 
```
and then go to `http://localhost:8000/classes/cs490/19-20b/` on your browser to browse all these files. 

## Terminal Output
`downloaded file ...` and `downloading file ...`, besides a couple other occasional debug messages that aren't very important. 
Eventually, the number of downloaded files should be equal to the number of downloading files. 

