package main

import (
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

  Secret-app is the name of the application, and it listens on port :8080.
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

	filestore.Init(dataFilePath)

	err := http.ListenAndServe(listenAddr, mux)
	if err != nil {
		log.Fatalf("Server could not start listening on %s. Error: %v", listenAddr, err)
	}
}
