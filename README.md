# Zip-Starter
### A Starter-Server for harvesting

zip-starter is a single-route app that only accepts JSON POST requests. It's job is to create a "Starter" zip archive for adding data to, properlyconfigured according to the guide:
https://github.com/datarefugephilly/workflow/tree/master/harvesting-toolkit#3-generate-html-json--directory

The archive derives it's name from the path posted to, or the "uuid" paremeter of the json document, favoring the json entry.

if a "url" key is specified in the json document, the server will fetch the document in question and write it
to [name].html

### Getting Started
* make sure you have go installed.
* in a terminal navigate to this directory & run `go build`, a binary called zip-starter will be created in this folder
* run `./zip-starter` from the terminal, a server will start
* in another terminal try using `curl` to post to the server: ```curl -d "{ \"uuid\" : \"test-zip-starter\", \"url\" : \"http://www.epa.gov\"}" http://localhost:3000 > test-zip-starter.zip```