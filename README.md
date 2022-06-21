# FileShare
This is simple distributed P2P file sharing network allowing nodes to connect, register and share files.   

## Run Project in Local
```
cd src/main
go run main.go
```
This starts a HTTP server running on port: 5000

## Contents 
* `src/`: Contains the code for FileShare, a peer-to-peer distributed file sharing system.
    * `fileshare/`: Contains the Peer and SwarmMaster code, all part of the `fileshare` package.  
    * `main/`: Contains `main.go` and Peer directories, used for running a test case of the FileShare system. 

## How to use
After running ```go run main.go``` a localhost http server is start that runs on port ```5000```. Below are available endpoints

## API Documentation

#### Create new peer:
```
curl --location --request POST 'localhost:5001/createPeer' \
--header 'Content-Type: application/json' \
--data-raw '{
    "id": 12
}'
```

#### Get nodes:
```
curl --location --request GET 'localhost:5001/getActivePeers'
```
#### Upload file:
```
curl --location --request POST 'localhost:5001/upload' \
--form 'file=@"/Users/vipulpanchal/Desktop/demo.txt"'
```

#### Search File
```
curl --location --request GET 'localhost:5001/searchFile/demo.txt?ownername=StorageTeam'
```