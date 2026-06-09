To work with the api you first need to have your docker postgres running with the data in it.
Once it is running, you can just launch the container and you should see the log
Listening on port 8080

Once it is done you can work with it.

The api is composed with 3 file : 
    - api/main.go
    - api/handler.go
    - db/query.go

The model is used from internal/model/model.go