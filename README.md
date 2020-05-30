# URLScraper
Parallel downloading of webpages within the same domain using Go. 

## Requirement
Docker for running the code, and python for simple https server. 

## Sample Usage
Here is how you can use the code to download **all** the Yale CS senior projects from https://zoo.cs.yale.edu/classes/cs490/.
```
$ ./run -url "https://zoo.cs.yale.edu/classes/cs490/" -filename "index.html"
```
This will take a while, for there are some obscenely big code zipfiles. \\
If you only want a specific term's project, feel free to modify the `-url` to accomplish that. 
```
$ python -m SimpleHTTPServer 
```
and then go to `http://localhost:8000/classes/cs490/` on your browser to browse all these files. 

