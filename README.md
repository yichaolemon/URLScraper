# URLScraper
Parallel downloading of webpages within the same domain using Go.

## Features
Is fast because of concurrency, but is not very fast because of much needed synchronization for dealing with various
annoying "edge" cases of html. Handles redirects and does a decent job at error logging. 

## Warning
To use this to download resources from an arbitrary domain, you might need to tweak things here and there, but it should not be too bad. Additionally, if something takes a long while to download (program seemingly got stuck), it could be that that thing is super big (like some insanely large .h5 models and datasets), or some other issues (perhaps the webpage not responding or super slow etc.). In that case, just Ctrl-C. 

## Requirement
Docker for running the code, and python for simple https server. Good internet connection.  

## Sample Usage
**Important**: you should download each term separately, for there is simply too much congestion and it will probably crash your
computer before you know it. I might revamp the logic so that it may mitigate these issues, but until then, below is the suggested usage. 


Here is how you can use the code to download Spring 2020 Yale CS senior projects from https://zoo.cs.yale.edu/classes/cs490/19-20b/index.html. You need to be on Yale network (on-campus/vpn) for this to work. 
```
$ mkdir Spring2020 # output_dir 
$ ./run -url "https://zoo.cs.yale.edu/classes/cs490/19-20b/index.html" -dir "Spring2020"
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

