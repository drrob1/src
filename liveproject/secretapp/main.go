package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"src/liveproject/secretapp/filestore"
	"src/liveproject/secretapp/handlers"
	//"github.com/amitsaha/manning-go-advanced-beginners/project-series-2/mini-project-1/milestone1-code/filestore"
	//"github.com/amitsaha/manning-go-advanced-beginners/project-series-2/mini-project-1/milestone1-code/handlers"
)

/*
  From manning Web application w/ 2 API endpoints to create and view a secret.

  The secrets are stored in a file.  When a secret is viewed, it is destroyed and cannot be viewed again.  The user requests will be a JSON body, and the returned response will also be a JSON body.

  Task here is to implement a way for the user to specify the file path to store the secrets.  If this is not found, exit.
  On startup, the server should create the file if it doesn't exist.

  Creating the http handlers
  Create a new serveMux object and register handler functions for 2 paths, /healthcheck using healthCheckHandler, and /, using secretHandler.
  The healthCheck handler should verify that our application is ready for requests.  If so, it should return "ok" as a response.

  The secretHandler should process both the GET and POST requests.

  When a post request is made, secretHandler will look for the JSON body, {"plaintext" : "my$super$secret"}.  If the JSON body is as expected, secretHandler will store the secret
  in a map, using for the key the MD5 hash of the secret text.  The MD5 hash is returned as a JSON response, {"id" : "md5 hash value"}.  The value of the id object is then used to retrieve
  the secret using a GET request.  The MD5 hash is not to be used as an encryption method, but merely as a map key.  Encryption will be covered in milestone 3.

  After updating the map, the map file must be written as atomic using a sync mutex.

  When a GET request is made to /, the handler must check if an id has been specified as a request path, ie, /id.  If not, it should return an error such as HTTP 400 indicating a bad request.
  If an id has been specified, then the file contents are read and the map is updated w/ the contents.

  The map is checked to see if there is an item w/ the key specified by id.  If found, the value is retrieved, the object is deleted from the map, and the map is written back to the file.
  The retrieved value is sent back as a JSON response, {"data" : "original$secret"}.

  If an object w/ key id is not found, an HTTP 404 error is returned.
  In both cases, a JSON response {"data" : ""} is returned.

  Secretapp is the name of the application, and it listens on port :8080.

  Start the application in one terminal
    DATA_FILE_PATH=./data/json  ./secretapp

  In another terminal, you should be able to create a secret w/ curl:
    curl -X POST http://localhost:8080 -d '{"plain_text":"my super secret"}' {"id": the MD5 hash of my super secret}
  $ curl -X POST http://localhost:8080 -d '{"plain_text":"My super secret123"}' {"id":"c616584ac64a93aafe1c16b6620f5bcd"}

  Now terminate the server w/ ctrl-C in its terminal.  And start the server again.

  To view the secret, copy the id from the response above:
  $ curl http://localhost:8080/c616584ac64a93aafe1c16b6620f5bcd
  Response should be: {"data":"My super secret123"}

  If you try to retrieve it again, you will not get any data:
  $ curl http://localhost:8080/c616584ac64a93aafe1c16b6620f5bcd
  Response should now be: {"data":""}


ServeMux is an HTTP request multiplexer. It matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most
closely matches the URL.

Patterns name fixed, rooted paths, like "/favicon.ico", or rooted subtrees, like "/images/" (note the trailing slash). Longer patterns take precedence over shorter ones,
so that if there are handlers registered for both "/images/" and "/images/thumbnails/", the latter handler will be called for paths beginning with "/images/thumbnails/"
and the former will receive requests for any other paths in the "/images/" subtree.

Note that since a pattern ending in a slash names a rooted subtree, the pattern "/" matches all paths not matched by other registered patterns, not just the URL with Path == "/".

If a subtree has been registered and a request is received naming the subtree root without its trailing slash, ServeMux redirects that request to the subtree root
(adding the trailing slash). This behavior can be overridden with a separate registration for the path without the trailing slash. For example, registering "/images/"
causes ServeMux to redirect a request for "/images" to "/images/", unless "/images" has been registered separately.

Patterns may optionally begin with a host name, restricting matches to URLs on that host only. Host-specific patterns take precedence over general patterns,
so that a handler might register for the two patterns "/codesearch" and "codesearch.google.com/" without also taking over requests for "http://www.google.com/".

ServeMux also takes care of sanitizing the URL request path and the Host header, stripping the port number and redirecting any request containing . or .. elements or
repeated slashes to an equivalent, cleaner URL.
*/

func main() {

	listenAddr := os.Getenv("LISTEN_ADDR")
	if len(listenAddr) == 0 {
		listenAddr = ":8080"
	}

	mux := http.NewServeMux()
	handlers.SetupHandlers(mux)

	dataFilePath := os.Getenv("DATA_FILE_PATH")
	if len(dataFilePath) == 0 {
		log.Fatal("Specify DATA_FILE_PATH")
	}
	fmt.Printf(" Using a listen address of %s and data file path of %s\n", listenAddr, dataFilePath)

	filestore.Init(dataFilePath)

	err := http.ListenAndServe(listenAddr, mux)
	if err != nil {
		log.Fatalf("Server could not start listening on %s. Error: %v", listenAddr, err)
	}
}
