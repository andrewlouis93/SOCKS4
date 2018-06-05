# SOCKS4
🧦 🧦 🧦 🧦

A very simple SOCKS4 server :)

### Scope
* Set up the listening socket and dump the bytes received from the SOCKS client when it connects
* Establish the connection to the target
* Once the connection is established to the target, how do you know when data is available to send back to the client?
* Likewise, how do you know when the client wants to send data (hint: non-blocking sockets)
* How do you ensure the entire proxy is non-blocking (you're not waiting for the client to send something before receiving from the server)?
* Can you add support for multiple clients?

### TODO
* Add better error handling
* Bubbling up errors to the main goroutine
* Connection logging
* Port arguments
* Refactor SOCKS server into separate package?

### Example Usage

Making a client request

```
curl -v --socks4 socks4://127.0.0.1:8080 https://www.google.com
````
